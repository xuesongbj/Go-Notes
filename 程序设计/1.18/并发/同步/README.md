# 同步

同步并非用来取代锁，各有不同使用场景。</br>

通道解决高级别逻辑层次并发架构，锁则用来保护低级别局部代码安全。 </br>

&nbsp;

* 竞态条件：多线程同时读写共享资源(竞态资源)。
* 临界区：读写竞态资源的代码片段。

&nbsp;

* 互斥锁：同一时刻，只有一个线程能进入临界区。
* 读写锁：写独占(其他读写均被阻塞)，读共享。
* 信号量：允许指定数量线程进入临界区。
* 自旋锁：失败后，以循环积极尝试。(无上下文切换，小粒度)。

&nbsp;

* 悲观锁：操作前独占锁定。
* 乐观锁：假定无竞争，后置检查。(Lock Free，CAS)

&nbsp;

标准库 `sync` 提供了多种锁，另有原子操作等。

&nbsp;

* `Mutex`：互斥锁。
* `RWMutex`：读写锁。

&nbsp;

* `WaitGroup`：等待一组任务结束。
* `Cond`：单播或广播唤醒其他任务。
* `Once`：确保只调用一次(函数)。

* `Map`：并发安全字典，(少写多度，数据不重叠)
* `Pool`：对象池。(缓存对象可被回收)

&nbsp;

## 竞争检查

测试阶段，以 `-race` 编译，注入竞争检查(data race detection) 指令。

* 有较大性能损失，避免在基准测试和发布版本中使用。
* 有不确定性，不能保证百分百测出。
* 单元测试有效完整，定期执行竞争检查。

```go
package main

import (
    "sync"
)

func main() {
    var wg sync.WaitGroup
    wg.Add(2)

    x := 0

    go func() {
        defer wg.Done()
        x++
    }()

    go func() {
        defer wg.Done()
        println(x)
    }()


    wg.Wait()
}
```

```bash
$ go build -race && ./test

==================
WARNING: DATA RACE

Read at 0x00c0000160d8 by goroutine 7:
  main.main.func2()
      /root/go/test/main.go:20 +0x74

Previous write at 0x00c0000160d8 by goroutine 6:
  main.main.func1()
      /root/go/test/main.go:15 +0x86

Goroutine 7 (running) created at:
  main.main()
      /root/go/test/main.go:18 +0x1d6

Goroutine 6 (finished) created at:
  main.main()
      /root/go/test/main.go:13 +0x12e
      
==================
1
Found 1 data race(s)
```

&nbsp;

## 互斥锁

文档中标明 "must not be copied"，应避免复制导致锁机制失效。

```go
type data struct {
    sync.Mutex
}

// go vet: passes lock by value
func (d data) test(s string) { 
    d.Lock()
    defer d.Unlock()
}
```

&nbsp;

控制在最小范围内，及早释放。

```go
// 错误用法
func {
    m.Lock()
    defer m.Unlock
    
    url := cache["key"]
    get(url)            // 该操作并不需要保护，延长锁占用。
}

// 正确用法
func {
    m.Lock()
    url := cache["key"]
    m.Unlock()
    
    get(url)
}
```

&nbsp;

不支持递归锁。

```go
func main() {
    var m sync.Mutex
    m.Lock()
    
    {
        // m.Lock()
              // ~~~~~~ fatal error: all goroutines are asleep - deadlock!

        // m.Unlock()
    }
    
    m.Unlock()
}
```

&nbsp;

## 读写锁

某些场合以读写锁替代互斥锁，可提升性能。

```go
package main

import (
    "fmt"
    "sync"
    "time"
)

func main() {
    var wg sync.WaitGroup
    var rw sync.RWMutex

    x := 0

    // 1 写
    wg.Add(1)
    go func() {
        defer wg.Done()
    
        for i := 0; i < 5; i++ {
            rw.Lock()
    
            time.Sleep(time.Second) // 模拟长时间操作！
            now := time.Now().Format("15:04:05")
            x++
            fmt.Printf("[W] %d, %v\n", x, now)
    
            rw.Unlock()
        }
    }()

    // n 读
    for i := 0; i < 5; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()

            for n := 0; n < 5; n++ {
                rw.RLock()

                time.Sleep(time.Second)
                now := time.Now().Format("15:04:05")
                fmt.Printf("    [R%d] %d, %v\n", id, x, now)

                rw.RUnlock()
            }
        }(i)
    }


    wg.Wait()
}

/*
[W] 1, 11:23:17         // 独占
    [R4] 1, 11:23:18    // 并发
    [R3] 1, 11:23:18
    [R1] 1, 11:23:18
    [R0] 1, 11:23:18
    [R2] 1, 11:23:18
[W] 2, 11:23:19
    [R4] 2, 11:23:20
    [R3] 2, 11:23:20
    [R1] 2, 11:23:20
    [R0] 2, 11:23:20
    [R2] 2, 11:23:20
*/
```

&nbsp;

## 条件变量

内部以计数器和队列作为单播(signal)和广播(broadcast)依据。

引入外部锁作为竞态资源保护，可与其他逻辑同步。

```go
func main() {
    var wg sync.WaitGroup

    cond := sync.NewCond(&sync.Mutex{})
    data := make([]int, 0)

    // 1 写
    wg.Add(1)
    go func() {
        defer wg.Done()
    
        for i := 0; i < 5; i++ {

            // 保护竟态资源。
            cond.L.Lock()
            data = append(data, i + 100)
            cond.L.Unlock()

            // 唤醒一个。
            cond.Signal()
        }
    
        // 唤醒所有（剩余）。
        // cond.Broadcast()
    }()

    // n 读
    for i := 0; i < 5; i++ {
        wg.Add(1)

        go func(id int) {
            defer wg.Done()

            // 锁定竟态资源。
            cond.L.Lock()

            // 循环检查是否符合后续操作条件。
            // 如条件不符，则继续等待。
            for len(data) == 0 {
                cond.Wait()
            }

            x := data[0]
            data = data[1:]

            cond.L.Unlock()
            println(id, ":", x)
        }(i)
    }

    wg.Wait()
}
```

&nbsp;

为什么 `Wait` 之前必须加锁？除锁定竞态资源外，还与其内部设计有关。

```go
// cond.go

func (c *Cond) Wait() {
    c.checker.check()
    t := runtime_notifyListAdd(&c.notify)
    c.L.Unlock()
    runtime_notifyListWait(&c.notify, t)
    c.L.Lock()
}
```

&nbsp;

## 单次执行

确保仅执行一次，无论后续是同一函数或不同函数都不行。

```go
func main() {
    var once sync.Once

    f1 := func() { println("1") }
    f2 := func() { println("2") }

    once.Do(f1)
    
    // 以下目标函数不会执行。
    once.Do(f1) 
    once.Do(f2)
    once.Do(f2)
}

// 1
```

&nbsp;

以内部状态(`done`)记录第一次执行，与具体什么函数无关。

```go
func main() {
    var once sync.Once

    once.Do(func() { println("1") })
    once.Do(func() { println("1") })
    once.Do(func() { println("2") })
    once.Do(func() { println("2") })
}

// 1
```

[Data Race Detector](https://go.dev/doc/articles/race_detector) </br>
[Google Sanitizers](https://github.com/google/sanitizers)
