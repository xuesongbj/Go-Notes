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
