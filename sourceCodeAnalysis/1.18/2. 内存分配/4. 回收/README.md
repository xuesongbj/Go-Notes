# 回收

内存回收通常由垃圾回收器引发，内存分配器具体执行。

## 缓存

垃圾回收将所有 `P.cache` 持有的 `mspan` 交还给 `central`，以便闲置内存可以调度给其他 `P.cache` 使用。

```go
// mgc.go

func gcMarkTermination(nextTriggerRatio float64) {
    // 并发状态下。
    startTheWorldWithSema(true)
    
    // 由所在 M/P schedule 执行，因此安全。
    forEachP(func(_p_ *p) {
        _p_.mcache.prepareForSweep()
    })
}
```

```go
// proc.go

func forEachP(fn func(*p)) {

    sched.safePointFn = fn

    // Ask all Ps to run the safe point function.
    for _, p := range allp {
        atomic.Store(&p.runSafePointFn, 1)
    }
}

func schedule() {
    if pp.runSafePointFn != 0 {
        runSafePointFn()
    }    
}

func runSafePointFn() {
    sched.safePointFn(p)    
}
```

由自己M/P执行，因此不会和分配操作起冲突。

```go
// mcache.go

func (c *mcache) prepareForSweep() {
    c.releaseAll()
}

func (c *mcache) releaseAll() {
    
    // 遍历 mcache.alloc 数组。
    for i := range c.alloc {
        s := c.alloc[i]

        if s != &emptymspan {
            // 交还给 central，并用占位替代。
            mheap_.central[i].mcentral.uncacheSpan(s)
            c.alloc[i] = &emptymspan
        }
    }
    
    // 释放 tiny 所持有的内存块。
    c.tiny = 0
    c.tinyoffset = 0
}
```

当mspan从cache交还给central时，其内存可能尚有剩余。如此，其他P.cache获取该mspan时，就能使充分利用剩余内存。一个mspan可被多个cache使用，但同一时刻仅有一个使用者，不存在竞争。

至于`emptymspan`，仅用于占位。

```go
// dummy mspan that contains no free objects.
var emptymspan mspan
```

收归central时，根据是否有剩余空间来决定放到哪个列表。

```go
// mcentral.go

// Return span from an mcache.
func (c *mcentral) uncacheSpan(s *mspan) {
    if stale {
        // 清理结束，会放到合适列表。
        ss := sweepLocked{s}
        ss.sweep(false)
    } else {
        if int(s.nelems)-int(s.allocCount) > 0 {
            c.partialSwept(sg).push(s)
        } else {
            c.fullSwept(sg).push(s)
        }
    }
}
```

&nbsp;

## 清理

垃圾标记完成后，以span为单位进行清理(sweep)操作。

> 所谓清理，并非挨个处理所有object。</br>
> 实际上，内存块除分配位图(allocBits)外，还有个一摸一样的垃圾标记位图(gcmarkBits)。</br>
> 垃圾回收器以此标记出存活对象。其余的是可回收的，是可重新使用的内存位置。</br>
>
> </br>
> 只需将markBits数据"复制"给allocBits，标记存活对象，就实现了清理操作。</br>
> 至于内存单元里的遗留数据清理与否，则是分配操作要考虑的。

```go
// mheap.go

type mspan struct {
    allocBits  *gcBits
    gcmarkBits *gcBits    
}
```

```go
// mgcsweep.go

// sweepLocked represents sweep ownership of a span.
type sweepLocked struct {
    *mspan
}


// Sweep frees or collects finalizers for blocks not marked in the mark phase.
// It clears the mark bits in preparation for the next GC round.
// Returns true if the span was returned to heap.
//
// If preserve=true, don't return it to heap nor relink in mcentral lists;

func (sl *sweepLocked) sweep(preserve bool) bool {

    s := sl.mspan
    spc := s.spanclass
    size := s.elemsize

    // 基于标记位图统计已分配数量 (不含可回收部分)。
    nalloc := uint16(s.countAlloc())
    
    // 计算本次回收数量 (标记前分配数量 - 标记后分配数量)。
    nfreed := s.allocCount - nalloc

    // 调整属性。
    s.allocCount = nalloc
    s.freeindex = 0 // reset allocation index to start of span.

    // 将标记位图直接当作分配位图使⽤，便可实现 “复制”。
    s.allocBits = s.gcmarkBits
    s.gcmarkBits = newMarkBits(s.nelems)

    // 初始化 nextFreeIndex 缓存。
    s.refillAllocCache(0)

    if spc.sizeclass() != 0 {

           // 小对象 ...

        if !preserve {

            // 回收后分配数量为 0，表示未使用，交还给堆。
            if nalloc == 0 {
                mheap_.freeSpan(s)
                return true
            }

            // 根据是否有剩余空间，放到 mcentral 合适列表内。
            if uintptr(nalloc) == s.nelems {
                mheap_.central[spc].mcentral.fullSwept(sweepgen).push(s)
            } else {
                mheap_.central[spc].mcentral.partialSwept(sweepgen).push(s)
            }
        }
    } else if !preserve {

        // 大对象 ...

        // 整个 mspan 仅一个对象，如果释放数不等于 0，那么归还给堆。
        if nfreed != 0 {
            mheap_.freeSpan(s)
            return true
        }

        // 没有释放，那么放到 mcentral 列表。
        mheap_.central[spc].mcentral.fullSwept(sweepgen).push(s)
    }
    
    return false
}
```

&nbsp;

### 二级平衡

首先，将 `P.cache.alloc` 数组内 `mspan` 上交 `central`，有剩余内存的 `mspan` 可调度给其他 `P.cache` 使用。此第一级平衡，避免P长时间闲置内存。

其次，将收回全部空间的`mspan`从`central`交换给`heap`，那么该mspan可以调剂给其他`cental`使用。无非重置属性，诸如`spanclass`之类的，完全没有影响。这就是第二级平衡，避免`central`长时间闲置内存。

总之，要充分使用闲置内存，而非向操作系统申请。
