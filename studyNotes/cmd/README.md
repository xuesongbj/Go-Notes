# Go命令行工具

## 编译

Go语言通过go build命令对源代码进行编译，go 工具链提供了丰富的编译调试工具。

* -a 

	强制编译;即使程序没有更新，重新执行build时也会进行编译。
	
* -n

	输出编译和链接过程，但不会进行真实编译。
	
* -p n

	编译过程中各任务并发数量。默认情况下，该数量等于cpu的逻辑核数。
	
* -race

	启用数据竞争检查。
		
```
$> go run -race test.go

源代码:
package main

import (
        "time"
        "fmt"
)

func main() {
        a := 1
        go func(){
                a = 2
        }()

        a = 3

        fmt.Println("a is ", a)

        time.Sleep(2 * time.Second)
}

输出:
==================
WARNING: DATA RACE
Write at 0x00c000016068 by goroutine 7:
  main.main.func1()
      /home/work/go/go.test/src/test.go:11 +0x38

Previous write at 0x00c000016068 by main goroutine:
  main.main()
      /home/work/go/go.test/src/test.go:14 +0x88

Goroutine 7 (running) created at:
  main.main()
      /home/work/go/go.test/src/test.go:10 +0x7a
==================
```

* -v

	进行编译时，打印包名称。
	
* -msan

	MemoryScanitizer(MSan)是程序中未初始化的内存读取检查器。仅支持在Linux/amd64,Linux/arm64 并且仅支持Clang/LLVM作为C编译器。
	
* -work

	输出并保存编译时的临时文件，不会删除。
	
```
$> go build -work
WORK=/tmp/go-build932351004
```

* -x

	打印编译过程
	
