# 信号量

### 定义
信号量是一种同步机制,它可以用来控制有多少个并发任务可以同时执行。Go语言是从原子上通过一个计数器来实现。

为了防止出现因多个程序同时访问一个共享资源而引发的一系列问题,我们需要一种方法,它可以通过生成并使用令牌来授权,在任一时刻只能有一个执行线程访问代码的临界区域。临界区域是执行数据更新的代码需要独占式地执行。而信号量就可以提供这样的一种访问机制,让一个临界区同一时间只有一个线程在访问它,也就是说信号量时用来协调进程对共享资源的访问的。

信号量是一个特殊的变量,程序对其访问的都是原子操作,且只允许对它进行等待(P(信号变量))和发送(V)信息操作。最简单的信号量是只能取0和1的变量,这也是信号量最常见的一种形式,叫做二进制信号量。而可以取多个正整数的信号量被称为通用信号量。


### 信号量实现源码剖析

#### Semaphore
Go语言信号量是在runtime实现,在runtime实现主要以下两个原因:

* 提升性能。
* 配合调度器实现阻塞、休眠、唤醒等。

Go语言Semaphore给予计数器(整数)实现信号量。

* acquire: 调用cansemacquire，获取并递减信号计数。
* release: 释放并递加信号计数。

备注: 信号计数为0时,请求操作阻塞。

#### semacquire
semacquire获取信号具体实现。

```
func semacquire(addr *uint32) {
	semacquire1(addr, false, 0, 0)
}

func semacquire1(addr *uint32, lifo bool, profile semaProfileFlags, skipframes int) {
	gp := getg()
	
	// 直接可以获取信号, 返回。
	if cansemacquire(addr) {
		return
	}
	
	// 如果无法获取到信号,如何做？
	// 1. 信号等待计数器自增1
	// 2. 将其放入休眠队列之前,再垂死挣扎一次看是否可以获得信号(调用cansemacquire函数)
	// 3. 垂死挣扎后还没有获得信号,则将放入等待队列
	// 4. 休眠
	
	// 获取信号失败,当前内存地址作为key,存储到sudog.elem
	s := acquireSudog()
	root := semroot(addr)
	
	// 将所有抢占该地址的等待者放入等待队列
	for {
		lock(&root.lock)
		
		// 增加等待计数
		atomic.Xadd(&root.nwait, 1)
		
		// 在进入休眠队列之前,再检查是否可以获得信号,如果可以获得信号,退出
		if cansemacquire(addr) {
			// 如果获取到了信号,刚增加的等待计数器减1
			atomic.Xadd(&root.nwait, -1)
			unlock(&root.lock)
			break
		}
		
		// 确定无法获取信号的情况下,将等待者信息添加到队列后休眠
		// 等待goroutine重新被唤醒,进入待运行队列(goparkunlock).
		root.queue(addr, s, lifo)
		goparkunlock(&root.lock, waitReasonSemacquire, traceEvGoBlockSync, 4+skipframes)
		
		// 唤醒后,检查是否能够获取信号量
		// 如果可以获取信号,则退出;否则继续循环,进入等待队列 -> 休眠 -> 被goparkunlock唤醒 -> 检查是否可以获取到信号(调度循环)。
		if s.ticket != 0 || cansemacquire(addr) {
			break
		}
	}
	
	releaseSudog(s)
}
```

* cansemacquire

```
// 获取并递减信号计数
// 如果可以获取到信号,返回true,否则返回false
func cansemacquire(addr *uint32) bool {
        for {
        		   // 判断当前信号计数是否为0,如果为0,则返回false.
                v := atomic.Load(addr)
                if v == 0 {
                        return false
                }
                
                // 信号计数不为0时,获取信号并将计数器减1
        			// 原子操作
                if atomic.Cas(addr, v, v-1) {
                        return true
                }
        }
}
```

#### semrelease
semrelease是Go语言对信号量的释放,即信号量计数器自增1。

```
func semrelease(addr *uint32) {
	semrelease1(addr, false, 0)
}

func semrelease1(addr *uint32, handoff bool, skipframes int) {
	// 增加信号计数
	root := semroot(addr)
	atomic.Xadd(addr, 1)
	
	// 检查是否有等待者,如果没有等待者,直接退出
	if atomic.Load(&root.nwait) == 0 {
		return
	}
	
	// 如果有等待者,搜索等待者并且唤醒
	lock(&root.lock)
	
	// 在Lock之后,再次判断等待者队列是否为空
	// 如果为空,退出
	if atomic.Load(&root.nwait) == 0 {
		// 在并发情况下,等待者可能被其它的goroutine唤醒
		unlock(&root.lock)
		return
	}
	
	// 等待队列是一个链表数据结构
	// 唤醒等待者步骤:
	// 1. 从等待队列中找到目标信号量,移除等待信息。
	// 2. 信号量计数器自减1
	// 3. 从等待队中pop出一个等待者
	s, t0 := root.dequeue(addr)
	if s != nil {
		atomic.Xadd(&root.nwait, -1)
	}
	unlock(&root.lock)
	
	if s != nil { // May be slow, so unlock first
		// 唤醒某个请求者
		// 1. 将请求者放入运行的等待队列。
		// 2. 至于是否可以执行,由Go调度器决定。
		readyWithTime(s, 5+skipframes)
	}
}
```