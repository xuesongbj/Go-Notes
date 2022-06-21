# slice
slice类型是一个组合类型,它由头部和底层数组两部分组成。其中header包括ptr,len和cap三个字段,共24byte。ptr指向底层存储数据的字节数组,len表示slice元素数据长度,cap是ptr指向底层数组的长度。


## slice 类型
slice通常使用make(type, len, cap)进行创建。make是内置类型,分配并初始化一个类型的对象slice,其返回值为type类型。

使用make创建slice时,如果cap是一个常量表达式,make通常在goroutine stack上分配内存(无需进行内存回收)。如果make的大小分配太大,例如:常量(64*1024)或变量(i+1),那么这两个都会在heap上分配(使用makeslice进行内存分配).

```
 +---------------------+
 | Ptr | len |   cap   |
 +---------------------+
 |
 |
 |   +-------------------+
 +-->| 0  | 0 |  0 |  0  |
     +-------------------+ 
```

### slice header 
```
type slice struct {
 array unsafe.Pointer          // 指向底层数组
 len   int                     // slice长度
 cap   int                     // 底层数组长度
}
```


### cap实例1: cap常量
```
// go:noinline
func f1() {
	s := make([]int, 0, 128)
	_ = s
}

// 编译
$> go build -gcflags "-l -m -N" -o test main.go
\# command-line-arguments
// 没有发生逃逸 
./main.go:10:14: main make([]int, 0, 128) does not escape         
```


### cap实例2: cap常量较大
```
// go:noinline
func f2() {
	s := make([]int, 0, 10000)
	_ = s
}

// 编译
$> go build -gcflags "-l -m -N" -o test main.go
\# command-line-arguments
./main.go:10:14: make([]int, 0, 10000) escapes to heap                   // make发生了逃逸 
```


### cap实例3: cap是变量
cap为变量i,i是一个运行时变量表达式,编译器必须进行数据流分析以证明make cap变量i值为1。

```
func f3() {
	var i int = 1
	s := make([]int, 0, i)
	_ = s
}

// 编译
$> go build -gcflags "-l -m -N" -o test main.go
\# command-line-arguments
./main.go:10:14: make([]int, 0, i) escapes to heap                    // make发生了逃逸 
```


## 创建slice
```
func makeslice(et *_type, len, cap int) slice {
	// maxElements: 可以分配最大cap长度
	maxElements := maxSliceCap(et.size)

	// 这里进行判读,主要是纺织len太长,超出内存寻址方位,具体可以看:
	// issue 4085
	if len < 0 || uintptr(len) > maxElements {
		panicmakeslicelen()
	}

	if cap < len || uintptr(cap) > maxElements {
		panicmakeslicecap()
	}

	// 创建slice
	// 被分配在heap上
	p := mallocgc(et.size*uintptr(cap), et, true)
	return slice{p, len, cap}
} 
```

### slice长度
```
var maxElems = [...]uintptr{
	^uintptr(0),
	maxAlloc / 1, maxAlloc / 2, maxAlloc / 3, maxAlloc / 4,
	maxAlloc / 5, maxAlloc / 6, maxAlloc / 7, maxAlloc / 8,
	maxAlloc / 9, maxAlloc / 10, maxAlloc / 11, maxAlloc / 12,
	maxAlloc / 13, maxAlloc / 14, maxAlloc / 15, maxAlloc / 16,
	maxAlloc / 17, maxAlloc / 18, maxAlloc / 19, maxAlloc / 20,
	maxAlloc / 21, maxAlloc / 22, maxAlloc / 23, maxAlloc / 24,
	maxAlloc / 25, maxAlloc / 26, maxAlloc / 27, maxAlloc / 28,
	maxAlloc / 29, maxAlloc / 30, maxAlloc / 31, maxAlloc / 32,
}

	// maxAlloc是分配的最大大小
	maxAlloc = (1 << heapAddrBits) - (1-_64bit)*1

	// maxSliceCap returns the maximum capacity for a slice.
	func maxSliceCap(elemsize uintptr) uintptr {
 	
 	// 如果elemsize小于maxElems长度,则直接从maxElems数组获取
	if elemsize < uintptr(len(maxElems)) {
		return maxElems[elemsize]
 	}
 
 	// 根据类型大小获取长度
 	return maxAlloc / elemsize
}
```

