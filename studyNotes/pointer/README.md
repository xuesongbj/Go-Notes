# 指针

### 指针定义

* 指针是一个实体(变量,需要分配内存空间)，指针字节长度是固定的。因为指针保存的是地址，由操作系统的位数决定长度，32位机器字节长度为4；64位字节机器字节长度为8。
* 指针内保存的是目标地址，指向真实存在的内存空间，称之为有效指针；如果指针内保存的是空地址(Null)，称之为空指针。
* 指针是一种高级语言的特性；在汇编中没有指针概念,就是一个uint64整数。


## 指针实例

### 指针内容

```
func test() {
	var i int = 100    // 0x64
	var p *int = &i
	
	println(p, *p)
}

func main() {
	test()
}
```

* 输出结果

```
0xc000030760 100
```

* 反汇编

```
$> go tool objdump -s "main\.test" test

  test.go:6		0x104e15d		48c744240864000000	MOVQ $0x64, 0x8(SP)					// var i int = 100，字面量100压入stack
  test.go:9		0x104e166		e87545fdff		CALL runtime.printlock(SB)
  test.go:9		0x104e16b		488d442408		LEAQ 0x8(SP), AX						// 将变量i指针压入AX寄存器
  test.go:9		0x104e170		48890424		MOVQ AX, 0(SP)
  test.go:9		0x104e174		e8474efdff		CALL runtime.printpointer(SB)
```

### 指针类型转换
Go语言可以通过unsafe进行指针类型转换，再进行转换前需要确定指针类型和转换后目标类型要匹配。

```
import (
	"fmt"
)

//go:noinline
//go:nosplit
func test() {
	var i int = 0x01020304
	
	var p *[8]byte                           // ptr1 --> unsafe.Pointer --> ptr2
	p = (*[8]byte)(unsafe.Pointer(&i))       // *int ---> *[8]byte
	
	p[0] = 5
	p[6] = 6
	
	fmt.Println("%v, %016X\n", p, i)
}
```

* 输出:

```
&[5 3 2 1 0 6 0 0] 6597086675717   => 6597086675717 16进制 == 0x060001020305
```
以上实例,将*int指针类型转换为\*[8]byte类型。内存字节数组按照小端排序,故0x01020304转换成字节数组格式为0x04030201。

* GDB调试

```
(gdb) disassemble
0x000000000048c9db <+59>:	mov    QWORD PTR [rax],0x1020304    // var i int = 0x01020304
0x000000000048c9e2 <+66>:	mov    BYTE PTR [rax],0x5           // p[0] = 5
0x000000000048c9e5 <+69>:	mov    BYTE PTR [rax+0x5],0x6       // p[5] = 6

// *int转换成字节数组后结果
// p[0] ==> 0x05
// p[5] ==> 0x06
(gdb) x/8xb $rax
0xc000084010:	0x05	0x03	0x02	0x01	0x00	0x06	0x00	0x00
```

### 指针运算
Go语言指针不支持显式的进行指针运算,但可以通过unitptr进行指针运算，然后再转成指针类型，以实现指针运算。

```
//go:noinline
//go:split
func test() {
	x := [8]byte{1,2,3,4}
	var p *byte = &x[0]
		
	u := uintptr(unsafe.Pointer(p))    // ptr -> unsafe.Pointer -> uintptr
	fmt.Println("%p, %x\n", p, u)
		
	u++   								// uintptr++
	fmt.Printf("%p, %x\n", p, u)
	
	p2 := (*byte)(unsafe.Pointer(u))	// uintptr -> unsafe.Pointer -> ptr
	*p2 += 100
	fmt.Println(x)
}
```

* 输出

```
0xc000014090, c000014090      // p, u
[1 102 3 4 0 0 0 0]           // x
```

通过以上输出结果可以看出,将x[0]地址转换成uintptr类型,然后进行++运算,然后再将结果转换成指针类型,完成指针运算。通过输出字节数组可以看到,第二个元素的值被修改。接下来通过汇编查看一下具体过程:

```
(gdb) b main.test
(gdb) r
(gdb) disassemble
0x000000000048e71c <+76>:	mov    rcx,QWORD PTR [rip+0x4c89d]
0x000000000048e723 <+83>:	mov    QWORD PTR [rax],rcx         // x := [8]byte{1, 2, 3, 4}

// x 变量底层字节数组
(gdb) x/4xb $rax
0xc000084010:	0x01	0x02	0x03	0x04

// u := uintptr(unsafe.Pointer(p))
// 将指针类型转换成uintptr类型
0x000000000048e726 <+86>:	mov    rax,QWORD PTR [rsp+0x80]
0x000000000048e733 <+99>:	mov    QWORD PTR [rsp+0x40],rax
0x000000000048e742 <+114>:	mov    rax,QWORD PTR [rsp+0x40]
0x000000000048e747 <+119>:	mov    QWORD PTR [rsp],rax
0x000000000048e74b <+123>:	call   0x408e50 <runtime.convT64>
0x000000000048e750 <+128>:	mov    rax,QWORD PTR [rsp+0x8]

// u++
0x000000000048e66b <+219>:	mov    rax,QWORD PTR [rsp+0x50]
0x000000000048e670 <+224>:	inc    rax

// p2 := (*byte)(unsafe.Pointer(u))
// *p2 += 100
0x000000000048e673 <+227>:	movzx  ecx,BYTE PTR [rax]
0x000000000048e676 <+230>:	add    ecx,0x64
0x000000000048e679 <+233>:	mov    BYTE PTR [rax],cl
 
(gdb) info locals
u = 824633802904
&x = 0xc000014098
p2 = 0xc000014099 "f\003\004"

// 通过将指针转换成uintptr实现指针运算,然后再转换回指针类型
// 将数组第二个元素2,进行了加100操作，即0x66
(gdb) x/4wb 824633802904			// u
0xc000014098:	0x01	0x66	0x03	0x04

(gdb) x/4wb 0xc000014098  			// x
0xc000014098:	0x01	0x66	0x03	0x04 
```

