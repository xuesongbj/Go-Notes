// A MemStats records statistics about the memory allocator.
type MemStats struct {
        // General statistics.

        // heap分配字节大小
        Alloc uint64

        // heap累计分配字节大小
        // TotalAlloc在分配堆对象时增加,但与Alloc和HeapAlloc不同,它在释放对象时不会减少
        TotalAlloc uint64

        // Sys是从OS获得的内存总字节数
        // Sys是这些字段的总和.HeapSys,StackSys,MSpanSys,MCacheSys,GCSys,OtherSys
        // Sys测量的大小是运行时heap、stack和其它内部数据结构的VA空间.
        Sys uint64

        // 扫描指针的数量,主要用于runtime debug内部调试
        Lookups uint64

        // Mallocs是分配的堆对象的累计计数.
        Mallocs uint64

        // Frees 是释放的堆对象的累计计数.
        Frees uint64


        // 堆内存统计信息
        // 理解堆统计信息需要了解Go如何组织内存。Go将堆内存的虚拟地址空间划分为"spans",它们是8KB或更大内存的连续区域。
        // 跨度可以处于以下三种状态之一:
        // 
        // idle:
        // 处于idle范围不包含任何数据或对象。处于该状态的内存空间的物理内存(PA)可能释放回OS,但虚拟地址空间永远都不会。
        // 这样做优点是,防止内存碎片。当再次需要内存时直接通过MMU将虚拟内存和物理内存映射即可使用。
        // 
        // in use:
        // 使用中的堆空间,像malloc、new或内存逃逸等使用的内存空间都在该区域.
        // 
        // stack:
        // stack span用于goroutine stack。stack不属于heap,span可以被heap或stack使用,但heap和stack不能同时
        // 使用该span。
    

        // HeapAlloc堆内存大小,和上文Alloc相同.
        // HeapAlloc包括所有的使用中的对象、GC垃圾回收器尚未释放的无法访问的对象。
        HeapAlloc uint64

        // HeapSys是从操作系统申请的堆内存大小.
        // HeapSys是heap的虚拟内存空间大小,其中包括已保留但尚未使用的VA空间(它不占用任何物理内存)以及返回给操作系统的VA地址空间(VA空间已经和PA空间解绑)
        HeapSys uint64

        // HeapIdle 空闲的Span字节大小.
        // HeapIdle可能会被返回给操作系统,或者重新用于堆内存分配,或者作为stack内存重用.
        // 返回给操作系统量 ＝ HeapIdle - HeapReleased
        HeapIdle uint64

        // HeapInuse 已经使用的Span字节大小.
        // HeapInuse至少存在一个对象,并且只能存储相同大小类型的对象.
        HeapInuse uint64

        // HeapReleased是返回给操作系统的物理内存字节数。
        // 这些回收给操作系统的heap内存是从HeapIdle状态进行回收。
        HeapReleased uint64

        // HeapObjects是分配的堆对象数量。
        // 和HeapAlloc一样,会随着object分配而增加,并随着GC扫描并且释放无法访问的对象而减少。
        HeapObjects uint64


        // Stack 内存统计
        // stack不属于Heap的一部分,但runtime可以重用heap区域的内存,反之亦然。
        // 
        // StackInuse是Stack span字节大小。
        // StackInuse至少需要一个stack,这些stack只能用于相同大小的stack。
        // 没有StackInuse是由于stack内存空间可以自动释放,将未使用的span释放会heap。
        StackInuse uint64


        // StackSys是从OS获取的stack字节大小。
        // StackSys = StackInuse + 直接从操作系统获取的操作系统线程stack内存。
        StackSys uint64
        
        // 堆外内存统计信息
        // 以下统计信息是运行时内部结构,不是从heap分配的内存(通常是因为它们是实现heap的一部分)。
        // 跟Stack或Heap内存不同,分配给这些结构的任何内存都专门用于这些结构。
        // 
        // 这些主要用于调试runtime时内存开销。

        // MSpanInuse是分配Mspan结构的字节大小。
        MSpanInuse uint64

        // MSpanSys是从OS获得的mspan结构的内存字节大小。
        MSpanSys uint64

        // MCacheInuse是分配的mcache结构的字节大小。
        MCacheInuse uint64

        // MCacheSys是从OS获取的mcache结构内存自己数。
        MCacheSys uint64

        // BuckHashSys是分析存储哈希表中的内存字节数。
        BuckHashSys uint64

        // GCSys 是垃圾回收器元数据内存大小。
        GCSys uint64

        // OtherSys是除以上字段外heap外的其它内存字节大小
        OtherSys uint64


        // 垃圾收集器分析统计
        // 垃圾收集器的目标是保持HeapAlloc<=NextGC,在每个GC循环结束时,基于可到达的数量的量和GOGC的值来计算下一个循环的目标。

        // NextGC是下一个GC循环的目标Heap大小.
        NextGC uint64

        // LastGC是最后一次垃圾收集完成的时间。(1970年以来的纳秒)
        LastGC uint64

        // PauseTotalNs是程序启动以来GC STW累计时间(ns)。
        // STW:在此期间,所有Goroutine都暂停,只有垃圾收集器可以运行。
        PauseTotalNs uint64

        // PauseNs是最近GC STW循环缓冲区(以Ns为单位)。
        // 最近的STW的是PauseNs[(NumGC + 255) % 256]。
        // 通常,PauseNs[N%256]记录在最近的N%256th GC循环中暂停的时间。
        // 每个GC循环可能有多个暂停;这是一个周期中所有暂停的总和。
        PauseNs [256]uint64

        // PauseEnd 最近GC暂停结束时间的循环缓冲区(自1970年以来的纳秒)。
        // 此缓冲区的填充方式与PauseNs相同。可能有每个GC周期多次暂停;这记录了结束的最后一个循环暂停。
        PauseEnd [256]uint64

        // NumGC 是已完成GC循环数量。
        NumGC uint32

        // NumForcedGC被应用程序强制调用GC函数的数量. 
        NumForcedGC uint32

        // GCCPUFraction是程序启动后GC程序可用cpu时间的一小部分。
        // GCCPUFraction取值范围0~1之间的数字,其中0表示GC没有消耗该程序的cpu。
        // 程序可用CPU时间定义为自程序启动以来GOMAXPROCS的积分。也就是说,如果GO
        // MAXPROCS为2并且程序已运行10s,则其"可用CPU"为20s.
        // GCCPUFraction不包括用于写屏障的cpu时间.
        // 
        // 该字段值和GODEBUG=gctrace=1报告的cpu分数相同
        GCCPUFraction float64

        // EnableGC indicates that GC is enabled. It is always true,
        // even if GOGC=off.
        // EnableGC表示GC已开启。即使GOGC=off也表示GC开启
        EnableGC bool

        // 当前尚未使用
        DebugGC bool

        // BySize报告每个大小累的分配统计信息。
        // BySize[N]给出大小为S的分配的统计信息,其中BySize[N-1].Size < S <= BaseSize[N].Size。BySize不会报告大于BySize[60].Size的分配。
        BySize [61]struct {
                // 最大字节大小的对象类型
                Size uint32

                // Mallocs是此大小类中分配的heap对象的累计计数。
                // 累计分配的字节大小 ＝ Size * Mallocs
                // 正在使用对象数 ＝ Mallocs - Frees
                Mallocs uint64

                // Frees是释放heap对象计数
                Frees uint64
        }
}
