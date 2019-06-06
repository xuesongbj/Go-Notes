# synchronous

## 实例
适用channel和mutex实现goroutine数据安全

### 适用场景
有些场景,比如不同的Goroutine之间进行通信,那么适用channel是最好不过了,但是再一些并发场景下使用channel来保证并发安全,那么性能表现肯定比不上Mutex。