### 利用指针提升性能
Go语言在进行类型转换时候，标准库提供的一般会发生内存拷贝操作。这样效率比较低,可以使用指针提升性能。

#### 字符串转换

* string函数

标准库通过string函数,将字节slice转换成字符串类型。会发生值的内存拷贝操作。

```
//go:noinline
//go:nosplit
func test() {
	d := []byte{'a', 'b', 'c', 'd'}
	s := string(d)   // value copy
	fmt.Println(d, s)
}
```

反汇编查看具体操作:

```
// d := []byte{'a', 'b', 'c', 'd'}
// runtime.newobject函数创建slice赋值给变量d,栈上$rsp+0x8栈帧空间，然后搬到寄存器RAX
0x000000000048c9b7 <+23>:	lea    rax,[rip+0x15022]        # 0x4a19e0
0x000000000048c9be <+30>:	mov    QWORD PTR [rsp],rax
0x000000000048c9c2 <+34>:	call   0x40b5a0 <runtime.newobject>
0x000000000048c9c7 <+39>:	mov    rax,QWORD PTR [rsp+0x8]
0x000000000048c9cc <+44>:	mov    QWORD PTR [rsp+0x58],rax
0x000000000048c9d1 <+49>:	mov    ecx,DWORD PTR [rip+0x4c0a5]        # 0x4d8a7c
0x000000000048c9d7 <+55>:	mov    DWORD PTR [rax],ecx

// s := string(d)
// string函数将字节slice转成字符串类型,返回结果存储在$rsp+0x20栈帧空间
// 发生值考本
0x000000000048c9d9 <+57>:	mov    QWORD PTR [rsp],0x0
0x000000000048c9e1 <+65>:	mov    QWORD PTR [rsp+0x8],rax      // ptr
0x000000000048c9e6 <+70>:	mov    QWORD PTR [rsp+0x10],0x4     // len
0x000000000048c9ef <+79>:	mov    QWORD PTR [rsp+0x18],0x4     // cap
0x000000000048c9f8 <+88>:	call   0x440790 <runtime.slicebytetostring>   // string()
0x000000000048c9fd <+93>:	mov    rax,QWORD PTR [rsp+0x20]     // return ==> $rsp+0x20
0x000000000048ca02 <+98>:	mov    QWORD PTR [rsp+0x48],rax
0x000000000048ca07 <+103>:	mov    rcx,QWORD PTR [rsp+0x28]
0x000000000048ca0c <+108>:	mov    QWORD PTR [rsp+0x40],rcx
```

* 通过unsafe 方式进行类型转换

```
package main

import (
        "fmt"
        "unsafe"
)

//go:noinline
//go:nosplit
func test() {
	// string header
    type str struct {
        data uintptr
        len  int
    }

	// slice header
    type sli struct {
        data uintptr
        len  int
        cap  int
    }

    d := []byte{'a', 'b', 'c', 'd'}
    sh := *(*str)(unsafe.Pointer(&d))

    fmt.Printf("%#v\n", sh)
}

func main() {
        test()
}
```

反汇编，剖析实现原理:

```
// d := []byte{'a', 'b', 'c', 'd'}
0x000000000048e2b7 <+23>:	mov    eax,DWORD PTR [rip+0x4cb1f]        # 0x4daddc
0x000000000048e2bd <+29>:	mov    DWORD PTR [rsp+0x54],eax
0x000000000048e2c1 <+33>:	lea    rax,[rsp+0x54]
0x000000000048e2c6 <+38>:	mov    QWORD PTR [rsp+0x78],rax		// slice header ptr -> array  
0x000000000048e2cb <+43>:	mov    QWORD PTR [rsp+0x80],0x4		// slice header len -> len(d)           
0x000000000048e2d7 <+55>:	mov    QWORD PTR [rsp+0x88],0x4		// slice header cap -> cap(d)

// sh := *(*str)(unsafe.Pointer(&d))
// 通过结果可以看出,sh字符串变量通过指针的方式指向了$rsp+0x78位置，该位置即d slice底层数组起始指针。
// 通过这种方式就无需进行内存拷贝,直接通过指针实现类型转换
// 性能肯定要比string函数高
0x000000000048e2e3 <+67>:	mov    rax,QWORD PTR [rsp+0x78]

(gdb) p/x $rsp+0x78
$2 = 0xc000068f28         // 指向底层数组

// 0x61 ASCII即'a'
// 0x62 ASCII即'b'
// 0x63 ASCII即'c'
// 0x64 ASCII即'd'
(gdb) x/4xb 0xc000068f28  // 查看底层数组内容
0xc000068f04:	0x61	0x62	0x63	0x64
```
### 备注

* 内存字节序

		计算机硬件有两种储存数据的方式：大端字节序(big endian)和小端字节序(little endian)。
	
		举例来说, 数值0x2211使用两个字节储存：高位字节是0x22，低位字节是0x11。
		
		大端: 高位字节在前，低位字节在后，这是人类读写数值的方法。
		小端: 低位字节在前，高位字节在后，即以0x1122形式储存。