## 内存分配
这部分内容其实属于内存分配器范畴,在以上例子中调用到mallocgc,这里提前讲解。
```
func mallocgc(size uintptr, typ *_type, needzero bool) unsafe.Pointer {
	// 如果申请底层数组大小为0,则返回0x00000000内存地址
	if size == 0 {
		return unsafe.Pointer(&zerobase)
	}

	// debug.sbrk 非0,在stack进行内存创建(function/type/debug-related等类型)
	// debug.sbrk 默认值0,采用内存分配器进行分配
	if debug.sbrk != 0 {
		align := uintptr(16)
		if typ != nil {
			align = uintptr(typ.align)
		}
		return persistentalloc(size, align, &memstats.other_sys)
	}

	// assistG is the G to charge for this allocation, or nil if
	// GC is not currently active.
	var assistG *g
	if gcBlackenEnabled != 0 {
		// Charge the current user G for this allocation.
		assistG = getg()
		if assistG.m.curg != nil {
			assistG = assistG.m.curg
		}
		
		// Charge the allocation against the G. We'll account
		// for internal fragmentation at the end of mallocgc.
		assistG.gcAssistBytes -= int64(size)

		if assistG.gcAssistBytes < 0 {
			// This G is in debt. Assist the GC to correct
			// this before allocating. This must happen
			// before disabling preemption.
			gcAssistAlloc(assistG)
		}
	}

	// Set mp.mallocing to keep from being preempted by GC.
	// 获取当前Goutine m(执行线程)的结构
	mp := acquirem()
    
	// 判断当前m是否正在执行分配操作
	if mp.mallocing != 0 {
		throw("malloc deadlock")
	}

	// 判断当前是否正在执行当前g分配任务
	if mp.gsignal == getg() {
		throw("malloc during signal")
	}

    // 抢占当前的分配信号,一旦被抢占其它分配操作则处于阻塞
	mp.mallocing = 1

	shouldhelpgc := false
	dataSize := size
	c := gomcache()
	var x unsafe.Pointer
	noscan := typ == nil || typ.kind&kindNoPointers != 0

	// Tiny allocator 一种微小的内存分配器.
	// 当申请内存小于maxSmallSize时,可以将这种小内存对象放在同一个内存块.当所有子对象都不可访问时,
	// 就可以将该内存块进行释放。
	// 使用Tiny allocator前提申请对象内不能包含指针类型子对象,否则禁止使用Tiny allocator,主要从
	// 内存浪费角度考虑.
	// Tiny allocator 主要永在小字符串和逃逸的变量。实验室对json对象压力测试,可以减少大约12%分配
	// 数量,减少heap大小越20%.
	// maxSmallSize 默认值32KB(俗称小对象)
	if size <= maxSmallSize {
		// 分配的对象不包含指针,并且分配内存大小小于maxTinySize(默认16字节,可以动态调整)
		if noscan && size < maxTinySize {
			// 进行内存对齐
			off := c.tinyoffset
			if size&7 == 0 {
				off = round(off, 8)
			} else if size&3 == 0 {
				off = round(off, 4)
			} else if size&1 == 0 {
				off = round(off, 2)
			}
			
			// tiny指针指向内存块的开始位置;如果tiny为nil,则表示没有分配内存块。
			// 1. 偏移量(当前已经占用的内存块)+新分配对象大小 <= tiny memory block大小
			// 2. tiny memory block存在
			// 以上2个条件都存在,则这个新对象命中了一个tiny block块。
			if off+size <= maxTinySize && c.tiny != 0 {
				x = unsafe.Pointer(c.tiny + off) // slice header ptr指向底层数组的内存地址
				c.tinyoffset = off + size        // tiny 偏移量更改
				c.local_tinyallocs++             // 计数器
				mp.mallocing = 0
				releasem(mp)                     // 释放m分配权
				return x
			}

			// 分配一个新的 tiny block
			span := c.alloc[tinySpanClass]
			v := nextFreeFast(span)
			x = unsafe.Pointer(v)               // slice header ptr指向底层数组的内存地址
			(*[2]uint64)(x)[0] = 0
			(*[2]uint64)(x)[1] = 0
			size = maxTinySize
		} else {
			// 这两个数组用于根据对象的大小得出相应的类的索引
			// size_to_class8用于大小小于1KB的对象
			// size_to_class128用于 1 – 32KB大小的对象
			var sizeclass uint8
			if size <= smallSizeMax-8 {
				sizeclass = size_to_class8[(size+smallSizeDiv-1)/smallSizeDiv]
			} else {
				sizeclass = size_to_class128[(size-smallSizeMax+largeSizeDiv-1)/largeSizeDiv]
			}

			// class_to_size用于将类(这里指其在全局类列表中的索引值)映射为其所占内存空间的大小
			size = uintptr(class_to_size[sizeclass])

			// 从heap申请一块sizeclass大小的空间
			spc := makeSpanClass(sizeclass, noscan)
			span := c.alloc[spc]

			// 从缓存中获取一块儿空闲内存
			v := nextFreeFast(span)
			if v == 0 {
				// 如果获取失败,则需要重新开辟一块儿内存,填满缓存,然后再向缓存中获取内存。
				// 再填满缓存之前,还是会尝试在缓存中是否可以获得空闲内存
				// 如果填满缓存池,需要触发GC
				v, span, shouldhelpgc = c.nextFree(spc)
			}

			// slice header ptr指向底层数组的内存地址
			x = unsafe.Pointer(v)
		}
	} else {
		// 申请大对象内存,直接从主存申请
		var s *mspan
		shouldhelpgc = true
		systemstack(func() {
			s = largeAlloc(size, needzero, noscan)
		})
		s.freeindex = 1
		s.allocCount = 1

		// slice header ptr指向底层数组的内存地址
		x = unsafe.Pointer(s.base())
	  	size = s.elemsize
	}

	return x
} 
```

