# 闭包

## 定义
一个匿名函数引用它的上下文对象,把这种状态称为闭包。闭包由两部分组成：

* 函数
* 环境变量

当一个函数有局部变量x,当这个函数返回一个闭包,闭包通过指针引用环境变量,使环境变量生命周期延长,此时变量x会逃逸到heap上。只有这样调用闭包函数时,才能访问到x的值。

### Go使用闭包
```
package main

func test(x int) func() {
	println("test.x :", &x)
	
	return func() {
		println("closure.x :", &x, x)	
	}
}
 
func main() {
	f := test(100)
	f()
}
```

### 闭包实现分析
查看源代码编译之后的elf文件符号表、gdb调试及调用堆栈分析等手段查看Go 闭包是如何实现的。

* 编译 & 查看符号表

```
$> go build -gcflags "-l -m -N" -o test main.go

$> nm test | grep "main\."
000000000104e2f0 t main.init
00000000010d5a82 s main.initdone.
000000000104e220 t main.main
000000000104e140 t main.test
000000000104e270 t main.test.func1
0000000001045b30 t runtime.main.func1
0000000001045b80 t runtime.main.func2
```

* gdb 调试

```
Dump of assembler code for function main.main:
   0x000000000104e220 <+0>:	mov    rcx,QWORD PTR gs:0x30
   0x000000000104e229 <+9>:	cmp    rsp,QWORD PTR [rcx+0x10]
   0x000000000104e22d <+13>:	jbe    0x104e263 <main.main+67>
   0x000000000104e22f <+15>:	sub    rsp,0x20
   0x000000000104e233 <+19>:	mov    QWORD PTR [rsp+0x18],rbp
   0x000000000104e238 <+24>:	lea    rbp,[rsp+0x18]
   
   ; [rsp]通过寄存器传参
=> 0x000000000104e23d <+29>:	mov    QWORD PTR [rsp],0x64
   0x000000000104e245 <+37>:	call   0x104e140 <main.test>
   
   ; rdx是test函数返回值
   0x000000000104e24a <+42>:	mov    rdx,QWORD PTR [rsp+0x8]
   0x000000000104e24f <+47>:	mov    QWORD PTR [rsp+0x10],rdx
   0x000000000104e254 <+52>:	mov    rax,QWORD PTR [rdx]
   0x000000000104e257 <+55>:	call   rax
   0x000000000104e259 <+57>:	mov    rbp,QWORD PTR [rsp+0x18]
   0x000000000104e25e <+62>:	add    rsp,0x20
   0x000000000104e262 <+66>:	ret
   0x000000000104e263 <+67>:	call   0x1046ae0 <runtime.morestack_noctxt>
   0x000000000104e268 <+72>:	jmp    0x104e220 <main.main>
   

; rdx是一个复合类型
; rdx[0]: main.test.func1函数,test函数内返回函数的内存地址
; rdx[1]: test函数环境变量
gdb) p/x $rdx
$4 = 0xc00000e020
(gdb) x/2xg  0xc00000e020
0xc00000e020:	0x000000000104e270	0x000000c0000160d0

(gdb) x/xg 0x000000000104e270
0x104e270 <main.test.func1>:	0x000030250c8b4865

(gdb) x/xg 0x000000c0000160d0
0xc0000160d0:	0x0000000000000064
```
通过分析,可以得知test函数返回的是一个复合结构体,由两个元素组成.第一个保存闭包函数地址,第二个是环境变量地址(环境变量会逃逸到heap上)

```
// 逃逸分析
$> go build -gcflags "-l -m -N" -o test main.go
# command-line-arguments
./closure.go:6:12: func literal escapes to heap
./closure.go:6:12: func literal escapes to heap

# 环境变量逃逸到heap
./closure.go:7:18: &x escapes to heap
./closure.go:3:11: moved to heap: x
./closure.go:4:13: test &x does not escape
./closure.go:7:17: test.func1 &x does not escape
```


### 使用闭包注意事项
当函数返回多个闭包函数时,如果闭包之间都引用了相同的环境变量,此时会有数据竞争的问题,在闭包调用时需要用锁进行处理处理。

```
package main

func test(x int) (func(), func()) {

    w := func() {
        for {
            x++
        }
    }

    r := func() {
        for {
            _ = x
        }
    }

    return w, r
}

func main() {
    w, r := test(1000)

    go w()
    go r()
}
```

* 数据竞争检查

```
$> go run -race closure.go
==================
WARNING: DATA RACE
Read at 0x00c00007e000 by goroutine 6:
  main.test.func2()
      /Users/David/data/go/go.test/src/Demo/closure.go:13 +0x46

Previous write at 0x00c00007e000 by goroutine 5:
  main.test.func1()
      /Users/David/data/go/go.test/src/Demo/closure.go:7 +0x5c

Goroutine 6 (running) created at:
  main.main()
      /Users/David/data/go/go.test/src/Demo/closure.go:24 +0x63

Goroutine 5 (running) created at:
  main.main()
      /Users/David/data/go/go.test/src/Demo/closure.go:23 +0x4d
==================
Found 1 data race(s)
exit status 66
```

