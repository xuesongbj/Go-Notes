# Go内存回收管理
当垃圾回收器触发之后,会对heap上内存进行扫描。将一些不再使用的垃圾内存进行标记,调用内存分配器进行内存回收。

## sysmon
在程序main.main函数有一个专门任务定时进行内存回收操作。可以通过反汇编找到该函数。
启动程序 -> _rt0_amd64(程序真正入口) -> rt0_go -> newproc -> newproc1 -> main -> sysmon -> 定时内存回收

### _rt0_amd64

```
TEXT _rt0_amd64(SB),NOSPLIT,$-8
	// main函数参数
	MOVQ    0(SP), DI   // argc
	LEAQ    8(SP), SI   // argv
	
	// 跳转runtime.rt0_go
	JMP runtime·rt0_go(SB)
```

### rt0_go

```

TEXT runtime·rt0_go(SB),NOSPLIT,$0
	...
	MOVL	16(SP), AX		// copy argc
	MOVL	AX, 0(SP)
	MOVQ	24(SP), AX		// copy argv
	MOVQ	AX, 8(SP)
	CALL	runtime·args(SB)
	CALL	runtime·osinit(SB)
	CALL	runtime·schedinit(SB)

	// create a new goroutine to start program
	MOVQ	$runtime·mainPC(SB), AX		// main.main 地址
	PUSHQ	AX
	PUSHQ	$0			// arg size
	CALL	runtime·newproc(SB)
	POPQ	AX
	POPQ	AX
	
	// start this M
	CALL    runtime·mstart(SB)           // 启动
```

### newproc

```
// 创建g0
func newproc(siz int32, fn *funcval) {
    argp := add(unsafe.Pointer(&fn), sys.PtrSize)
    gp := getg()
    pc := getcallerpc()
    systemstack(func() {
        newproc1(fn, (*uint8)(argp), siz, gp, pc)
    })
}

func newproc1(fn *funcval, argp *uint8, narg int32, callergp *g, callerpc uintptr) {
	...

	// main.main函数
	newg.startpc = fn.fn   
	...
}
```

### sysmon定时内存回收

```
// main.main 函数
func main(){
	systemstack(func() {
            newm(sysmon, nil)
    })
}

func sysmon() {
	for {
		// 定时进行内存清理
		if lastscavenge+scavengelimit/2 < now {
			mheap_.scavenge(int32(nscavenge), uint64(now), uint64(scavengelimit))
			lastscavenge = now
		}
	}
}
```
如下就是Go内存分配器内存回收具体实现。


## Mcache内存回收
Mcache内存回收分为两部分:

* 将alloc中未用完的内存回收到mcentral。releaseAll函数将mcentral.alloc各种规格的span回收到mcentral(加锁)。
* 将mcache归还給mheap.cachealloc(mcache插入free list)。

### Mcache内存回收实现

```
func freemcache(c *mcache) {
	systemstack(func() {
		// mcache 内存回收
		c.releaseAll()

		// 释放stack
		stackcache_clear(c)

		// 将mcache回收到mheap.cachealloc（将mcache插入free list表头）
		mheap_.cachealloc.free(unsafe.Pointer(c))
	})
}

```

```
func (c *mcache) releaseAll() {
	for i := range c.alloc {
		s := c.alloc[i]

		if s != &emptymspan {
			// mcache中span回收到mcentral
			mheap_.central[i].mcentral.uncacheSpan(s)
			c.alloc[i] = &emptymspan
		}
	}

	// 清空存放微小对象span池
	c.tiny = 0
	c.tinyoffset = 0
}
```

```
// Return span from an MCache.
func (c *mcentral) uncacheSpan(s *mspan) {
	s.incache = false


	cap := int32((s.npages << _PageShift) / s.elemsize)
	n := cap - int32(s.allocCount)                         // Span未分配对象数量
	if n > 0 {
		// 从mcache中删除span,回收到mcentral 未分配队列.
		c.empty.remove(s)
		c.nonempty.insert(s)
	}
}
```

```
//go:systemstack
func stackcache_clear(c *mcache) {
	for order := uint8(0); order < _NumStackOrders; order++ {
		// 将stack添加到空闲Pool
		x := c.stackcache[order].list
		for x.ptr() != nil {
			y := x.ptr().next
			stackpoolfree(x, order)
			x = y
		}
		c.stackcache[order].list = 0
		c.stackcache[order].size = 0
	}
}
```

```
func (f *fixalloc) free(p unsafe.Pointer) {
	f.inuse -= f.size
	v := (*mlink)(p)
	v.next = f.list
	f.list = v
}
```

## Mcentral内存回收
当span为对象为空时,将span归还給heap。


### Mcentral内存回收实现