## slice扩容
当slice空间不足需要扩容时,需要调用growslice进行扩容.

### 实例1:
```
func main() {
     s := make([]int, 0, 2)
     int a []int = []int{1, 2, 3, 4, 5,6}
     s = append(s, a)
}
```

* 反汇编

```
0x000000000104abfc <+300>:    mov    QWORD PTR [rsp+0x8],rdx
0x000000000104ac01 <+305>:    mov    QWORD PTR [rsp+0x10],rcx
0x000000000104ac06 <+310>:    mov    QWORD PTR [rsp+0x18],rax
0x000000000104ac0b <+315>:    mov    rax,QWORD PTR [rsp+0x48]
0x000000000104ac10 <+320>:    mov    QWORD PTR [rsp+0x20],rax
0x000000000104ac15 <+325>:    call   0x10333c0 <runtime.growslice> 
```

### Growslice 源码实现
slice 在进行append()时,当cap容量不够用时,才会调用growslice函数进行内存扩容。该函数至少返回一个新slice容量的长度;新slice长度还是为旧slice长度,不是新slice容量;计算新加入元素的位置。
 
SSA在Go1.7版本被引入,作为Go新的后端。SSA后端更喜欢新slice长度为旧slice长度或仅返回指针以节省stack空间.现在还是使用之前的方式,会持续跟进Go growslice发展。

