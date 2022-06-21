# RWmutex(读写锁)
读写锁分为读锁和写锁，读数据的时候上读锁，写数据的时候上写锁。有写锁的时候不可读不可写。有读锁的时候，数据可读，不可写。


* RWmutex可以有多个reader或单个writer。
* RWmutex的零值是未锁定的互斥锁。
* RWmutex不能按值传递。
* 如果Goroutine持有RWmutex进行读取而另一个Goroutine可能会调用Lock,在释放读锁之前，Goroutine不能获取读锁。
* 禁止递归读锁Lock,为保证锁可用。

### RWmutex数据结构

```
sync/rwmutex.go:

type RWMutex struct {
   // 写锁
   // 让多个锁互斥,保证同时只有一个写操作在进行
	w           Mutex
	
	// 写操作信号量
	// 通过唤醒/睡眠控制写操作
	writerSem   uint32
	
	// 读操作信号量
	// 通过唤醒/睡眠控制读操作
	readerSem   uint32
	
	// 等待者计数器
	// 所有读操作的计数器
	readerCount int32

	// 在写操作之前，还有多少读操作正在进行
	readerWait  int32
}
```

### RLock实现
读操作依然可以累加或降低readerCount,但因为这个负值过大,结果依然是负值,从而知道读操作正在等待。当然，计数加减操作结果并不影响readCounter计数,因为只要再次加上阈值，就可以恢复正常。

```
// RLock读锁，RLock不支持递归锁
func (rw *RWMutex) RLock() {	
	// 写操作:
	// 写操作标志就是将一个readerCounter减去一个rwmutexMaxReaders阈值，
	// 这会将其设置为一个极低值。
	// 

	
	// 如果累加计数的结果是负数，表明写操作正在进行
	// 计数累计依然有效,因为低阈值是固定的
	if atomic.AddInt32(&rw.readerCount, 1) < 0 {
		// 获取RLock失败,当前有写锁。
		// 读操作信号休眠, 等待被唤醒,直到获取到读锁
		runtime_SemacquireMutex(&rw.readerSem, false, 0)
	}
}
```

#### 原子操作

```
// AddInt32原子操作
// 将delta值添加到*addr,并返回新值。
func AddInt32(addr *int32, delta int32) (new int32）

// sync/atomic/asm_linux_arm.s
// 汇编实现
TEXT ·AddInt32(SB),NOSPLIT,$0
   // B 跳转指令,跳转至AddUint32执行
	B   ·AddUint32(SB)
	
// sync/atomic/asm_linux.arm.s
TEXT ·AddUint32(SB),NOSPLIT,$0-12
        // 通过寄存器进行数据拷贝
        MOVW    addr+0(FP), R2    // R2为拷贝目标地址(*addr)
        MOVW    delta+4(FP), R4   // R4为数据源(delta)
addloop1:
        MOVW    0(R2), R0
        MOVW    R0, R1
        ADD     R4, R1
        
        // BL跳转
        // 跳转之前会在寄存器R14中保存PC的当前内容，
        // 因此，可以通过将R14的内容重新加载到PC中，
        // 来返回到跳转指令之后的那个指令处执行(现场恢复)。
        // 该指令是实现子程序调用的一个基本但常用的手段。
        BL      cas<>(SB)               // 使用内核CAS
        BCC     addloop1                // 循环
        MOVW    R1, new+8(FP)
        RET
```

#### runtime_SemacquireMutex
用于获取读锁/写锁。

```
// sync/runtime.go:

// s: runtime_SemacquireMutex等待,直到*s大于0时,原子性递减.
// lifo: 如果lifo == true,等待者为队列头部。
// skipframs: skipframes是在跟踪期间省略的帧数,
// 从runtime_SemacquireMutex的调用者计算。
func runtime_SemacquireMutex(s *uint32, lifo bool, skipframes int)

//go:linkname sync_runtime_SemacquireMutex sync.runtime_SemacquireMutex
func sync_runtime_SemacquireMutex(addr *uint32) {
    semacquire(addr, semaBlockProfile|semaMutexProfile)
}


func semacquire(addr *uint32, profile semaProfileFlags) {
    gp := getg()
    if gp != gp.m.curg {
        throw("semacquire not on the G stack")
    }

    // 低成本情况
    if cansemacquire(addr) {
        return
    }

    // 高成本情况:
    // 1. 增加等待者计数器值(waiter count)
    // 2. 再尝试调用一次cansemacquire, 成功了就直接返回(性能考虑)
    // 3. 没成功,就把自己(g)作为一个waiter入休眠队列
    // 4. 休眠
    // (之后waiter的descriptor被signaler用dequeue踢出)

    s := acquireSudog()     		// 获取等待g
    root := semroot(addr)           // sudog等待队列

    // 初始化... 
    t0 := int64(0)
    s.releasetime = 0
    s.acquiretime = 0

    for {
        lock(&root.lock)

        // 休眠队列waiter计数器加1
        atomic.Xadd(&root.nwait, 1)

        // 再尝试调用一次cansemacquire
        if cansemacquire(addr) {
        	// 调用cansemacquire之前,waiter计数器自加1
        	// 调用cansemacquire成功了,需要将waiter计数器自减1
            atomic.Xadd(&root.nwait, -1)
            unlock(&root.lock)

            // 获取信号成功, 退出
            break
        }


        // Any semrelease after the cansemacquire knows we're waiting
        // (we set nwait above), so go to sleep.

        // 将当前g加入等待队列
        root.queue(addr, s)

        // 休眠,等待被唤醒
        goparkunlock(&root.lock, "semacquire", traceEvGoBlockSync, 4)

        // 唤醒之后,调用cansemacquire
        if cansemacquire(addr) {
            break
        }
    }

    // 释放sudog
    releaseSudog(s)
}


// 判断通过原子操作是否可以获取到共享资源的访问权限
func cansemacquire(addr *uint32) bool {
    for {
        v := atomic.Load(addr)
        if v == 0 {
            return false
        }
        if atomic.Cas(addr, v, v-1) {
            return true
        }
    }
}
```

