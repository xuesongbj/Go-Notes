# Go内存分配器 - 数据结构

## 数据结构

### Mcache
mcache用于缓存小对象。mcache保存在P资源中,无需进行加锁。

```
type mcache struct {
 	// trigger heap sample after allocating this many bytes
	next_sample int32
	
	// bytes of scannable heap allocated  
	local_scan  uintptr 

	// tiny 实际上是由多个无指针对象共享一个object(节约内存),所以不能直接释放。
	// 必须等待该tiny内所有对象都终结时,才能回收该object。
	// 指向tiny block块开始位置指针,如果没有tiny block,则为nil
	tiny             uintptr     
	
	// tinyoffset偏移量(tiny实际上由多个无指针对象共享一个object(节约内存))
	tinyoffset       uintptr
	
	// number of tiny allocs not counted in other stats     
	local_tinyallocs uintptr 	 


	// The rest is not accessed on every malloc.
	// 以sizeclass为索引管理多个用于分配的span(67种规格,每一种规格包括scan和noscan)
	alloc [numSpanClasses]*mspan 

	
	stackcache [_NumStackOrders]stackfreelist

	
	// 大对象的释放字节数(>maxsmallsize)
	local_largefree  uintptr         
	
	// 大对象的释放数量(>maxsmallsize)         
	local_nlargefree uintptr 
	
	// 小对象的释放数量(<=maxsmallsize)                 
	local_nsmallfree [_NumSizeClasses]uintptr 
}
```

### Mcentral
存储不同规格大小Span全局列表

```
type mcentral struct {
	spanclass spanClass  // span规格(noscan spanClass和scan spanClass)
	nonempty  mSpanList  // 链表: 尚有空闲object的span
	empty     mSpanList  // 链表: 没有空闲object,或已被cache取走的span
}
```

### Mspan
Go fixalloc内存分配器,通过span对内存进行管理。

```
type mspan struct {
	// 双向链表
	next *mspan    
	prev *mspan


	startAddr uintptr 		// span在arena的起始地址
	npages    uintptr 		// span由多少页组成

	// object - 将span按特定大小切分成多个小块,每个小块可存储一个对象
	// freeindex == nelem,则span中没有空闲的object
	// n >= freeindex && allocBits[n/8] & (1<<(n%8)) == 0,则object n是空闲的;否则object是被分配。
	// 从nelem开始的bit是未定义的,不应该被引用.
	freeindex uintptr      // span槽位索引
	nelems uintptr 		   // span中object数量
	allocBits  *gcBits     // span中object的位图

	allocCount  uint16     // 已分配Object数量
	spanclass   spanClass  // span规格
	incache     bool       // 是否由mcache使用
}
```

### Mheap
mheap包括free[]、large数组和其他全局数据.mheap包含mSpanLists,所以不能被分配。

```
type mheap struct {
	free      [_MaxMHeapList]mSpanList 	// 未分配的span列表(page < 127)
	freelarge mTreap                   	// 未分配的span列表(page > 127)
	busy      [_MaxMHeapList]mSpanList 	// 已分配的span列表(page < 127)
	busylarge mSpanList                	// 已分配的span列表(page > 127, >1MB)


	allspans []*mspan 				   	// 所有申请的mspan会被记录


	// 小对象规格的central空闲列表
	// 每个central对应一种sizeclass
	central [numSpanClasses]struct {
		mcentral mcentral
		pad      [sys.CacheLineSize - unsafe.Sizeof(mcentral{})%sys.CacheLineSize]byte
	}

    treapalloc				// treapNodes的大对象内存分配器
    spanalloc				// span的内存分配器
    cachealloc				// mcache的内存分配器
    specialfinalizeralloc	// specialprofile的内存分配器(specialfinalizer是从非GC内存分配的,因此必须专门处理heap上指针)
    specialprofilealloc		// specialprofile的内存分配器(用于堆内存分析)
    arenaHintAlloc 			// heapHints内存分配器(管理heap内存增长)
}
```

### Fixalloc
fixalloc是go语言内存分配器。

```
type fixalloc struct {
	size   uintptr							// 分配内存块规格大小尺寸
	first  func(arg, p unsafe.Pointer)		// called first time p is returned
	arg    unsafe.Pointer
	list   *mlink							// 内存块链表
	chunk  uintptr							// 内存块分配开始地址
	nchunk uint32
	inuse  uintptr							// 内存已用字节数
	stat   *uint64							// 计数器
	zero   bool								// 初始化内存是否使用0填充,默认true
}
```