```
// et : slice类型
// old : 旧slice
// cap : 所需要的cap
func growslice(et *_type, old slice, cap int) slice {
	// 数据类型为nil
	if et.size == 0 {
		// append不能创建slice len大于0的nil指针
		// 这种情况下,slice不需要创建底层数组。直接返回nil指针即可
		return slice{unsafe.Pointer(&zerobase), old.len, cap}
	}

	newcap := old.cap
	doublecap := newcap + newcap
	if cap > doublecap {
		// 所需要的cap > 两倍旧slice.cap大小,新的cap为所需要的cap大小
		newcap = cap
	} else {
		if old.len < 1024 {
			// 如果所需要的cap小于旧slice.cap*2
			// 旧的slice len小于1024
			// 扩容后的slice cap大小为就slice cap的2倍
			newcap = doublecap
		} else {
			// 旧的slice len大于1024,则扩容后的cap以1/4的比例增长
			for 0 < newcap && newcap < cap {
				newcap += newcap / 4
			}

			// 旧的slice没有指定cap长度,则扩容后的cap长度为所需的cap长度
			if newcap <= 0 {
			newcap = cap
			}
		}
	}

	// 以下主要计算扩容前后内存占用大小,针对几种常见的类型进行了优化处理
	// overflow: 判断申请的cap大小是否导致heap溢出
	// lenmem: 旧slice len内存大小
	// newlenmem: 扩容后slice len内存大小
	// capmem: 扩容后slice cap内存大小
	var overflow bool
	var lenmem, newlenmem, capmem uintptr
	switch {
	case et.size == 1:
		lenmem = uintptr(old.len)
		newlenmem = uintptr(cap)
		capmem = roundupsize(uintptr(newcap))       // 通过mallocgc创建size大小内存,作为扩容后slice的底层数组
		overflow = uintptr(newcap) > maxAlloc
		newcap = int(capmem)
	case et.size == sys.PtrSize:
		lenmem = uintptr(old.len) * sys.PtrSize
		newlenmem = uintptr(cap) * sys.PtrSize
		capmem = roundupsize(uintptr(newcap) * sys.PtrSize)
		overflow = uintptr(newcap) > maxAlloc/sys.PtrSize
		newcap = int(capmem / sys.PtrSize)
	case isPowerOfTwo(et.size):
		// eg: int64,uint64, int64...
		var shift uintptr
		if sys.PtrSize == 8 {
			// sys.Ctz64判断et.size 从低位起有多少个0
			// 16 => 10000  => 4
			// 4 & 63 ==> 4
			shift = uintptr(sys.Ctz64(uint64(et.size))) & 63
		} else {
			shift = uintptr(sys.Ctz32(uint32(et.size))) & 31
		}
		
		lenmem = uintptr(old.len) << shift
		newlenmem = uintptr(cap) << shift
		capmem = roundupsize(uintptr(newcap) << shift)
		overflow = uintptr(newcap) > (maxAlloc >> shift)
		newcap = int(capmem >> shift)
	default:
		lenmem = uintptr(old.len) * et.size
		newlenmem = uintptr(cap) * et.size
		capmem = roundupsize(uintptr(newcap) * et.size)
		overflow = uintptr(newcap) > maxSliceCap(et.size)
		newcap = int(capmem / et.size)
	}

	// 满足以下三个条件,cap发生内存溢出
	// 1. 扩容后cap小于旧cap
	// 2. 扩容后cap大于maxSliceCap(et.size)
	// 3. 扩容后cap内存大于heap内存范围
	if cap < old.cap || overflow || capmem > maxAlloc {
		panic(errorString("growslice: cap out of range"))
	}

	var p unsafe.Pointer
	
	// heap上对内存进行重新分配
	if et.kind&kindNoPointers != 0 {
		// 指针类型
		p = mallocgc(capmem, nil, false)
		memmove(p, old.array, lenmem)

		// 旧数据覆盖扩容后底层数组位置
		// 仅清除不会被覆盖的部分
		memclrNoHeapPointers(add(p, newlenmem), capmem-newlenmem)
	} else {
		p = mallocgc(capmem, et, true)
		memmove(p, old.array, lenmem)
	}

	// 返回扩容后slice类型
	return slice{p, old.len, newcap}
} 
```

## convT2Eslice创建slice
创建slice时除使用make方式创建,还可以通过convT2Eslice创建。

```
var a []int 
```
以上源代码经过编译之后,会通过runtime.convT2Eslice函数创建slice。

```
func convT2Eslice(t *_type, elem unsafe.Pointer) (e eface) {
	var x unsafe.Pointer
	
	// elem数据转换成slice类型
	if v := *(*slice)(elem); uintptr(v.array) == 0 {
		// 如果slice array指针为空,则返回nil
		x = unsafe.Pointer(&zeroVal[0])
	} else {
		// 申请内存
		x = mallocgc(t.size, t, true)
		*(*slice)(x) = *(*slice)(elem)
	}

	// 返回
	e._type = t
	e.data = x
	return
}
```

