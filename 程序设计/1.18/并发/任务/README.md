# 任务

简单将 `goroutine` 归为协程(coroutine) 并不合适。</br>

类似多线程和协程的综合体，最大限度提升执行效率，发挥多核处理能力。</br>

* 极小初始栈(2KB)，如需扩张。
* 无锁内存分配和复用，提升并发性能。
* 调度器平衡任务队列，充分利用多处理器。
* 线程自动休眠和唤醒，减少内核开销。
* 基于信号(signal) 实现抢占式任务调度。

&nbsp;

关键字 `go` 将目标函数和参数**打包**(非执行)成并发任务单元，放入待运行队列。</br>

无法确定执行时间、执行次序，以及执行线程，由调度器负责处理。

```go
func main() {
    
    // 打包并发任务（函数 + 参数）。
    // 并非立即执行。
    
    go println("abc")
    go func(x int) { println(x) }(123)

    // 上述并发任务会被其他饥饿线程取走。
    // 执行时间未知，下面这行可能先输出。
    
    println("main")
    time.Sleep(time.Second)
}
```

&nbsp;

参数立即计算并复制。

```go
var c int

func inc() int {
    c++
    return c
}

func main() {
    a := 100

    // 立即计算出参数 (100, 1)，并复制。
    // 内部 sleep 目的是让 main 先输出。
    
    go func(x, y int) {
        time.Sleep(time.Second)
        println("go:", x, y)     // 100， 1
    }(a, inc())            

    a += 100                   
    println("main:", a, inc())   // 200， 2

    time.Sleep(time.Second * 3)
}

// main: 200 2
//   go: 100 1
```
