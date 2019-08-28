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
	
	