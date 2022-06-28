# Go程序卡死问题

### 代码段

```
package main

func test(id int, wg *sync.WaitGroup) {
    defer wg.Done()
    
    println(id)
    
    x := 0
    
    // 
    for {
        x++
        
        /*
        if x % 10000 == 0 {
            println(id)
        }
        */
    }
}

func main() {
    runtime.GOMAXPROCS(1)   // MP
    
    var wg sync.WaitGroup
    
    for i := 0; i < 2; i++ {
        wg.Add(1)
        go test(i, &wg)
    }
    
    wg.Wait()
}
```

* 编译 & 运行

```
shell> go build -gcflags "-L -m -n" -o test main.go

shell> ./test 
1            			// 卡死
```

### 问题分析
   当只有一个P资源时，只能有一个并发。当G1长时间执行时，其它G就会被饿死。此时需要将CPU时间片释放,执行其它任务。cpu切换方式:
   
* 手工通过runtime.Gosched函数释放CPU时间片。
* runtime每执行一次G任务，计数器加1。当达到CPU执行时间片，则将P调度出去，执行其它G。

每创建一个新的函数，都会在该函数stack内插入runtime.morestack_noctxt(SB),该函数有两个作用:

* 查看当前stack是否沟通
* 检查是否发生抢占调度,如果有抢占调度，则会让出CPU时间片，去执行其它G。


`以上实例中,test函数for{ x++ }没有函数调用，没有触发runtime.morestack_noctxt函数,无法进行抢占调度，程序出现卡死情况。之所在函数调用的时候进行调度抢占是因为函数切换的时候是一个safe point，这个时候寄存器是空的，上下文切换不需要保存这些寄存器，而且stack中的root point是确定的，能够精确执行GC扫描。`


```
 shszzhang@ZHSZZHANG-MB0  ~/data/mine/application/go_study  go tool objdump -s "main\.test" test
TEXT main.test(SB) /Users/shszzhang/data/mine/application/go_study/main.go
  main.go:38		0x1051e50		65488b0c2530000000	MOVQ GS:0x30, CX
  main.go:38		0x1051e59		483b6110		CMPQ 0x10(CX), SP
  main.go:38		0x1051e5d		7675			JBE 0x1051ed4
  main.go:38		0x1051e5f		4883ec48		SUBQ $0x48, SP
  main.go:38		0x1051e63		48896c2440		MOVQ BP, 0x40(SP)
  main.go:38		0x1051e68		488d6c2440		LEAQ 0x40(SP), BP
  main.go:39		0x1051e6d		c744240808000000	MOVL $0x8, 0x8(SP)
  main.go:39		0x1051e75		488d050c4f0200		LEAQ go.func.*+1074(SB), AX
  main.go:39		0x1051e7c		4889442420		MOVQ AX, 0x20(SP)
  main.go:39		0x1051e81		488b442458		MOVQ 0x58(SP), AX
  main.go:39		0x1051e86		4889442438		MOVQ AX, 0x38(SP)
  main.go:39		0x1051e8b		488d442408		LEAQ 0x8(SP), AX
  main.go:39		0x1051e90		48890424		MOVQ AX, 0(SP)
  main.go:39		0x1051e94		e82718fdff		CALL runtime.deferprocStack(SB)
  main.go:39		0x1051e99		85c0			TESTL AX, AX
  main.go:39		0x1051e9b		7527			JNE 0x1051ec4
  main.go:39		0x1051e9d		eb00			JMP 0x1051e9f
  main.go:41		0x1051e9f		e83c32fdff		CALL runtime.printlock(SB)
  main.go:41		0x1051ea4		488b442450		MOVQ 0x50(SP), AX
  main.go:41		0x1051ea9		48890424		MOVQ AX, 0(SP)
  main.go:41		0x1051ead		e8ae39fdff		CALL runtime.printint(SB)
  main.go:41		0x1051eb2		e8b934fdff		CALL runtime.printnl(SB)
  main.go:41		0x1051eb7		e8a432fdff		CALL runtime.printunlock(SB)
  main.go:45		0x1051ebc		eb00			JMP 0x1051ebe
  main.go:46		0x1051ebe		eb00			JMP 0x1051ec0
  main.go:46		0x1051ec0		eb00			JMP 0x1051ec2
  main.go:46		0x1051ec2		ebfa			JMP 0x1051ebe
  main.go:39		0x1051ec4		90			NOPL
  main.go:39		0x1051ec5		e8f61dfdff		CALL runtime.deferreturn(SB)
  main.go:39		0x1051eca		488b6c2440		MOVQ 0x40(SP), BP
  main.go:39		0x1051ecf		4883c448		ADDQ $0x48, SP
  main.go:39		0x1051ed3		c3			RET
  main.go:38		0x1051ed4		e8477cffff		CALL runtime.morestack_noctxt(SB)
  main.go:38		0x1051ed9		e972ffffff		JMP main.test(SB)
  :-1			0x1051ede		cc			INT $0x3
  :-1			0x1051edf		cc			INT $0x3
```

