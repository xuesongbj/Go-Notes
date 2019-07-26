# time

## 单调时间和壁挂时间

### CLOCK\_MONOTONIC 和 CLOCK\_REALTIME

CLOCK_MONOTONIC是monotonic time;CLOCK_REALTIME是wall time。

monotonic time字面意思是单调时间,实际它指的是系统启动以后流逝的时间,这是由变量jiffies记录系统每次启动时jiffies初始化为0，每来一个timer interrupt，jiffies加1，也就是说它代表系统启动后流逝的tick数。jiffies一定是单调递增的，因为时间不够逆。


wall time字面意思是挂钟时间，实际上就是指的是现实的时间，这是由变量xtime来记录的。系统每次启动时将CMOS上的RTC时间读入xtime，这个值是"自1970-01-01起经历的秒数、本秒中经历的纳秒数"，每来一个timer interrupt，也需要去更新xtime。


#### monotonic time Vs. wall time

wall time不一定单调递增的。wall time是指现实中的实际时间，如果系统要与网络中某个节点时间同步、或者由系统管理员觉得这个wall time与现实时间不一致，有可能任意的改变这个wall time。最简单的例子是，我们用户可以去任意修改系统时间，这个被修改的时间应该就是wall time，即xtime，它甚至可以被写入RTC而永久保存。



### Time时注意事项

1. 使用Time时通常应该按值进行存储,而不是指针。
2. Time值是并发安全的,除了GobDecode,Unmarshalbinary,UnmarshalJSON和UnmarshalText。
3. 可以使用Before、After和Equal方法进行时间比较运算。

## 源码剖析

### 数据结构

#### Time

```
type Time struct {

	// 壁挂时间,包括秒、纳秒和可选的单调时钟读数(以纳秒为单位)
	wall uint64
	
	// 64位有符号单调时钟读数,纳秒
	ext  int64
	
	// 设置时区,
	// nil表示UTC时区, 默认UTC时区
	loc *Location
}
```