# Go内存回收管理
当垃圾回收器触发之后,会对heap上内存进行扫描。将一些不再使用的垃圾内存进行标记,调用内存分配器进行内存回收。


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