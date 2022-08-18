# GODEBUG

查看运行时内部垃圾回收、并发调度和物理内存释放信息。

```go
$ go version
go version go1.18.3 linux/amd64
```

&nbsp;

## 垃圾回收

定时输出垃圾回收状态。重点是时间占比，以及回收后堆大小。

```go
package main

import "time"

//go:noinline
func test() []byte {
    x := make([]byte, 0, 10<<20)
    return x
}


func main() {
    for i := 0; i < 100; i++ {
        _ = test()
        time.Sleep(time.Second)
    }
}
```

```bash
$ go build -o test
$ GODEBUG=gctrace=1 ./test

gc 1 @0.001s 7%: 0.031+... ms clock, 0.063+... ms cpu, 10->10->0 MB,  10 MB goal, ..., 2 P
gc 2 @1.011s 0%: 0.035+... ms clock, 0.071+... ms cpu, 10->10->10 MB, 10 MB goal, ..., 2 P
------------------------------------------------------------------------------------------
   1 2       3   4                                     5              6                7
```

* `1`：执行次数。(自增)
* `2`: 程序运行时间。(当前时间 - 启动时间)
* `3`: 程序启动以来，垃圾回收耗时百分比。
* `4`: 回收各阶段耗时。
* `5`: 回收前后堆大小。(开始前、开始后、存活)
* `6`: 期望堆大小。(下次启动的粗略阈值)
* `7`: 所使用P数量。

&nbsp;

## 并发调度

调度器相关组件(GMP) 状态跟踪。

```bash
$ GODEBUG=schedtrace=1000 ./test

SCHED    0ms: gomaxprocs=4 idleprocs=4 threads=5 spinningthreads=0 idlethreads=3 runqueue=0 [0 0 0 0]
SCHED 1010ms: gomaxprocs=4 idleprocs=4 threads=5 spinningthreads=0 idlethreads=3 runqueue=0 [0 0 0 0]
SCHED 2014ms: gomaxprocs=4 idleprocs=4 threads=7 spinningthreads=0 idlethreads=5 runqueue=0 [0 0 0 0]
-----------------------------------------------------------------------------------------------------
      1       2            3           4         5                 6             7
```

* `1`: 跟踪启动时间。
* `2`: P总数。
* `3`: 空闲P数量。
* `4`: M总数。
* `5`: 自旋M数量。
* `6`: 空闲M数量。
* `7`: 全局及各P本地队列任务数量。

&nbsp;

> 正常情况下M及底层线程不会释放。</br>
> 只有在出错时(比如没有调用UnlockOSThread)才会退出并释放。

```go
var count int64

func test() {
    atomic.AddInt64(&count, 1)
    defer atomic.AddInt64(&count, -1)
    
    runtime.LockOSThread()
    // defer runtime.UnlockOSThread()
    
    time.Sleep(time.Second)
}


func main() {
    for i := 0; i < 2000; i++ {
        go test()
    }
    
    for {
        time.Sleep(time.Second)
    }
}
```

```bash
$ GODEBUG=schedtrace=1000 ./test

SCHED 0ms:    gomaxprocs=2 idleprocs=1 threads=5 spinningthreads=0 idlethreads=2 runqueue=0 [0 0]
SCHED 1006ms: gomaxprocs=2 idleprocs=2 threads=2004 spinningthreads=0 idlethreads=3 runqueue=0 [0 0]

SCHED 2014ms: gomaxprocs=2 idleprocs=2 threads=9 spinningthreads=0 idlethreads=6 runqueue=0 [0 0]
SCHED 3017ms: gomaxprocs=2 idleprocs=2 threads=9 spinningthreads=0 idlethreads=6 runqueue=0 [0 0]
SCHED 4019ms: gomaxprocs=2 idleprocs=2 threads=9 spinningthreads=0 idlethreads=6 runqueue=0 [0 0]
```

> 使用 `GODEBUG=schedtrace=1000, scheddetail=1` 可输出详细信息。

