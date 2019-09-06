# Go1.13 defer
go 1.13版本对defer进行了优化，对符号不发生逃逸的defer函数通过deferprocStack定义一个延迟调用对象，放入队列，等待被调用。deferprocStack函数是Go1.13版本新加入函数，该函数不会发生逃逸。

### 实例
通过Go1.12版本和Go1.13版本编译如下源代码，确定

#### 源代码
```
package main

func main() {
    b := 0
    defer func(){
        a := make([]int, 128)
        b = b + len(a)
    }()
}
```

#### Go1.12版本编译
```
➜  src git:(master) ✗ go build -gcflags "-l -m -N" -o m12 main.go
# command-line-arguments
./main.go:19:11: main func literal does not escape
./main.go:20:18: main.func1 make([]int, 128) does not escape
```

#### Go1.13版本编译
```
➜  src git:(master) ✗ ~/data/go-go1.13/bin/go build -gcflags "-l -m -N" -o m13 main.go
# command-line-arguments
./main.go:19:11: main func literal does not escape
./main.go:20:18: main.func1 make([]int, 128) does not escape
```

通过Go1.12和Go1.13两个版本编译结果可以看出，函数defer函数 main.main.func1没有发生逃逸。接下来查看反汇编有什么差异。

#### Go1.12 反汇编

通过反汇编可以看出，Go1.12版本defer通过deferproc定义一个延迟调用对象。

```
➜  src git:(master) ✗ go tool objdump m12 | grep -A 40 "main.main"
TEXT main.main(SB) /Users/zhangshaozhi/data/go/go.test/src/main.go
  main.go:7		0x104ea20		65488b0c2530000000	MOVQ GS:0x30, CX
  main.go:7		0x104ea29		483b6110		CMPQ 0x10(CX), SP
  main.go:7		0x104ea2d		765f			JBE 0x104ea8e
  main.go:7		0x104ea2f		4883ec28		SUBQ $0x28, SP
  main.go:7		0x104ea33		48896c2420		MOVQ BP, 0x20(SP)
  main.go:7		0x104ea38		488d6c2420		LEAQ 0x20(SP), BP
  main.go:18		0x104ea3d		48c744241800000000	MOVQ $0x0, 0x18(SP)
  main.go:19		0x104ea46		c7042408000000		MOVL $0x8, 0(SP)
  main.go:19		0x104ea4d		488d050c2b0200		LEAQ go.func.*+61(SB), AX
  main.go:19		0x104ea54		4889442408		MOVQ AX, 0x8(SP)
  main.go:19		0x104ea59		488d442418		LEAQ 0x18(SP), AX
  main.go:19		0x104ea5e		4889442410		MOVQ AX, 0x10(SP)
  main.go:19		0x104ea63		e8b831fdff		CALL runtime.deferproc(SB)
  main.go:19		0x104ea68		85c0			TESTL AX, AX
  main.go:19		0x104ea6a		7512			JNE 0x104ea7e
  main.go:19		0x104ea6c		eb00			JMP 0x104ea6e
  main.go:23		0x104ea6e		90			NOPL
  main.go:23		0x104ea6f		e83c3afdff		CALL runtime.deferreturn(SB)
  main.go:23		0x104ea74		488b6c2420		MOVQ 0x20(SP), BP
  main.go:23		0x104ea79		4883c428		ADDQ $0x28, SP
  main.go:23		0x104ea7d		c3			RET
  main.go:19		0x104ea7e		90			NOPL
  main.go:19		0x104ea7f		e82c3afdff		CALL runtime.deferreturn(SB)
  main.go:19		0x104ea84		488b6c2420		MOVQ 0x20(SP), BP
  main.go:19		0x104ea89		4883c428		ADDQ $0x28, SP
  main.go:19		0x104ea8d		c3			RET
  main.go:7		0x104ea8e		e8cd84ffff		CALL runtime.morestack_noctxt(SB)
  main.go:7		0x104ea93		eb8b			JMP main.main(SB)
  :-1			0x104ea95		cc			INT $0x3
  :-1			0x104ea96		cc			INT $0x3
  :-1			0x104ea97		cc			INT $0x3
  :-1			0x104ea98		cc			INT $0x3
  :-1			0x104ea99		cc			INT $0x3
  :-1			0x104ea9a		cc			INT $0x3
  :-1			0x104ea9b		cc			INT $0x3
  :-1			0x104ea9c		cc			INT $0x3
  :-1			0x104ea9d		cc			INT $0x3
  :-1			0x104ea9e		cc			INT $0x3
  :-1			0x104ea9f		cc			INT $0x3
```

#### Go1.13 反汇编

通过反汇编可以看出，Go1.13版本defer通过deferprocStack定义一个延迟调用对象。