* -asmflags

	用于控制Go语言编译器编译汇编语言文件时行为，具体参数含义如下:
	
	[详情参考官网](https://golang.org/doc/asm)
	
```
-D 预定义符号，-D=identifier=value
-I 从指定目录导入
-S 输出汇编和机器码对照关系
-V 输出版本并退出
-o 输出二进制可文件(可重定向文件)
-shared 生成动态链接文件
-trimpath 从源文件记录文件删除前缀
```

* buildmode mode

	编译模式,目前Go编译器支持8种模式，具体如下：
	
```
1. archive    静态编译库，将非main包以外的所有包代码构建成归档(archive或.a)文件.
2. c-archive  静态链接库，这里构建的是供C程序调用的库。将Go程序构建为archive(.a)文件，这里C类的程序可以静态链接.a文件，并调用其中代码。

import "C"
// export Hello   // 需要导出给C调用的函数，必须通过注释添加这个构建信息，否则不会生成C所需的头文件。

3. c-shared   动态链接库，Go代码创建一个动态链接库(unix: .so, windows: .dll),然后用C语言程序动态加载运行。

-o hello.so  // 明确指定动态链接文件

4. default     默认编译模式，将main和所有import包编译成二进制文件；然后将非main以外的包编译成archive文件。

5. shared      动态链接库，将所有非main包编译并合并到一个共享库中，该模式将在-linkshared选项构建时使用。main package将被忽略

6. exe         构建main和所有import包的二进制文件
7. pie         这是构建运行地址无关的二进制可执行文件格式，这是一种安全特性，可以在支持PIE的操作系统中，让可执行文件加载时，每次的地址都是不同的避免已知地址的跳跃式的攻击。
8. plugin      插件形式和c-shared、shared相似,都是构建一个动态链接库。不同点是，动态链接库并非在启动时加载，而是由程序决定何时加载和释放。

package main
import "plugin"
func main() {
	//	加载 myplugin 库
	p, err := plugin.Open("myplugin.so")
	if err != nil {
		log.Fatal(err)
	}
	//	取得 Hello 函数
	fn, err := p.Lookup("Hello")
	if err != nil {
		log.Fatal(err)
	}
	//	调用函数
	fn.(func())()
}
```

* -gccgoflgs

	gccgo一种Go语言编译器。gccgo编译器是gcc新的前端，GCC是使用GNU编译器。
	
	[详情参考go官网](https://golang.org/doc/install/gccgo)
	
* -gcflags

	编译时传递给编译器的参数，具体可以通过go tool compile查看。
	
```
-E : 输出符号表信息
-% : 非静态初始化debug信息
-N : 关闭优化
-C : 关闭错误信息中列信息输出
-L : 错误信息中输出完整文件名称
-D path : 设置本地相对路径导入
-I directory : 增加imort扫描路径
-K : debug信息中不输出行号
-S : 输出编译后生成的汇编代码(go tool compile -s 或 go build -gcflags "-S")
-V : 输出编译当前版本信息
-W : 输出debug解析树
-m : 输出优化信息
-o : 编译后指定输出名字，默认当前包名称
-race : 开启内存竞争检查
-wb : 开启写屏障(默认开启)
-l : 关闭内链
-dwarf : 生成DWARF符号表，默认输出格式。之前采用stabs格式
-dwarfbasentries : 在DWARF中使用基地址条目，默认开启
-newescape : 启用新的内存逃逸分析，默认开启
-nolocalimports : 决绝package相对路径导入
-p path : 设置package导入路径
-r : debug装饰器
-symabis file : 从文件中读取符号表ABIs信息
-traceprofile file : 将程序执行的trace信息写入文件
-v : 显示调试详细信息
-w : debug类型检查
-live : debug活跃度分析
-shared : 生成动态链接库,不包含main package
-smallframes : 设置stack分配最大变量大小,默认10MB
-memprofilerate rate : 设置内存分析器对内存对象抽样率,默认的采样率是是每512KB的内存分配一个样本。通过修改runtime.MemProfileRate变量进行修改。
-memprofile file : 程序内存分析pprof写入到指定文件
-mutexprofile file : 程序mutex分析结果写入指定文件
-asmhdr file : 将汇编头信息写入文件
-bench file: 将benchmark时间追加到文件
-blockprofile file: 将block分析信息写入文件
-buildid id : 在导出元数据时,将构建id作为记录id
-c int : 编译时的并发数
-cpuprofile file: cpu分析结果写入文件
```

```
$> go build -gcflags=-S fmt        // 打印fmt反汇编代码
$> go build -gcflags=all=-S fmt    // 打印fmt及所依赖的反汇编代码
```

* -ldflags

	程序编译完之后,需要通过链接器将多个目标文件合并成一个可执行文件(ELF)。go进行编译和链接时可以对链接器传递参数。具体参数如下

```
-B note: 添加ELF NT_GNU_BUILD_ID

$> go build -ldflags "-B 0x1234" test
$> readelf  -n test(ELF文件)

Displaying notes found in: .note.gnu.build-id
  Owner                 Data size	Description
  GNU                  0x00000002	NT_GNU_BUILD_ID (unique build ID bitstring)
    Build ID: 1234
    
或者直接查询.note.gnu.build-id section数据:

$> readelf -x .note.gnu.build-id test(ELF文件)
Hex dump of section '.note.gnu.build-id':
  0x00400fec 04000000 02000000 03000000 474e5500 ............GNU.
  0x00400ffc 12340000
```

```
-E entry: 设置符号名称
-H type: 设置Header类型
-I linker: 使用linker作为ELF linker
-L directory: 将指定目录添加到库文件路径
-T address: 设置.text端地址
-n: 导出符号表
-s: 关闭符号表
-u: 拒绝使用unsafe package
-v: 打印连接器trace信息
-w: 禁用DWARF
```

* -linkshared

	链接以前使用-buildmode=shared创建的动态链接
	
* -mod

	go mod对模块管理
	
```
download: 下载模块到本地缓存,缓存路径是$GOPATH/pkg/mod/cache
vendor: 把依赖拷贝到vendor/目录下
graph: 把模块之间的依赖图显示出来
init: 当前目录初始化新的模块
edit: 提供了命令版修改go.mod文件
tidy: 对go.mod依赖的package版本进行更新
verify: 确认依赖关系
why: 解释为什么需要包和模块
```

* -pkgdir dir

	从指定目录安装并加载包,而不从通常的位置安装
	
* -trimpath

	从生成的二进制文件中删除所有文件系统路径。记录的文件名称以go、path@version或者GoPATH开头，而不是绝对文件系统路径。
	
* -toolexec 'cmd args'

	用于调用vet和asm等工具链程序的程序。
	

## go tool

### objdump

可以通过go tool objdump工具对Go编译后的二进制文件进行反汇编。

* go tool objdump -S elf_file
   
```
// dump elf_file文件所有汇编代码
TEXT main.main.func1(SB) /home/work/go/go.test/src/test.go
  test.go:11		0x48cd70		488b442408		MOVQ 0x8(SP), AX
  test.go:11		0x48cd75		48c70002000000		MOVQ $0x2, 0(AX)
  test.go:12		0x48cd7c		c3			RET
```

* go tool objdump -s main.main elf_file

```
// 仅dump main.main函数汇编代码
TEXT main.main(SB) /home/work/go/go.test/src/test.go
  test.go:8		0x48cc60		64488b0c25f8ffffff	MOVQ FS:0xfffffff8, CX
  test.go:8		0x48cc69		483b6110		CMPQ 0x10(CX), SP
  test.go:8		0x48cc6d		0f86ef000000		JBE 0x48cd62
  test.go:8		0x48cc73		4883ec70		SUBQ $0x70, SP
  test.go:8		0x48cc77		48896c2468		MOVQ BP, 0x68(SP)
  test.go:8		0x48cc7c		488d6c2468		LEAQ 0x68(SP), BP
  test.go:9		0x48cc81		488d05d80b0100		LEAQ 0x10bd8(IP), AX
  test.go:9		0x48cc88		48890424		MOVQ AX, 0(SP)
  test.go:9		0x48cc8c		e80fe9f7ff		CALL runtime.newobject(SB)
  test.go:9		0x48cc91		488b442408		MOVQ 0x8(SP), AX
  test.go:9		0x48cc96		4889442440		MOVQ AX, 0x40(SP)
  test.go:9		0x48cc9b		48c70001000000		MOVQ $0x1, 0(AX)
  test.go:10		0x48cca2		c7042408000000		MOVL $0x8, 0(SP)
  test.go:10		0x48cca9		488d0db0ba0300		LEAQ 0x3bab0(IP), CX
  test.go:10		0x48ccb0		48894c2408		MOVQ CX, 0x8(SP)
  test.go:10		0x48ccb5		4889442410		MOVQ AX, 0x10(SP)
  test.go:10		0x48ccba		e88157faff		CALL runtime.newproc(SB)
  test.go:14		0x48ccbf		488b442440		MOVQ 0x40(SP), AX
  test.go:14		0x48ccc4		48c70003000000		MOVQ $0x3, 0(AX)
  test.go:16		0x48cccb		48c7042403000000	MOVQ $0x3, 0(SP)
  test.go:16		0x48ccd3		e878c1f7ff		CALL runtime.convT64(SB)
  test.go:16		0x48ccd8		488b442408		MOVQ 0x8(SP), AX
  test.go:16		0x48ccdd		0f57c0			XORPS X0, X0
  test.go:16		0x48cce0		0f11442448		MOVUPS X0, 0x48(SP)
  test.go:16		0x48cce5		0f11442458		MOVUPS X0, 0x58(SP)
  test.go:16		0x48ccea		488d0def140100		LEAQ 0x114ef(IP), CX
  test.go:16		0x48ccf1		48894c2448		MOVQ CX, 0x48(SP)
  test.go:16		0x48ccf6		488d0da3c20400		LEAQ 0x4c2a3(IP), CX
  test.go:16		0x48ccfd		48894c2450		MOVQ CX, 0x50(SP)
  test.go:16		0x48cd02		488d0d570b0100		LEAQ 0x10b57(IP), CX
  test.go:16		0x48cd09		48894c2458		MOVQ CX, 0x58(SP)
  test.go:16		0x48cd0e		4889442460		MOVQ AX, 0x60(SP)
  print.go:274		0x48cd13		488b05b6140d00		MOVQ os.Stdout(SB), AX
  print.go:274		0x48cd1a		488d0d5fd80400		LEAQ go.itab.*os.File,io.Writer(SB), CX
  print.go:274		0x48cd21		48890c24		MOVQ CX, 0(SP)
  print.go:274		0x48cd25		4889442408		MOVQ AX, 0x8(SP)
  print.go:274		0x48cd2a		488d442448		LEAQ 0x48(SP), AX
  print.go:274		0x48cd2f		4889442410		MOVQ AX, 0x10(SP)
  print.go:274		0x48cd34		48c744241802000000	MOVQ $0x2, 0x18(SP)
  print.go:274		0x48cd3d		48c744242002000000	MOVQ $0x2, 0x20(SP)
  print.go:274		0x48cd46		e87597ffff		CALL fmt.Fprintln(SB)
  test.go:18		0x48cd4b		48c7042400943577	MOVQ $0x77359400, 0(SP)
  test.go:18		0x48cd53		e80872fbff		CALL time.Sleep(SB)
  test.go:19		0x48cd58		488b6c2468		MOVQ 0x68(SP), BP
  test.go:19		0x48cd5d		4883c470		ADDQ $0x70, SP
  test.go:19		0x48cd61		c3			RET
  test.go:8		0x48cd62		e81947fcff		CALL runtime.morestack_noctxt(SB)
  test.go:8		0x48cd67		e9f4feffff		JMP main.main(SB)
```


### vet
go tool vet是Go程序静态分析的工具，用于检查Go语言源码中静态错误的简单工具。go vet是go tool vet命令的简单封装。它会首先载入和分析指定的代码包,并把指定代码包中的所有Go语言源文件和".s"结尾的文件的相对路径作为参数传递给go tool vet命令。

如果go vet命令的参数是Go语言源码文件的路径,则会直接将这些参数传递给go tool vet命令。

go tool vet命令的作用是检查Go语言源代码并且报告可疑的代码编写问题。比如，在调用Printf函数时没有传入格式化字符串，以及某些不标准的方法签名，等等。该命令使用试探性的手法检查错误，因此并不能保证报告的问题确实需要解决。但是，他确实能够找到一些编译器没有捕捉到的错误。

go tool vet命令在被执行后首先解析标记并检查标记值。go tool vet命令支持的所有标记如下:

```
-all: 进行全部检查。如果有其它检查标志被设置,则命令程序会将此值设为false,默认值为true
-asmdecl: 对汇编语言源代码文件进行检查。默认值为false
-assign: 自赋值语句检查。默认值为false
-atomic: 检查代码中对代码包sync/atomic的使用是否正确。默认值false
-bool: 不推荐使用-bools别名
-bools: 检查bools
-buildtag: 检查编译标签的有效性。默认值false
-cgocall: 检查cgocall
-composites: 检查复合结构实例的初始化代码。默认值false
-c int: 查看系统调用条目限制。默认值-1,没有限制
-compositewhitelist: 是否使用复合结构检查的白名单。默认为false
-copylocks: 检查lock是否按值传递
-errorsas: 检查errors.as函数第二个参数是否指向errors类型指针
-flags: 在Json中打印分析器标志
-httpresponse: 使用Http请求返回检查错误
-json: 按json格式输出
-loopclosure: 检查嵌套函数中对循环变量的引用
-lostcancel: 检查context.WithCancel返回函数是否被调用
-methods: 检查那些拥有标准命名的方法的签名。默认值为false
-nilfunc: 空函数检测
-printf: 检查代码中对打印函数的使用是否正确。默认值为false
-printfuncs value: 需要检查的代码中使用的打印函数的名称的列表，多个函数名称之间用英文半角逗号分隔。默认值为空字符串
-rangeloops: 检查代码中对在```range```语句块中迭代赋值的变量的使用是否正确。默认值为false
-shift: 检查移位是否等于或超过整数宽度
-stdmethods: 检查已知接口方法的签名
-structtag: 检查struct字段标记是否符合reflect.StructTag.Get
-tests: 检查tests和example常见的错误用法
-unmarshal: 将非指针或非接口值传递给unmarshal
-unreachable: 死代码检查
-unsafeptr: 检查uintptr转成unsafe.Pointer类型是否无效
-unusedresult: 检查没有返回值的函数
```

#### 详细解析

* -all标记

	如果标记-all有效，那么命令程序会对目标文件进行所有已知的检查。实际上，标记-all的默认值为true。也就是说，在执行go tool vet命令且不加任何标记的情况下，命令程序对目标文件进行全面的检查。但是，只要有一个另外的标记有效，命令程序就会吧标记-all设置为false，并只会进行与有效的标记对应的检查。

* -assign标记

	该命令程序会对目标文件中的赋值语句进行自赋值操作检查(自赋值: 将一个值或者实例赋值给它本身)。
	
	```
	var s string = "Hello, World!"
	s = s    // 自赋值
	```
	
* -atomic标记

	该命令程序会对目标文件中的使用代码包sync/atomic进行原子赋值的语句进行检查。
	
	```
	// 函数AddInt32会原子性的将变量i的值加3,并返回这个新值。
	var i int32
	i = 0
	
	ni := atomic.AddInt32(&i, 3)
	fmt.Printf("i: %d, newi: %d.\n", i, ni)
	```
	
	在代码包sync/atomic中，与AddInt32类似的函数还有AddInt64、AddUint32、AddUint64和AddUintptr。-atomic会对上述这些函数的使用进行检查,检查的关注点在破坏原子性的使用上。
	
	```
	i32 = 1
	i32 = atomic.AddInt32(&i32, 3)
	```
	
	以上实例破坏了原子赋值的原子性，等号右边的atomic.AddInt32(&i32, 3)的作用时原子性的将变量i32的值加3.但该语句又将函数的返回结果赋值给了变量i32，这个第二次赋值属于对变量i32重新赋值，也使原本拥有原子性的赋值操作被拆分为了两个步骤的非原子操作。上面实例的代码应该改为:
	
	```
	atomic.AddInt(&i32, 3)
	```
	
* -buildtags标记

	该命令会对目标文件中的编译标签的格式进行检查。编译器需要根据源文件中的编译标签的内容来决定是否编译当前文件。编译标签必须出现在任何源代码文件中的头部的单行注释中，并且在后面需要有空行。
	
	条件编译就是一个实际应用场景，有些源代码文件中包含了平台相关的代码。我们希望只在某些特定平台下才编译它们。这种有选择的编译方法就是条件编译。在Go语言中，条件编译的配置就是通过编译标签来完成的。
	
	```
	// 条件编译
	$> GOOS=Linux go build -o test test.go
	```
	
* -composites标记和-compositeWhiteList标记

	该命令会对目标文件中的复合字段进行检查。
	
	```
	type Counter struct {
		Name string
		Number int
	}
	...
	c := counter{Name: "c1", Number: 0}
	```
	
	以上是一个实例，如果复合结构中的字段出现的类型错误，那么检查会打印错误信息并且会将错误代码设置为1，并且取消后续的检查。退出代码为1意味着检查程序已经报告了一个或多个问题。
	
	如果复合字段中包含了对结构体类型的字段赋值但却没有指明字段名称，那么检查程序会打印错误信息。
	
	```
	var v = flag.Flag{
		"David",
		"Usage",
		nil,
		....,
	}
	```
	
	
	这里有一个特殊的情况,当标记-compositeWhiteList有效。只要类型在白名单中,即使其初始化语句中含有未指明的字段赋值也会被提示。
	
	```
	type sliceType []string
	...
	s := sliceType{"1", "2", "3"}
	```
	
	以上实例中,它初始化的类型实际上是一个slice类型，只不过这个切面值被别名成另外一种类型而已。在这种情况下，复合类型中的字段名不需要指定，实际上这样的类型也不包含任何字段。白名单中所包含的类型都是这种情况。它们是在标准库中包含了slice类型，它们不会被检查。
	
	默认情况下，标记-compositeWhiteList是有效的。检查程序不会对它们的初始化代码进行检查，除非-compositeWhiteList标记值设置为false。


* -unreachable

	该指令用于检查不会被执行到的代码，也称为死代码。
	
	```
	// test.go
	func main() {
		return
		
		fmt.Println("XXXX")
	}
	
	$> go vet -unreachable -json test.go
	{
	"command-line-arguments": {
		"unreachable": [
			{
				"posn": "/home/work/go/go.test/src/test.go:25:2",
				"message": "unreachable code"
			}
		]
	  }
   }
	```
	
### pprof

go tool pprof命令来交互式的访问概要文件的内容。命令将会分析指定的概要文件,并会根据我们的要求为我们提供高可读性的输出信息。

在Go语言中，我们可以通过标准库runtime和runtime/pprof中的程序来生成三种包含实时性数据的概要文件，分别是CPU概要文件、内存概要文件和程序阻塞概要文件。

#### CPU概要文件

Go语言的运行时系统会以100 Hz的频率对CPU使用情况进行取样。也就是说每秒取样100次，即每10毫秒会取样一次(100Hz既足够产生有用数据，又不至于让系统产生停顿)。这里所说的对CPU使用情况情况的取样就是对当前的Goroutine的stack上的程序计数器的取样。由此，我们就可以从样本记录中分析出哪些代码时计算时间最长或者说最耗CPU资源的部分。可以通过如下代码启动对CPU的使用情况做记录。

```
var cpuProfile = flag.String("cpuprofile", "", "write cpu profile to `file`")
func StartCPUProfile() {
	if *cpuProfile != "" {
		f, err := os.Create(*cpuProfile)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Can not create cpu profile output file: %s",
                err)
            return
        }
        if err := pprof.StartCPUProfile(f); err != nil {
            fmt.Fprintf(os.Stderr, "Can not start cpu profile: %s", err)
            f.Close()
            return
        }
	}
}
```

函数StartCPUProfile首先创建了一个用于存放CPU使用情况记录的文件，这个文件就是CPU概要文件，其绝对路径由*cpuProfile的值表示。然后，我们把这个文件的实例作为参数传到函数pprof.StartCPUProfile，此时开始记录cpu使用情况。

如果需要停止CPU使用情况记录操作，调用如下函数:

```
func StopCPUProfile() {
	if *cpuProfile != "" {
		pprof.StopCPUProfile()    // 把记录的概要信息写入到指定文件
	}
}
```

在启动CPU使用情况记录操作后，runtime就会以每秒100次的频率将取样数据写入到CPU概要文件中。pprof.StopCPUProfile函数通过把cpu使用情况取样的频率设置为0来停止取样操作。当所有CPU使用情况记录到写入到CPU概要文件后，pprof.StopCPUProfile函数才会退出，从而保证cpu概要文件完整性。

#### 内存概要文件
内存概要文件(.prof)用于保存用户程序执行期间的内存使用情况。这里所说的内存使用情况，其实是程序运行过程中在heap上分配情况。Go runtime会对用户程序运行期间所有heap内存分配进行记录。无论在取样时、heap内存是否增长，只要有heap存在，分析器就会对其进行取样。

```
var memProfile = flag.String("memprofile", "", "write memory profile to `file`")

func MemProfile() {
	if *memProfile != "" {
		f, err := os.Create(*memprofile)
        if err != nil {
            log.Fatal("could not create memory profile: ", err)
        }
        defer f.Close()
        runtime.GC() // get up-to-date statistics
        if err := pprof.WriteHeapProfile(f); err != nil {
            log.Fatal("could not write memory profile: ", err)
        }
	}
}
```

默认情况下，内存使用情况的取样数据只会被保存在runtime内存中，而保存到文件的操作由pprof.WriteHeapProfile函数实现。


#### 程序IO阻塞概要文件
IO阻塞概要文件用于保存Goroutine阻塞事件的记录。可以通过runtime.SetBlockProfileRate函数设置采样间隔。

```
func startBlockProfile() {
	if *blockProfile != "" && *blockProfileRate > 0 {
		runtime.SetBlockProfileRate(*blockProfile)
	}
}
```

当*blockProfile和\*blockProfileRate的值有效时，我们会设置对Goroutine阻塞事件的取样间隔。runtime.SetBlockProfileRate函数为int类型，它的含义是分析器会在每发生几次Goroutine阻塞事件时对这些事件进行取样，默认值为1。如果我们通过runtime.SetBlockProfileRate 函数将这个取样间隔设置为0或负数,那么这个取样操作就会被取消。

程序在结束之前可以将被保存Goroutine IO阻塞事件记录存放到指定文件中。

```
fucn stopBlockProfile() {
	if *blockProfile != "" && *blockProfileRate >= 0 {
		f, err := os.Create(*blockProfile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Can not create block profile output file: %s", err)
          	return
		}
		
		if err = pprof.Lookup("block").WriteTo(f, 0); err != nil {
            fmt.Fprintf(os.Stderr, "Can not write %s: %s", *blockProfile, err)
        }
        f.Close()
	}
}
```

以上实例通过pprof.Lookup将保存在runtime时内存数据(IO阻塞采样数据)取出,并在记录的实例上调用WriteTo方法将记录写入到文件中。

我们可以通过pprof.Lookup函数去除更多种类型取样记录:

```
goroutine: 活跃Goroutine的信息记录(仅在获取时取样一次)。
threadcreate: 系统线程创建情况的记录(仅在获取时取样一次).
heap: heap内存分配情况的记录(默认每分配512KB取样一次).
block: Goroutine阻塞事件的记录(默认每发生一次阻塞事件时取样一次)。
```
在上表中，前两种记录均为一次取样的记录，具有即时性。而后两种记录均为多次取样的记录，具有实时性。实际上，后两种记录“heap”和“block”正是我们前面讲到的内存使用情况记录和程序阻塞情况记录。