```
func (c *mcentral) freeSpan(s *mspan, preserve bool, wasempty bool) bool {
	// 对象在分配时规零
	s.needzero = 1

	// 仅在MCentral_CacheSpan调用时需要被保留该span.该span必须在empty列表内.
	if preserve {
		atomic.Store(&s.sweepgen, mheap_.sweepgen)
		return false
	}

	lock(&c.lock)

	// 如果必要,需要将span插入到nonempty
	if wasempty {
		c.empty.remove(s)
		c.nonempty.insert(s)
	}


	// 更新sweepgen状态
	// seepgen == h->sweepgen ==> 2   span等待被扫描
	// seepgen == h->sweepgen ==> 1	  span正在被扫描
	// seepgen == h->sweepgen ==> 0   span已经被扫描,直接可以使用
	atomic.Store(&s.sweepgen, mheap_.sweepgen)

	// 确保当前span的为空
	if s.allocCount != 0 {
		unlock(&c.lock)
		return false
	}

	// 1. 从mcentral 删除该span
	// 2. 将该span归还到heap上.
	c.nonempty.remove(s)
	mheap_.freeSpan(s, 0)
	return true
}
```


## Mheap内存回收
Mheap并不会立即将回收的空闲内存还给操作系统。它会尝试合并相邻空闲span,然后将合并后的空闲span返回到内存分配器(FixAlloc)的复用链表进行复用。

#### Mspan 四种状态

```
const (
	_MSpanDead   mSpanState = iota      // 未使用未清理  
	_MSpanInUse             			// 被分配在heap上
	_MSpanManual            			// 手工分配(比如在stack上分配)
	_MSpanFree                          // 未使用
)
```

### Mheap 内存回收实现

```
// Mheap内存回收
func (h *mheap) freeSpan(s *mspan, acct int32) {
	systemstack(func() {
		h.freeSpanLocked(s, true, true, 0)
	})
}
```

```
// s must be on a busy list (h.busy or h.busylarge) or unlinked.
func (h *mheap) freeSpanLocked(s *mspan, acctinuse, acctidle bool, unusedsince int64) {
	// 将当前span设置为MSpanFree状态
	s.state = _MSpanFree

	// 将mspan页从busy或busylarge列表移除
	if s.inList() {
		h.busyList(s.npages).remove(s)
	}

	// span标记为空闲状态.当内存回收时候,根据该标记可能将部分页面返回給操作系统.
	s.unusedsince = unusedsince
	if unusedsince == 0 {
		s.unusedsince = nanotime()
	}

	// 归还給操作系统的内存页数量
	s.npreleased = 0

	// 相邻之间空闲span合并
	// 
	// 前一个span不为空 && 前一个span处于_MSpanFree状态,则span进行合并
	if before := spanOf(s.base() - 1); before != nil && before.state == _MSpanFree {
		s.startAddr = before.startAddr  // mspan开始地址调整到前一个span开始地址
		s.npages += before.npages       // mspan页数 ＝ 前一个span页数 + 当前mspan页数
		s.npreleased = before.npreleased // 归还給操作系统的内存页数(不归还給操作系统)
		s.needzero |= before.needzero    // 分配之前归零

		// mspan合并
		h.setSpan(before.base(), s)


		// 合并之后大小发生了改变,需要从treap上删除合并后的节点,并作为新的节点插回。
		if h.isLargeSpan(before.npages) {
			// We have a t, it is large so it has to be in the treap so we can remove it.
			h.freelarge.removeSpan(before)
		} else {
			h.freeList(before.npages).remove(before)
		}

		// 标记为未使用未清理状态
		before.state = _MSpanDead

		// 将合并后的span放到fixalloc(内存分配器)空闲链表,重复利用
		h.spanalloc.free(unsafe.Pointer(before))
	}

	
	// 后一个span不为空 && 后一个span处于_MSpanFree状态,则相邻span进行合并
	if after := spanOf(s.base() + s.npages*pageSize); after != nil && after.state == _MSpanFree {
		s.npages += after.npages           // mspan页数 ＝ 后一个span页数 + 当前mspan页数
		s.npreleased += after.npreleased   // 归还給操作系统的内存页数(不归还給操作系统)
		s.needzero |= after.needzero       // 分配之前归零

		// mspan合并
		h.setSpan(s.base()+s.npages*pageSize-1, s)

		// 合并之后大小发生了改变,需要从treap或链表上删除合并后的节点,并作为新的节点插回。
		if h.isLargeSpan(after.npages) {
			h.freelarge.removeSpan(after)
		} else {
			h.freeList(after.npages).remove(after)
		}

		// 标记为未使用未清理状态
		after.state = _MSpanDead

		// 将合并后的span放到fixalloc(内存分配器)空闲链表,重复利用
		h.spanalloc.free(unsafe.Pointer(after))
	}

	// 将合并后mspan插入到heap.free链表或heap.freelarge mtreap等待被再次利用
	if h.isLargeSpan(s.npages) {
		h.freelarge.insert(s)
	} else {
		h.freeList(s.npages).insert(s)
	}
}
```

### 手工回收Mheap内存

```
// 手工释放内存
// freeManual必须在系统堆栈上调用,防止堆栈增长
func (h *mheap) freeManual(s *mspan, stat *uint64) {
	// 分配之前归零
	s.needzero = 1

	h.freeSpanLocked(s, false, true, 0)
}
```