```
➜  src git:(master) ✗ go tool objdump m13 | grep -A 40 "main.main"
TEXT main.main(SB) /Users/zhangshaozhi/data/go/go.test/src/main.go
  main.go:7		0x10512e0		65488b0c2530000000	MOVQ GS:0x30, CX
  main.go:7		0x10512e9		483b6110		CMPQ 0x10(CX), SP
  main.go:7		0x10512ed		7669			JBE 0x1051358
  main.go:7		0x10512ef		4883ec50		SUBQ $0x50, SP
  main.go:7		0x10512f3		48896c2448		MOVQ BP, 0x48(SP)
  main.go:7		0x10512f8		488d6c2448		LEAQ 0x48(SP), BP
  main.go:18		0x10512fd		48c744240800000000	MOVQ $0x0, 0x8(SP)
  main.go:19		0x1051306		c744241008000000	MOVL $0x8, 0x10(SP)
  main.go:19		0x105130e		488d059b420200		LEAQ go.func.*+66(SB), AX
  main.go:19		0x1051315		4889442428		MOVQ AX, 0x28(SP)
  main.go:19		0x105131a		488d442408		LEAQ 0x8(SP), AX
  main.go:19		0x105131f		4889442440		MOVQ AX, 0x40(SP)
  main.go:19		0x1051324		488d442410		LEAQ 0x10(SP), AX
  main.go:19		0x1051329		48890424		MOVQ AX, 0(SP)
  main.go:19		0x105132d		e83e20fdff		CALL runtime.deferprocStack(SB)
  main.go:19		0x1051332		85c0			TESTL AX, AX
  main.go:19		0x1051334		7512			JNE 0x1051348
  main.go:19		0x1051336		eb00			JMP 0x1051338
  main.go:23		0x1051338		90			NOPL
  main.go:23		0x1051339		e83226fdff		CALL runtime.deferreturn(SB)
  main.go:23		0x105133e		488b6c2448		MOVQ 0x48(SP), BP
  main.go:23		0x1051343		4883c450		ADDQ $0x50, SP
  main.go:23		0x1051347		c3			RET
  main.go:19		0x1051348		90			NOPL
  main.go:19		0x1051349		e82226fdff		CALL runtime.deferreturn(SB)
  main.go:19		0x105134e		488b6c2448		MOVQ 0x48(SP), BP
  main.go:19		0x1051353		4883c450		ADDQ $0x50, SP
  main.go:19		0x1051357		c3			RET
  main.go:7		0x1051358		e83381ffff		CALL runtime.morestack_noctxt(SB)
  main.go:7		0x105135d		eb81			JMP main.main(SB)
  :-1			0x105135f		cc			INT $0x3
```

通过反汇编结果可以看出，go1.13对defer 创建延迟函数进行了优化。



## defer源码剖析

### SSA对Defer处理

```
// src/cmd/compile/internal/gc/ssa.go

func (s *state) stmt(n *Node) {
	// ...
	
	switch n.Op {
	case ODEFER:
		d := callDefer
		
		// 如果defer 延迟函数没有发生逃逸，则内存分配在Stack上
		if n.Esc == EscNever {
			d = callDeferStack
		}
		s.call(n.Left, d）
	}
	
	// ...
}
```

* state.call 函数调用

```
func (s *state) call(n *Node, k callKind) *ssa.Value {
	if k == callDeferStack {
		// ....
		
		// 使用指向_defer记录的指针调用deferprocStack
		arg0 := s.constOffPtrSP(types.Types[TUINTPTR], Ctxt.FixedFrameSize())
		s.store(types.Types[TUINTPTR], arg0, addr)
		call = s.newValue1A(ssa.OpStaticCall, types.TypeMem, deferprocStack, s.mem())
		
		// ...
	}
}
```

### deferprocStack

deerprocStack函数主要将_defer记录压入到stack上,而非heap上。等待函数退出时,从队列拿出defer延迟函数进行调用。

```
// 1. 将defer记录压入队列，该队列内存分配在stack上
// 2. defer记录的size和fn必须被初始化
func deferprocStack(d *_defer) {
	gp := getg()

	d.started = false
	d.heap = false
	d.sp = getcallersp()
	d.pc = getcallerpc()

	// 下面代码实现如下功能:
	//   d.panic = nil
	//   d.link = gp._defer
	//   gp._defer = d
	*(*uintptr)(unsafe.Pointer(&d._panic)) = 0
	*(*uintptr)(unsafe.Pointer(&d.link)) = uintptr(unsafe.Pointer(gp._defer))
	*(*uintptr)(unsafe.Pointer(&gp._defer)) = uintptr(unsafe.Pointer(d))

	return0()
	// No code can go here - the C return register has
	// been set and must not be clobbered.
}
```