#### morestack_noctxt
runtime.morestack_notxt函数会插入到每个函数中，主要作用有两个:

* 当前stack扩容
* Goroutine的抢占调度 

```
rumtime/asm_amd64.s:

// morestack but not preserving ctxt.
TEXT runtime·morestack_noctxt(SB),NOSPLIT,$0
	MOVL	$0, DX                      // 清空DX低32位，DX寄存器用于保存函数上下文
	JMP	runtime·morestack(SB)       // 跳转到morestack函数执行

```

```
TEXT runtime·morestack(SB),NOSPLIT,$0-0
	// Cannot grow scheduler stack (m->g0).
	get_tls(CX)
	MOVQ	g(CX), BX                       			// 保存当前的g到BX
	MOVQ	g_m(BX), BX                     			// 保存m到BX
	MOVQ	m_g0(BX), SI                    			// 保存g0到SI
	CMPQ	g(CX), SI                       			// 如果当前g不处于g0,则跳转PC+3
	JNE	3(PC)                           			// PC+3
	CALL	runtime·badmorestackg0(SB)      			// g0不允许扩展
	CALL	runtime·abort(SB)

	// Cannot grow signal stack (m->gsignal).
	MOVQ	m_gsignal(BX), SI               			// gsignal用于处理信号量的栈
	CMPQ	g(CX), SI
	JNE	3(PC)
	CALL	runtime·badmorestackgsignal(SB) 			// gsignal栈不允许扩展
	CALL	runtime·abort(SB)

	// Called from f.                       			// 把调用morestack的函数标记为f
	// Set m->morebuf to f's caller.        			// 保存f's 调用者信息到m.morebuf中
	NOP	SP	// tell vet SP changed - stop checking offsets	// 禁止检查Go源码静态错误
	MOVQ	8(SP), AX	                    			// 8(SP)保存f的返回地址，即f's caller的PC
	MOVQ	AX, (m_morebuf+gobuf_pc)(BX)    			// 将f's caller的PC保存到 m.morebuf.pc
	LEAQ	16(SP), AX	                    			// 16(SP)保存f的SP寄存器; f.SP保存到AX寄存器
	MOVQ	AX, (m_morebuf+gobuf_sp)(BX)    			// f.sp ==> m.morebuf.sp  
	get_tls(CX)                             			// MOVQ TLS, r
	MOVQ	g(CX), SI                       
	MOVQ	SI, (m_morebuf+gobuf_g)(BX)     			// 当前g保存到m.morebuf.g

	// Set g->sched to context in f.
	MOVQ	0(SP), AX // f's PC             			// f's PC, morestack的frameSize为0，此时0(SP)为f的返回地址
	MOVQ	AX, (g_sched+gobuf_pc)(SI)      			// f.pc ==> g.sched.pc 
	MOVQ	SI, (g_sched+gobuf_g)(SI)       			// 当前g保存到g.sched.g
	
    LEAQ	8(SP), AX                       			// f's SP          
	MOVQ	AX, (g_sched+gobuf_sp)(SI)      			// f's SP ==> g.sched.sp
	MOVQ	BP, (g_sched+gobuf_bp)(SI)      			// f's BP ==> g.sched.bp
	MOVQ	DX, (g_sched+gobuf_ctxt)(SI)    			// f's 上下文信息 ==> g.sched.ctxt 

	// Call newstack on m->g0's stack.
	MOVQ	m_g0(BX), BX                    			// 获取G0
	MOVQ	BX, g(CX)                       			// 设置当前g为G0
	MOVQ	(g_sched+gobuf_sp)(BX), SP      			// 设置SP寄存器为g0.sched.sp
	CALL	runtime·newstack(SB)            			// 调用newstack,该方法不会返回
	CALL	runtime·abort(SB)	            			// crash if newstack returns
	RE
```

