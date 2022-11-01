# 性能监控

采集测试或运行数据，分析问题，针对性能改进代码。

&nbsp;

* 尽可能排除外在干扰，比如硬件和系统被抢占。
* 开启profile会导致性能损失。
* 不同profile可能存在干扰，每次采集一种。

&nbsp;

* 目标类型
    * `cpu`
    * `alloc`
    * `heap`
    * `threadcreate`
    * `goroutine`
    * `block`
    * `mutex`

&nbsp;

* 采集方式
    * 测试：`go test -memprofile mem.out`
    * 在线: `import _ "net/http/pprof"`
    * 手工: `runtime/pprof`

&nbps;

## 测试采集

```bash
root@8d75790f92f5:/# go test -run NONE -bench . -memprofile mem.out net/http
```

&nbsp;

* `-cpuprofile`：执行时间。
* `-memprofile`：内存分配。
* `-blockprofile`：阻塞。
* `-mutexprofile`：锁竞争。

&nbsp;

* `-memprofilerate`：`runtime.MemProfileRate`.
* `-blockprofilerate`：`runtime.SetBlockProfileRate`.
* `-mutexprofilefraction`：`runtime.SetMutexProfileFraction`.

&nbsp;

命令行、服务、交互三种模式查看采集结果。

```bash
root@8d75790f92f5:/# go tool pprof -top mem.out                     # 命令行参数
root@8d75790f92f5:/# go tool pprof -http 0.0.0.0:8080 mem.out       # 服务. 推荐！
```

```bash
root@8d75790f92f5:/# go tool pprof http.test mem.out                # 交互
File: http.test
Type: alloc_space
Time: Oct 31, 2022 at 5:47pm (CST)
Entering interactive mode (type "help" for commands, "o" for options)
(pprof) top 5
Showing nodes accounting for 5827.80MB, 60.55% of 9624.10MB total
Dropped 344 nodes (cum <= 48.12MB)
Showing top 5 nodes out of 94
      flat  flat%   sum%        cum   cum%
 2407.20MB 25.01% 25.01%  2407.20MB 25.01%  net/textproto.(*Reader).ReadMIMEHeader
 1024.25MB 10.64% 35.65%  4037.02MB 41.95%  net/http.readRequest
 1014.64MB 10.54% 46.20%  1014.64MB 10.54%  net/http.copyValues
  787.63MB  8.18% 54.38%   787.63MB  8.18%  net/http.readCookies
  594.08MB  6.17% 60.55%   594.08MB  6.17%  net/url.parse
```

* `flat`：仅当前函数，不包括它调用的其他函数。
* `sum`：列表前几行所占百分比总合。
* `cum`：当前函数调用堆栈累计。

&nbsp;

找出目标，用 `peek` 命令列出调用来源。

```bash
(pprof) peek malg
Showing nodes accounting for 9624.10MB, 100% of 9624.10MB total
----------------------------------------------------------+-------------
      flat  flat%   sum%        cum   cum%   calls calls% + context
----------------------------------------------------------+-------------
                                               3MB   100% | runtime.newproc1
       3MB 0.031% 0.031%        3MB 0.031%                | runtime.malg
----------------------------------------------------------+-------------
```

&nbsp;

也可用 `list` 输出源码样式，更直观定位。

```bash
(pprof) list malg
Total: 9.40GB
ROUTINE ======================== runtime.malg in /usr/local/go/src/runtime/proc.go
       3MB        3MB (flat, cum) 0.031% of Total
         .          .   4063:   execLock.unlock()
         .          .   4064:}
         .          .   4065:
         .          .   4066:// Allocate a new g, with a stack big enough for stacksize bytes.
         .          .   4067:func malg(stacksize int32) *g {
       3MB        3MB   4068:   newg := new(g)
         .          .   4069:   if stacksize >= 0 {
         .          .   4070:       stacksize = round2(_StackSystem + stacksize)
         .          .   4071:       systemstack(func() {
         .          .   4072:           newg.stack = stackalloc(uint32(stacksize))
         .          .   4073:       })
```

&nbsp;

打开浏览器查看

```bash
(pprof) web
(pprof) web malg
```

&nbsp;

## 在线采样

```go
package main

import (
    "net/http"
    _ "net/http/pprof"
)

func main() {
    http.ListenAndServe(":6060", http.DefaultServeMux)
}
```

```bash
$> go tool pprof http://localhost:6060/debug/pprof/heap

$> curl http://localhost:6060/debug/pprof/heap -o mem.out
$> go tool pprof mem.out
```

&nbsp;

## 手工采集

```go
package main

import (
    "runtime/pprof"
    "os"
)

func main() {
    pprof.StartCPUProfile(os.Stdout)
    defer pprof.StopCPUProfile()
}
```

&nbsp;

## 执行跟踪

捕获运行期事件。</br>
相比 `profile` 采样统计，`trace`关注一个时段内的执行过程。

* G 如何执行。
* GC 等待核心事件。
* 不良的并行化。

```bash
$> go test -trace trace.out net/http            # 采样
$> go tool trace -http 0.0.0.0:8080 trace.out   # 服务
```

&nbsp;

注入方式，指定采样时长。

```bash
$> curl http://localhost:8080/debug/pprof/trace?seconds=5 -o trace.out
```

&nbsp;

手工方式。

```go
package main

import (
    "os"
    "runtime/trace"
)

func main() {
    out, _ := os.Create("trace.out")
    defer out.Close()

    trace.Start(out)
    defer trace.Stop()

    test()
}
```

&nbsp;

[Diagnostics](https://go.dev/doc/diagnostics)
