# 清理

默认情况下，采用后台并发模式执行清理操作。

```go
// mgc.go

func gcSweep(mode gcMode) {
    
    // STW !!!
    assertWorldStopped()

    if gcphase != _GCoff {
        throw("gcSweep being done but phase is not GCoff")
    }

    // 累加代龄，重置状态。
    mheap_.sweepgen += 2
    mheap_.pagesSwept.Store(0)

    // 同步阻塞模式。
    if !_ConcurrentSweep || mode == gcForceBlockMode {

        // 清理所有 spans。
        for sweepone() != ^uintptr(0) {
            sweep.npausesweep++
        }

        // 释放队列。
        prepareFreeWorkbufs()
        for freeSomeWbufs(false) {
        }

        return
    }

    // 并发后台清理。
    if sweep.parked {
        sweep.parked = false
        ready(sweep.g, 0, true)
    }
}
```

&nbsp;

并发清理同样使用专门的goroutine完成。

```go
// proc.go, mgc.go

// The main goroutine.
func main() {
    gcenable()
}

// It kicks off the background sweeper goroutine, the background
// scavenger goroutine, and enables GC.
func gcenable() {
    go bgsweep(c)
}
```

```go
// mgcsweep.go

func bgsweep(c chan int) {
    
    // 全局变量，存储清理状态。
    sweep.g = getg()
    sweep.parked = true
    
    // 休眠。
    goparkunlock(&sweep.lock, waitReasonGCSweepWait, traceEvGoBlock, 1)

    for {

        // 清理所有 spans。
        for sweepone() != ^uintptr(0) {
            sweep.nbgsweep++
            Gosched()
        }

        // 释放队列。
        for freeSomeWbufs(true) {
            Gosched()
        }

        // 检查是否有遗漏未清理。
        if !isSweepDone() {
            // This can happen if a GC runs between
            // gosweepone returning ^0 above
            // and the lock being acquired.
            continue
        }

        // 休眠，等待下个清理周期。
        sweep.parked = true
        goparkunlock(&sweep.lock, waitReasonGCSweepWait, traceEvGoBlock, 1)
    }
}
```

```go
// mgcsweep.go

var sweep sweepdata

// State of background sweep.
type sweepdata struct {
    g       *g
    parked  bool
    started bool
    
    // centralIndex is the current unswept span class.
    // It represents an index into the mcentral span
    // sets. Accessed and updated via its load and
    // update methods. Not protected by a lock.
    //
    // Reset at mark termination.
    // Used by mheap.nextSpanForSweep.
    centralIndex sweepClass    
}
```

&nbsp;

遍历所有使用中的`mspan`， 按标记结果进行清理。

> 有关 mspan.sweep，请阅读《内存分配器：回收》

```go
// mgcsweep.go

// sweepone sweeps some unswept heap span and returns the number of pages returned
// to the heap, or ^uintptr(0) if there was nothing to sweep.
func sweepone() uintptr {

    npages := ^uintptr(0)
    
    // 遍历 mheap.central[] 内所有 mspan。
    for {
        s := mheap_.nextSpanForSweep()

        if s == nil {
            noMoreWork = sweep.active.markDrained()            
            break
        }

        if state := s.state.get(); state != mSpanInUse {
            continue
        }

        // 清理。
        npages = s.npages
        if s.sweep(false) {
            // Whole span was freed. 
        } else {
            // Span is still in-use, so this returned no
            // pages to the heap and the span needs to
            // move to the swept in-use list.
            npages = 0
        }
        break
    }

    // 清理列表已空，尝试释放物理内存。
    if noMoreWork {
        readyForScavenger()
    }

    return npages
}
```

```go
func (h *mheap) nextSpanForSweep() *mspan {
    sg := h.sweepgen
    
    // 按 sizeclass 遍历 mheap.central[] 数组。
    for sc := sweep.centralIndex.load(); sc < numSweepClasses; sc++ {

        spc, full := sc.split()
        c := &h.central[spc].mcentral

        // 分别检查 full 和 partial 列表。
        var s *mspan
        if full {
            s = c.fullUnswept(sg).pop()
        } else {
            s = c.partialUnswept(sg).pop()
        }

        // 记下本次 sizeclass，作为下次调用起点。
        if s != nil {
            // Write down that we found something so future sweepers
            // can start from here.
            sweep.centralIndex.update(sc)
            return s
        }
    }
    
    // Write down that we found nothing.
    sweep.centralIndex.update(sweepClassDone)
    return nil
}
```

&nbsp;

## 代龄

不同于其他语言里描述对象存活周期的代岭概念，此带龄表示mspan清理状态。

```go
// mheap.go

type mheap struct {
    sweepgen uint32 // sweep generation
}

type mspan struct {
    
    // sweep generation:
    // if sweepgen == h->sweepgen - 2, the span needs sweeping
    // if sweepgen == h->sweepgen - 1, the span is currently being swept
    // if sweepgen == h->sweepgen, the span is swept and ready to use
    // if sweepgen == h->sweepgen + 1, the span was cached before sweep began and is still cached, and needs sweeping
    // if sweepgen == h->sweepgen + 3, the span was swept and then cached and is still cached
    // h->sweepgen is incremented by 2 after every GC

    sweepgen    uint32    
}
```