```

//go:nowritebarrierrec
func newstack() {
    	// 获取当前g,即g0
	thisg := getg()

    	...
    
    	// curg是触发了morestack的g,即不是g0
	gp := thisg.m.curg

    	...

	morebuf := thisg.m.morebuf    // f's caller
	thisg.m.morebuf.pc = 0
	thisg.m.morebuf.lr = 0
	thisg.m.morebuf.sp = 0
	thisg.m.morebuf.g = 0

	// NOTE: stackguard0 may change underfoot, if another thread
	// is about to try to preempt gp. Read it just once and use that same
	// value now and below.

    	// 检查是否需要抢占，当发现一个线程需要抢占时,会将其g.stackguard0标记为stackPreempt
	preempt := atomic.Loaduintptr(&gp.stackguard0) == stackPreempt

    	// 触发了抢占
	if preempt {

        	// 没有成功抢占, 等待下次抢占
		if thisg.m.locks != 0 || thisg.m.mallocing != 0 || thisg.m.preemptoff != "" || thisg.m.p.ptr().status != _Prunning {
			// Let the goroutine keep running for now.
			// gp->preempt is set, so it will  be preempted next time.
			gp.stackguard0 = gp.stack.lo + _StackGuard      // gp.stackguard0还原
			gogo(&gp.sched) // never return                      
		}
	}

	sp := gp.sched.sp

    	// 再次检查抢占
	if preempt {

		// Synchronize with scang.
        	// scang 必须在系统栈上运行，以防扫描自己stack
        	// 抢占扫描
		casgstatus(gp, _Grunning, _Gwaiting)                	// 状态更新为_Gwaiting
		if gp.preemptscan {
			for !castogscanstatus(gp, _Gwaiting, _Gscanwaiting) {
				// Likely to be racing with the GC as
				// it sees a _Gwaiting and does the
				// stack scan. If so, gcworkdone will
				// be set and gcphasework will simply
				// return.
			}
			if !gp.gcscandone {
				// gcw is safe because we're on the
				// system stack.
				gcw := &gp.m.p.ptr().gcw
				scanstack(gp, gcw)                	// 扫描gp的栈                          
				gp.gcscandone = true
			}
			gp.preemptscan = false
			gp.preempt = false
			casfrom_Gscanstatus(gp, _Gscanwaiting, _Gwaiting)
			
            		// This clears gcscanvalid.
			casgstatus(gp, _Gwaiting, _Grunning)
			gp.stackguard0 = gp.stack.lo + _StackGuard
			gogo(&gp.sched) // never return
		}

        	// runtime.Gosched, 让出当前cpu时间片,让其它线程执行
		casgstatus(gp, _Gwaiting, _Grunning)

        	// 抢占调度
        	// 将gp 压入P队列,进行调度
		gopreempt_m(gp) // never return
	}

    	// 如果不是由于抢占而执行morestack,而是stack扩容
    	// 新扩容栈是原来2倍,stack最大限制1<<20Byte
	oldsize := gp.stack.hi - gp.stack.lo
	newsize := oldsize * 2

    	// 设置G状态
	casgstatus(gp, _Grunning, _Gcopystack)

    	// copystack会创建一个新的stack, 然后把旧的栈内容拷贝到新的栈中
	copystack(gp, newsize, true)

    	// 开始执行
	casgstatus(gp, _Gcopystack, _Grunning)
	gogo(&gp.sched)
}
```


#### 解决方案

* 在for { x++ }循环内插入函数调用。
* 手工执行runtime.scheduler进行cpu抢占调度。
