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
