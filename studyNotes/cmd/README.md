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
-N : 关闭优化
-S : 输出编译后生成的汇编代码(go tool compile -s 或 go build -gcflags "-S")
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
```