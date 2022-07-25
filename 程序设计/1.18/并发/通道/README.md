# 通道

鼓励使用CSP通道，以通信来代替内存共享，实现并发安全。</br>
通道(channel)行为类似消息队列。不限收发人数，不可重复消费。 </br>

> Do not communicate by sharing memory; instead, share memory by communicating. </br>
> CSP：Communicating Sequential Process.

&nbsp;

## 同步

没有数据缓冲区，须收发双方到场直接交换数据。

&nbsp;

* 阻塞，直到另一方准备妥当或当通道关闭。
* 可通过 `cap == 0` 判断为无缓冲通道。

```go
func main() {
    quit := make(chan struct{})
    data := make(chan int)

    go func() {
        data <- 11
    }()

    go func() {
        defer close(quit)
    
        println(<- data)
        println(<- data)
    }()

    data <- 22
    <- quit
}
```

&nbsp;

## 异步

通道自带固定大小缓冲区(buffer)。有数据或空位时，不会阻塞。

* 用 `cap`、 `len` 获取缓冲区大小和当前缓冲数据量。

```go
func main() {
    quit := make(chan struct{})
    data := make(chan int, 3)

    data <- 11
    data <- 22
    data <- 33
    
    println(cap(data), len(data))  // 3 3

    go func() {
        defer close(quit)

        println(<- data)
        println(<- data)
        println(<- data)

        println(<- data)  // block
    }()

    data <- 44
    <- quit
}
```

&nbsp;

缓冲区大小仅是内部属性，不属于类型组成部分。</br>
通道变量本身就是指针，可判断是否为同一对象或`nil`。

```go
func main() {
    var a, b chan int = make(chan int, 3), make(chan int)
    var c chan bool
    
    println(a == b)               // false
    println(c == nil)             // true
    println(a, unsafe.Sizeof(a))  // 0xc..., 8
}
```

&nbsp;

## 关闭

对于 `closed` 或 `nil` 通道，规则如下：

* 无论收发，`nil` 通道都会阻塞。
* 不能关闭 `nil` 通道。

&nbsp;

* 重复关闭通道，引发 `panic` ！

&nbsp;

* 向已关闭通道发送数据，引发 `panic` ！
* 从已关闭通道接收数据，返回缓冲数据或零值。

&nbsp;

没有判断通道是否已被关闭的直接方法，只能透过收发模式获知。

```go
func main() {
    c := make(chan int)
    close(c)

    // close(c)
    // ~~~~~ panic: close of closed channel
    
    // 不会阻塞，返回零值。
    
    println(<- c) // 0
    println(<- c) // 0
}
```

&nbsp;

为避免重复关闭，可包装 `close` 函数。</br>
也可以类似方式封装 `send`、`recv` 操作。</br>

&nbsp;

```go
func closechan[T any](c chan T) {
    defer func(){
        recover()
    }()

    close(c)
}

func main() {
    c := make(chan int, 2)

    closechan(c)
    closechan(c)
}
```

&nbsp;

保留关闭状态。注意，为并发安全，关闭和获取关闭状态应保持同步。

> 可使用 `sync.RWMutex`、`sync.Once` 优化设计。

```go
type Queue[T any] struct {
    sync.Mutex

    ch     chan T
    cap    int
    closed bool
}

func NewQueue[T any](cap int) *Queue[T] {
    return &Queue[T]{
        ch: make(chan T, cap),
    }
}

func (q *Queue[T]) Close() {
    q.Lock()
    defer q.Unlock()

    if !q.closed {
        close(q.ch)
        q.closed = true
    }
}

func (q *Queue[T]) IsClosed() bool {
    q.Lock()
    defer q.Unlock()

    return q.closed
}

// ---------------------------------

func main() {
    var wg sync.WaitGroup
    q := NewQueue[int](3)

    for i := 0; i < 10; i++ {
        wg.Add(1)
    
        go func() {
            defer wg.Done()
            defer q.Close()
            println(q.IsClosed())
        }()
    }

    wg.Wait()
}
```

&nbsp;

利用 `nil` 通道阻止退出。

```go
func main() {
    <-(chan struct{})(nil)    // select{}
}
```
