# Go 1.12变化


## port 
1. Linux/ARM64 支持竞争检查(race detector)。
2. Go1.12是FreeBSD 10.x支持的最后一个版本。Go1.13需要FreeBSD 11.2＋或FreeBSD 12.0+以上版本支持。
3. Linux/ppc64支持CGO。


### Windows
1. Go1.12新ARM接口支持Windows10 IOT Core 32位架构。

### AIX 
1. Go1.12 支持AIX7.2+ power 8 架构。


### Darwin
1. Go1.12是MacOS10.10最后一个版本的支持。Go1.13需要MocOS10.11+系统支持。
2. Go1.12 系统调用使用libSystem,确保向下兼容。现在的一些系统调用无法使(syscall.Getdirentries等)。



## Tools
1. go vet替代go tool vet。go vet是对go tool vet改写;go vet不再支持-shadow选项,直接使用go vet进行shadow检查。

```
bash# go install golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow
bash# go vet -vettool=$(which shadow)
```
2. Go 1.12不再支持GOCACHE=off。
3. Go 1.12最后一个版本支持使用binary-only发布包。
4. go tour从go安装包中移除。移除后被迁移到golang.org/x/tour。

```
bash# go get -u golang.org/x/tour
bash# tour   // 启动tour
```



## CGO
 1. C EGLDisply转换为Go unitptr。
 

## Modules
1. 同时执行download和extract可以同时进行(safe)。
2. go.mod支持指定go版本,如果没有指定默认使用当前系统版本。
3. go无法导入package时,将从缓存中或网络源的模块中获取,如果找到的package没有版本信息,默认使用time.Time生成一个伪版本。


## 编译器工具链
1. 编译器对运行时的变量进行了改进,这样finalizers(类似OOP析构方法)更快执行。(如果这是一个问题,适当添加runtime.KeepAlive进行解决)。
2. 更多函数可以被内联,性能优化。(内联使用runtime.CallersFrames调用;非内联使用runtime.Callers调用)
3. 支持指定Go版本,-lang=<go_version>
4. Linux/arm64支持关闭perf。(make.bash设置GOEXPERIMENT=noframepointer)
5. 编译器生成的DWARF调试信息有了很多改进,包括对参数打印和变量位置信息的改进。
6. Go语言现在还在Linux/ARM64上维护stack frame pointers,方便使用perf等分析工具使用。这种方式的stack frame pointers需要的运行时间开销很小,平均大约为3%。如果构建时不使用该工具链,在make.bash设置GOEXPERIMENT=noframepointer。
7. "safe"编译模式(由-u gcflag)已经被删除。


## godoc
1. godoc命令行界面删除,仅支持一个Web服务器。
2. go doc新增-all标识,功能和godoc相同。


## 跟踪工具(Trace)
1. 跟踪工具现在支持绘制mutator利用率曲线图,包括对执行跟踪的交叉引用。对于分析GC对程序延迟和吞吐量的影响非常有用。


## 汇编(Assembler)
ARM64平台R18寄存器重命名为R18_PLATFORM.


## 运行时(runtime)
1. 提高了GC扫描性能,这会在垃圾回收后立即减少分配延迟。
2. 当操作系统内存空间压力较大时,runtime更加积极释放给操作系统(VM地址空间不会被释放,仅PA被MMU重新映射)。在Linux上,runtime使用MADV_FREE释放未使用的内存,更为高效,但可能会导致RSS不会释放,仅当操作系统需要时(内存不够使用时)才进行回收未使用的内存空间。操作系统内核可以按需重用这些内存。
3. 在多CPU上,runtime的timer和deadline代码运行更为高效,对于提升网络连接的deadline性能有极大提升。
3. 支持GODEBUG=cpu.extension=off。标准库、runtime禁止使用可选CPU指令集扩展。
4. Go 1.12通过修复Large heap过度计数来提升内存分配准确性。

## 标准库(Core library)
1. 标准库最大的改变应该就是对TLS 1.3的支持了。默认不开启,Go 1.13中将成为默认开启功能。编码中涉及TLS的代码无须更改,使用Go 1.12重新编译后即可无缝支持TLS 1.3。
2. unsafe.Pointer(nil),增加对nil验证。如果对nil取地址,会导致编译器异常。
