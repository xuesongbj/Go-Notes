# GMP

## G

### G结构

```
type g struct {
	stack       stack   		  		// 执行栈

	// 当协程stack不够用时,用于stack扩容.
	stackguard0 uintptr
	stackguard1 uintptr

	m              *m			   	// 当前绑定M线程
	sched          gobuf 		   		// 用于保存执行现场
	goid           int64 		   		// goroutine ID
	gopc           uintptr         			// 调用者PC/IP寄存器
	startpc        uintptr         			// 任务函数
}
```

### G状态

```
1. _Gidle = 0: g已经创建,但未初始化。
2. _Grunnable = 1: g放入队列,等待被调度
3. _Grunning = 2: GPM绑定,执行用户态代码
4. _Gsyscall = 3: G分配一个M执行用户态代码;此时g不会被放到p队列;
5. _Gwaiting = 4: runtime期间Goroutine被阻塞,有可能被channel阻塞.此时goroutine不会拥有stack.
6. _Gmoribund_unused = 5: 未用
7. _Gdead = 6: g没有执行用户态代码,可能处于刚退出? 刚被初始化?
8. _Genqueue_unused = 7: 未使用

_Gscan : G的某一种状态,g不执行用户代码.stack由占用_Gscan位的goroutine拥有.
_Gscan         = 0x1000
_Gscanrunnable = _Gscan + _Grunnable // 0x1001
_Gscanrunning  = _Gscan + _Grunning  // 0x1002
_Gscansyscall  = _Gscan + _Gsyscall  // 0x1003
_Gscanwaiting  = _Gscan + _Gwaiting  // 0x1004
```

## P

### P结构

```
type p struct {
	id          int32               		// P id
	status      uint32 				// P状态

	// 本地队列Locked-free
	runqhead uint32                 		// 本地队列头
	runqtail uint32                 		// 本地队列尾
	runq     [256]guintptr          		// 本地队列长度(256)

	// runnext不为空,则优先执行runnext G
	// 而不是从runq中获取G任务执行
	runnext guintptr				// 优先执行
}
```

### P状态
```
1. _Pidle: P处于空闲状态
2. _Prunning: P处于运行状态,执行用户代码; 拥有当前P的M可以更改P状态。
  (M可以将P转换为_Pidle(如果M没有更多任务执行);_Psyscall(进入系统调用);_Pgcstop(停止GC)。M还可以将P的所有权直接交给另一个M)
3. _Psyscall: 进行系统调用,此时P可能被另外一个M窃取(当前P可能和另外一个M绑定)。
  （更改_Psyscall状态必须使用CAS.此时P也有可能被重新抢回来。）
4. _Pgcstop: P处于暂停状态(由于STW),此时M继续和P绑定。如果P状态从_Prunning转换到_Pgcstop时,会导致M释放P。
5. _Pdead: P处于不可用状态(GOMAXPROCS收缩),P大部分资源会被剥夺。
```

## M

### M结构

```
type m struct {
	g0            *g     	    	// 系统系统栈空间
	tls           [6]uintptr   	// TLS
	mstartfn      func()        	// 启动函数
	curg          *g       		// 当前正在运行的G
	p             puintptr 		// 和P绑定,执行g代码
	nextp         puintptr		// 临时存放P
	oldp          puintptr 		// 系统调用之前的绑定的P
	id            int64         	// 线程ID
	preemptoff    string 		// 当preemptoff非空时,禁用抢占调度(disable preemption)
	spinning      bool 		// m处于自旋状态
	incgo         bool   		// 当前M线程执行cgo代码;不会和p绑定
	park          note		// 休眠锁
	schedlink     muintptr 		// 链表
	mcache        *mcache
	createstack   [32]uintptr    	// 创建M线程栈空间
}
```

