# Go 内存分配器概念
Go最初内存分配器是基于tcmalloc,现在版本内存分配器已经偏离很多。


## 涉及数据结构

* fixalloc: 固定大小的内存分配器(Go默认内存分配器)。
* mheap: 分配堆,以8KB为粒度管理内存页。
* mscan: 由mheap管理页面。
* mcentral: 收集所有类型大小的spans。
* mcache: 每个P上空闲的mspans缓存。
* mstats: 分配统计信息。


## 内存申请

### 内存申请顺序
mcache -> mcentral -> mheap


### 申请小对象
P mcache -> mcentral mspan -> mheap page -> system

* 如果本地P mcache有该类型classes的span,则进行扫描mspan;如果mspan中有空闲槽位,直接分配。此操作为无锁操作。
* 如果mspan没有空闲槽位,则从mcentral指定类型的mspans列表中获取可用的mspan。使用共享锁进行获取。
* 如果mcentral中没有可用的mspan,则从mheap获取一批内存页分割成指定类型mspan。
* 如果mheap没有可用的内存页或没有足够大的内存页,则从操作系统申请(至少1MB)。

### 内存分配器设计名词

* FixAlloc: 是一个简单的固定大小对象无锁列表内存分配器.用于管理MCache和MSpan对象.
* treapalloc: treapNodes的大对象内存分配器
* spanalloc: span的内存分配器
* cachealloc: mcache的内存分配器
* specialfinalizeralloc: specialprofile的内存分配器(specialfinalizer是从非GC内存分配的,因此必须专门处理heap上指针)
* specialprofilealloc: specialprofile的内存分配器(用于堆内存分析)
* arenaHintAlloc: heapHints内存分配器(管理heap内存增长) 


## 内存回收

### 设计思路

* 如果在内存分配时进行空闲内存扫描操作,则空闲的mspan继续放到当前P mcache中。
* 如果mspan中仍然存在对象,则将该类型对象放到全局缓存mcentral 空闲的mspan中。
* 如果mspan中所有对象已经被释放,则把该mspan返回给mheap。可以和相邻空闲的mspans合并。
* 如果mspan长时间处于空闲状态,则将mheap page返还给操作系统。
* 分配和释放大对象直接使用mheap，绕过mcache和mcentral。



## 虚拟内存

* heap是由一组arena组成。64位为64MB,32位为4MB(heapArenaBytes)。
* 每个arena都有一个heapArena对象,用于存储arena的元数据。
* 由于arenas是内存对齐的,地址空间可以看作是一系列arenas帧。arena被映射成*heapArena、nil。arena struct由L1、L2两级数组组成。
* arena覆盖整个可能的地址空间,分配器尽可能保证arena连续,以便可以跨多个arena。

```
+-------------+--------------+------------------------------+
|             |              |                              |
|             |              |                              |
|   span      |   bitmap     |            arena             |
|             |              |                              |
|             |              |                              |
+-------------+--------------+------------------------------+                                                      
	512MB           16GB                   512GB    

```

* arena区域就是通常所说的heap,Go从heap分配的内存都在这个区域中。
* bitmap区域用于表示arena区域中哪些地址保存了对象,并且对象中哪些地址包含了指针。bitmap区域1byte对应了arena区域中的4个指针大小的内存,也就是2bit对应一个指针大小的内存。所以bitmap区域大小是512GB/指针大小(8byte)/4 = 16GB。

## 常量

