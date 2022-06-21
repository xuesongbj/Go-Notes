# 匿名函数

## 大纲

* 匿名函数symbal。
* 匿名函数调用方式。
* 最为返回值的匿名函数.


## 匿名函数symbal

### 匿名函数示例
```
package main

func test() func(x int) int {
    return func(x int) int {
        x += 100
        return x
    }
}

func main() {
    f := test()
    z := f(100)
    println(z)
}
```

### symbal
查看以上示例代码有哪些符号表.

```
00000000004010d0 T main.init
00000000004b3720 B main.initdone.
0000000000401020 T main.main
0000000000401000 T main.test
0000000000401090 T main.test.func1      // 由编译器生成的随机符号名
000000000046ed78 R main.test.func1.f    // T .text R .readonly
000000000046ef38 R runtime.main.f
0000000000444410 T runtime.main.func1
000000000046ef20 R runtime.main.func1.f
0000000000444460 T runtime.main.func2
000000000046ef28 R runtime.main.func2.f
```

通过查看符号名称发现有main.test.func1 和main.test.func1.f符号名称。其中main.test.func1.f是test函数返回匿名函数值。它是一个包装对象,里面第一个元素存储的指针是就是main.test.func1匿名函数。

```
(gdb) info locals
z = 0
f = {void (int, int *)} 0xc420035f30                    // 这个看上去好像是test函数返回值
(gdb) x/1xg &f                                          // 查看局部变量f内容; x 查看变量内容 1 一组 x 十六进制显示 g 组
0xc420035f30:    0x000000000046ed78
(gdb) x/1xg 0x000000000046ed78                          // 变量f存储的是一个地址.即test函数返回的是一个指针,指向main.test.func1.f,该符号内存储的是一个地址.
0x46ed78 <main.test.func1.f>:    0x0000000000401090
(gdb) x/1xg 0x0000000000401090                          // 再次查看该地址内的内容
0x401090 <main.test.func1>:  0x246c894810ec8348         // 存储的内容就是符号表内的main.test.func1,即test返回的匿名函数.
```

当匿名函数作为一个返回值进行返回时,返回的是一个包装对象(main.test.func1.f)。包装对象通过二次寻址获取到匿名函数，然后进行调用匿名函数(main.test.func1)。


## 匿名函数直接调用
```
package main

func main() {
    func() {
        println("hello, world!")
    }()
}
```

当匿名函数不作为返回值进行返回时,再看看该匿名函数的符号表.

```
'' 0000000000401080 T main.init
'' 00000000004b3720 B main.initdone.
'' 0000000000401000 T main.main
'' 0000000000401020 T main.main.func1
'' 000000000046eeb0 R runtime.main.f
'' 00000000004443c0 T runtime.main.func1
'' 000000000046ee98 R runtime.main.func1.f
'' 0000000000444410 T runtime.main.func2
'' 000000000046eea0 R runtime.main.func2.f
```

匿名函数不作为返回值进行返回时,这个时候编译器只生成了一个符号main.main.func1.现在通过反汇编汇,看一下匿名函数的具体调用情况.

```
'' Dump of assembler code for function main.main:
''    0x0000000000401000 <+0>:	mov    %fs:0xfffffffffffffff8,%rcx
''    0x0000000000401009 <+9>:	cmp    0x10(%rcx),%rsp
''    0x000000000040100d <+13>:	jbe    0x401015 <main.main+21>
''    0x000000000040100f <+15>:	callq  0x401020 <main.main.func1>
'' => 0x0000000000401014 <+20>:	retq
''    0x0000000000401015 <+21>:	callq  0x449830 <runtime.morestack_noctxt>
''    0x000000000040101a <+26>:	jmp    0x401000 <main.main>
```

通过反汇编得知,当匿名函数直接调用时，他和普通函数调用没有任何区别。