### RUnlock实现
RUnlock用于解锁RLock的资源,它不会影响当前已经开始读的readers。如果RUnlock一个未RLOck的资源，则会出现运行时错误。

```
// sync/rwmutex.go
func (rw *RWMutex) RUnlock() {
	if r := atomic.AddInt32(&rw.readerCount, -1); r < 0 {
		// Outlined slow-path to allow the fast-path to be inlined
		rw.rUnlockSlow(r)
	}
}

func (rw *RWMutex) rUnlockSlow(r int32) {
	if r+1 == 0 || r+1 == -rwmutexMaxReaders {
		race.Enable()
		throw("sync: RUnlock of unlocked RWMutex")
	}
	// A writer is pending.
	if atomic.AddInt32(&rw.readerWait, -1) == 0 {
		// The last reader unblocks the writer.
		runtime_Semrelease(&rw.writerSem, false, 1)
	}
}

// sync/runtime.go
func runtime_Semrelease(sema *uint32)

// runtime/sema.go
//go:linkname sync_runtime_Semrelease sync.runtime_Semrelease
func sync_runtime_Semrelease(addr *uint32) {
    semrelease(addr)
}

func semrelease(addr *uint32) {
    root := semroot(addr)
    atomic.Xadd(addr, 1)

    // 低成本情况: 如果没有waiter,直接退出(该检查必须发生在Xadd之后，以避免错误唤醒)
    if atomic.Load(&root.nwait) == 0 {
        return
    }

    // 高成本情况: 搜索waiter,并唤醒它
    lock(&root.lock)
    if atomic.Load(&root.nwait) == 0 {
        // 唤醒之前再次检查一下等待计数器是否为0
        // 如果为0，说明已经被另一个goroutine消费了
        // 所以，就不需要唤醒其它goroutine了
        unlock(&root.lock)
        return
    }

    // 查找到要释放的g,进行释放
    s := root.head
    for ; s != nil; s = s.next {
        if s.elem == unsafe.Pointer(addr) {
            atomic.Xadd(&root.nwait, -1)
            root.dequeue(s)
            break
        }
    }
    unlock(&root.lock)

    // 可能会很慢，所以先解锁
    if s != nil {
    	// 唤醒一个goroutine
        readyWithTime(s, 5)
    }
}

// 主要功能是唤醒一个goroutine,将该g转换到runnable状态，并将其放入P的本地队列,等待被调度
func readyWithTime(s *sudog, traceskip int) {
    goready(s.g, traceskip)
}

// runtime/proc.go
// 该函数主要就是切换到g0的栈空间,然后执行ready函数
func goready(gp *g, traceskip int) {
    systemstack(func() {
        ready(gp, traceskip, true)
    })
}
```

### Go RWmutex fast-path存在的bug

Go1.12和Go1.13版本RLock和UNlock时对其进行了inline优化，在for loop中,如果RLock和RUnlock被inline之后就没有函数调用了，会产生死锁。因为只有函数调用,才有可能被Go scheduler将当前Goroutine调度出去,flushed才会被执行。

```
func (c *queue) wait() {
    for {
        c.RLock()
        flushed := c.flushed
        c.RUnlock()
        if flushed {
            return
        }
    }
}
```

#### 协作式抢占
当前goroutine的“抢占式”调度依靠的是compiler在函数中自动插入的“cooperative preemption point”来实现的，但这种方式在使用过程中依然有各种各样的问题，比如：检查点的性能损耗、诡异的全面延迟问题以及调试上的困难。近期负责go runtime gc设计与实现的Austin Clements提出了一个proposal：non-cooperative goroutine preemption ，该proposal将去除cooperative preemption point，而改为利用构建和记录每条指令的stack和register map的方式实现goroutine的抢占， 该proposal预计将在go 1.12中实现。

参考文献:

[github issue](https://github.com/golang/go/issues/24543)

[Go scheduler](https://www.ardanlabs.com/blog/2018/08/scheduling-in-go-part2.html)

### 写锁实现

#### Lock加锁
```
// 对写锁进行"加锁"
// 写锁之前，首先将已经获取到读锁的"读者"消费完
func (rw *RWMutex) Lock() {
	// 锁定M，同一时刻仅有一个goroutine可以获得M。
	// 即Mutex
	rw.w.Lock()

	// 获取"写锁"之前，先将剩下的读操作读完。
	r := atomic.AddInt32(&rw.readerCount, -rwmutexMaxReaders) + rwmutexMaxReaders
	if r != 0 && atomic.AddInt32(&rw.readerWait, r) != 0 {
		runtime_SemacquireMutex(&rw.writerSem, false)
	}
}
```

#### Unlock解锁

```
// 对写锁进行"解锁"
// 与互斥锁一样,锁定的RWMutex与特定的goroutine无关。一个goroutine锁定RWMutex,然后由另一个goroutine解锁它。
func (rw *RWMutex) Unlock() {
	// 向读者发送唤醒信号
	r := atomic.AddInt32(&rw.readerCount, rwmutexMaxReaders)
	for i := 0; i < int(r); i++ {
		runtime_Semrelease(&rw.readerSem, false)
	}

	rw.w.Unlock()
}
```