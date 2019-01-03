# Go1.12.beta1 改动


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
2. Go1.12不再支持GOCACHE=off
3. Go1.12最后一个版本支持使用binary-only发布包

```
bash# go install golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow
bash# go vet -vettool=$(which shadow)
```


## CGO
 Go1.12将C EGLDisply 转换为Go unitptr。
 

## Modules
1. 同时执行download和extract可以同时进行(safe)。
2. go.mod支持指定go版本,如果没有指定默认使用当前系统版本。
3. go无法导入package时,将从缓存中或网络源的模块中获取,如果找到的package没有版本信息,默认使用time.Time生成一个伪版本。


## 编译器工具链
1. Live variable analysis 进行了改进,这样finalizers(类似OOP析构方法)更快执行。(如果这是一个问题,适当添加runtime.KeepAlive进行解决)。
2. 更多函数可以被内联,性能优化。(内联使用runtime.CallersFrames调用;非内联使用runtime.Callers调用)
3. 支持指定Go版本,-lang=<go_version>
4. Linux/arm64支持关闭perf。(make.bash设置GOEXPERIMENT=noframepointer)
5. safe编译模式被删除(-u gcflag启用)


## Go doc
1. godoc命令行界面删除,仅支持一个Web服务器。
2. go doc新增-all标识,功能和godoc相同。


## Trace
支持利用率曲线图,分析垃圾收集器对应用程序延迟和吞吐量很有用。


## Assembler
ARM64平台R18寄存器重命名为R18_PLATFORM.


## runtime
1. 提升heap扫描能力,提升垃圾回收能力。
2. heap不可重用的空间,Go1.12更加积极释放给操作系统。Linux上,仅操作系统内存空间压力较大时,才会将heap空间释放给操作系统(处于效率考虑,可观产RSS变化情况)。
3. 支持GODEBUG=cpu.extension=off,标准库、runtime禁止使用可选CPU指令集扩展。

## Core library
1. crypto/tls 支持TLS1.3。
2. unsafe.Pointer(nil),增加对nil验证。如果对nil取地址,会导致编译器异常。


其它library只是微调整,可以参考:[Go 1.12.beta1](https://github.com/golang/go/blob/master/doc/go1.12.html)
