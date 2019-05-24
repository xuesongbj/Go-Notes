# Goroutine 源码解析


```
// 创建新的goroutine运行函数fn.放入g等待队列,等待被调度
func newproc(siz int32, fn *funcval) {
    // 参数
    argp := add(unsafe.Pointer(&fn), sys.PtrSize)

    // goroutine
    gp := getg()

    // 调用者pc寄存器
    // 现场恢复时使用
    pc := getcallerpc()

    systemstack(func() {
        newproc1(fn, (*uint8)(argp), siz, gp, pc)
    })
}

// 1. 创建g,然后将g任务放入p等待运行队列.
func newproc1(fn *funcval, argp *uint8, narg int32, callergp *g, callerpc uintptr) {
    // 当前g
    _g_ := getg()

    // fn参数大小
    siz := narg
    siz = (siz + 7) &^ 7

    // 从P本地队列获取G,如果本地队列为空,则批量从全局队列抢夺(Lock).
    // gFree队列内的G状态为_Gdead,未初始化状态.
    _p_ := _g_.m.p.ptr()
    newg := gfget(_p_)

    // 本地P队列、全局队列都为空,创建一个新的G
    if newg == nil {
        newg = malg(_StackMin)              // G最小默认2KB大小
        casgstatus(newg, _Gidle, _Gdead)    // G状态_Gdead(没有执行用户代码)
        allgadd(newg)                       // G放入allgs内(allgs 全局变量)
    }

    totalSize := 4*sys.RegSize + uintptr(siz) + sys.MinFrameSize // extra space in case of reads slightly beyond frame
    totalSize += -totalSize & (sys.SpAlign - 1)                  // align to spAlign
    sp := newg.stack.hi - totalSize
    spArg := sp

    
    // 从g.sched位置开始,清理unsafe.Sizeof(newg.sched)大小内存空间,用于进行现场保护。
    // 当清理的对象不包含heap指针时,才使用 memclrNoHeapPointers进行内存清理过,否则使用typedmemclr进行清理。
    memclrNoHeapPointers(unsafe.Pointer(&newg.sched), unsafe.Sizeof(newg.sched))

    // 将用户栈sp、pc、g放入当前G内部,用户现场保护
    newg.sched.sp = sp
    newg.stktopsp = sp
    newg.sched.pc = funcPC(goexit) + sys.PCQuantum
    newg.sched.g = guintptr(unsafe.Pointer(newg))
    gostartcallfn(&newg.sched, fn)                 // 现场保护,调用gosave

    // 调用者pc寄存器
    newg.gopc = callerpc                           // 调用者PC寄存器
    newg.ancestors = saveAncestors(callergp)       // callergp, 调用者G
    newg.startpc = fn.fn                           // 
    if _g_.m.curg != nil {                         // 当前G已经和M有绑定关系,打标签
        newg.labels = _g_.m.curg.labels
    }

    // G状态更改为_Grunnable
    casgstatus(newg, _Gdead, _Grunnable)

    // G放到p队列
    // 1. 首先尝试放到_p_.runnext槽中(被调度优先级最高).
    // 2. 如果放入_p_.runnext失败,则将g添加到_p_本地队列尾部。
    // 3. 如果本地_p_队列满了,则将g放入全局队列。
    runqput(_p_, newg, true)
}
```