```
maxTinySize = 16byte    		// 最大微小对象大小
tinySizeClass = 2       		// span存储微小对象类型按2字节
maxSmallSize = 32768   	 		// 最大小对象大小
pageShift = _PageShift
pageSize  = 8KB   				// 内存页大小
pageMask  = _PageMask

_StackCacheSize = 32 * 1024  	// P缓存大小(32KB)
_NumStackOrders = 4 - sys.PtrSize/4*sys.GoosWindows - 1*sys.GoosPlan9   									// 指令缓存数量

// heapAddrBits是堆可以寻址范围.
// 
// AMD64:
// 在64位平台上,Go基于硬件和操作系统限制的组合将其限制为48位.
// amd64硬件限制地址为48位，符号扩展为64位。前16位不是全0或全1是“非规范”和无效地址。由于这个限制,在计算arena索引之前,在amd64平台上将地址左移1<47(arenaBaseOffset)。
// 2017年,amd64硬件增加了对57位地址的支持;但是，只有Linux支持此扩展，并且除非使用高于1<<47的提示地址，否则内核将永远不会选择高于1<<47的地址.
// 
// ARM64:
// arm64硬件(ARMv8)将地址限制在1<<48地址范围内。
// 
// 其它:
// ppc64,mips64和s390x支持硬件中的任何64位地址。但是，由于Go仅支持Linux，因此我们依赖于操作系统限制。
// 
// 
// 基于Linux的processor.h，在64位架构上，用户地址空间限制如下:
// Architecture  Name              Maximum Value (exclusive)
// ---------------------------------------------------------------------
// amd64         TASK_SIZE_MAX     0x007ffffffff000 (47 bit addresses)
// arm64         TASK_SIZE_64      0x01000000000000 (48 bit addresses)
// ppc64{,le}    TASK_SIZE_USER64  0x00400000000000 (46 bit addresses)
// mips64{,le}   TASK_SIZE64       0x00010000000000 (40 bit addresses)
// s390x         TASK_SIZE         1<<64 (64 bit addresses)
// 
// 
// 32位平台:
// 在32位平台上,Go完全支持完整的32位地址空间。Mips32只能访问2GB虚拟内存,因此地址寻址限制在1<<31位。
heapAddrBits = (_64bit*(1-sys.GoarchWasm))*48 + (1-_64bit+sys.GoarchWasm)*(32-(sys.GoarchMips+sys.GoarchMipsle))


// heap最大分配空间大小
// 在64位上,理论上可以分配1<<heapAddrBits字节.
// 在32位上,这是一个小于1<<32,因为地址空间中的字节数实际上不适合uintptr。
maxAlloc = (1 << heapAddrBits) - (1-_64bit)*1


// 堆地址中位数、Arena大小及L1和L2映射相关内容.
// 公式: (1 << addrBits) = arenaBytes * L1entries * L2entries
// Go通过计算,得出如下分析结果:
//       Platform  Addr bits  Arena size  L1 entries  L2 size
// --------------  ---------  ----------  ----------  -------
//       */64-bit         48        64MB           1     32MB
// windows/64-bit         48         4MB          64      8MB
//       */32-bit         32         4MB           1      4KB
//     */mips(le)         31         4MB           1      2KB


// Arena堆大小
// Heap是由Arena组成,按heapArenaBytes进行内存对齐.初始化时由1个arena映射组成
// logHeapArenaBytes是heapArenaBytes 2倍数
heapArenaBytes = 1 << logHeapArenaBytes


// Arena位图大小
heapArenaBitmapBytes = heapArenaBytes / (sys.PtrSize * 8 / 2)

// 每个Arena由多少内存组组成
pagesPerArena = heapArenaBytes / pageSize

// Arena L1位数
arenaL1Bits = 6*(_64bit * sys.GoosWindows)

// Arena L2位数(32MB)
arenaL2Bits = heapAddrBits - logHeapArenaBytes - arenaL1Bits

// Arena栈帧位移操作,以计算L1 arena索引位置
arenaL1Shift = arenaL2Bits

// Arena L1和L2索引总位数
arenaBits = arenaL1Bits + arenaL2Bits

// Arena map中开始位置指针
arenaBaseOffset uintptr = sys.GoarchAmd64 * (1 << 47)

// 垃圾回收最大线程数
_MaxGcproc = 32


// 最小的合法指针
// 最小可能是当前硬件架构的最小内存页大小(假设第一页从未映射过),该值应该与编译器中的minZeroPage一致。
minLegalPointer uintptr = 4096

// physPageSize是操作系统物理页面的大小(以字节为单位)。
// 必须在mallocinit之前由OS初始化代码(通常在osinit中)设置
var physPageSize uintptr
```

## 操作系统相关定义

* sysAlloc: 从操作系统获取大量的初始化(归零)内存,通常大约为100KB或1MB.sysAlloc返回是对齐内存,再使用时需要重新调整sysAlloc内存对齐。
* sysUnused: 操作系统不再需要内存区域的内容,可以其它用途使用。
* sysUsed: 通知操作系统再次需要内存区域的内容。
* sysFree: 内存释放,仅在内存不足时才会对其进行操作。
* sysReserver: 保留的地址空间不会被分配,如果是一个不可用的指针,则选择一个其它区域空间。
* sysMap: 使用以前保留的地址空间。
* sysFault: 内存空间访问异常。

## 非托管内存(不使用FixAlloc内存分配器)
以下三种分配方式都应该标记为 // go:notinheap

* sysAlloc: 直接从OS获取内存，可以获取任何系统页大小(4k)整数倍的内存，也可以被sysFree释放。

* persistentalloc: 把多个小内存组合为单次sysAlloc防止内存碎片。 然而，没有办法释放其分配的内存。内存不会分配在heap区域。

* fixalloc是slab风格的分配器，用于分配固定大小的对象。 fixalloc分配的对象可以被释放，但是这个内存可能会被fixalloc pool复用， 因此它只能给相同类型的对象复用。分配heap内存,使用GC进行内存回收扫描

