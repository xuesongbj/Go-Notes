# time

## 单调时间和壁挂时间
操作系统中如果需要显示时间的时候,会使用wall time;而需要测量时间的时候,会使用monotonic time。

但为了避免拆分API,time包将这两个时间合并在time.Time这个结构中,当需要读取时间的时候会使用wall time,如果需要测量时间就会使用monotonic time。

#### wall time

* time.Since(start)
* time.Until(deadline)
* time.Now().Before(deadline)

#### monotonic time

* t.AddDate(y, m, d)
* t.Round(d)
* t.Truncate(d)
	

### CLOCK\_MONOTONIC 和 CLOCK\_REALTIME

CLOCK_MONOTONIC是monotonic time;CLOCK_REALTIME是wall time。

monotonic time字面意思是单调时间,实际它指的是系统启动以后流逝的时间,这是由变量jiffies记录系统每次启动时jiffies初始化为0，每来一个timer interrupt，jiffies加1，也就是说它代表系统启动后流逝的tick数。jiffies一定是单调递增的，因为时间不够逆。


wall time字面意思是挂钟时间，实际上就是指的是现实的时间，这是由变量xtime来记录的。系统每次启动时将CMOS上的RTC时间读入xtime，这个值是"自1970-01-01起经历的秒数、本秒中经历的纳秒数"，每来一个timer interrupt，也需要去更新xtime。


#### monotonic time Vs. wall time

wall time不一定单调递增的。wall time是指现实中的实际时间，如果系统要与网络中某个节点时间同步、或者由系统管理员觉得这个wall time与现实时间不一致，有可能任意的改变这个wall time。最简单的例子是，我们用户可以去任意修改系统时间，这个被修改的时间应该就是wall time，即xtime，它甚至可以被写入RTC而永久保存。



### Time时注意事项
1. Time能够代表纳秒精度的时间。
2. 因为Time并非并发安全,所以在存储或者传递的时候,都应该使用值引用。
3. 在Go中, == 运算符不仅仅比较时刻,还会比较Location以及单调时钟,因此在不保证所有时间设置为相同的位置的时候,不应该将time.Time作为map或者database的健。如果必须要使用,应该通过UTC或者Local方法将单调时间剥离。

## 源码剖析

### 数据结构

#### Time


```
type Time struct {
   // wall 和 ext 字段共同组成 wall time 秒级、纳秒级，monotonic time 纳秒级
   // 的时间精度，先看下 wall 这个 无符号64位整数 的结构。
   //
   //          +------------------+--------------+--------------------+
   // wall =>  | 1 (hasMonotonic) | 33 (second)  |  30 (nanosecond)   |
   //          +------------------+--------------+--------------------+
   // 
   // 所以 wall 字段会有两种情况，分别为
   // 1. 当wall字段的hasMonotonic为0时,second位也全部为0,ext字段会存储
   //    从 1-1-1 开始的秒级精度时间作为 wall time 。
   // 2. 当 wall 字段的 hasMonotonic 为 1 时，second 位会存储从 1885-1-1 开始的秒
   //    级精度时间作为 wall time，并且 ext 字段会存储从操作系统启动后的纳秒级精度时间
   //    作为 monotonic time 。
   // 壁挂时间,包括秒、纳秒和可选的单调时钟读数(以纳秒为单位)
	wall uint64
	
	// 64位有符号单调时钟读数,纳秒
	ext  int64
	
	// Location 作为当前时间的时区，可用于确定时间是否处在正确的位置上。
    // 当 loc 为 nil 时，则表示为 UTC 时间。
    // 因为北京时区为东八区，比 UTC 时间要领先 8 个小时，
    // 所以我们获取到的时间默认会记为 +0800
	loc *Location
}
```

#### Now
Now函数主要通过now函数实现,返回三个变量second/nsec/mono,而now函数的实现由runtime·walltime实现,通过调用runtime·vdsoClockgettimeSym或runtime·vdsoGettimeofdaySym实现,返回本机当前秒、纳秒以及单调时钟.

```
// 单调开始时间
func runtimeNano() int64
var startNano int64 = runtimeNano() - 1

// 返回本机时间
func Now() Time {
    // 秒、毫秒、单调时钟
	sec, nsec, mono := now()

	// 单调时间
	mono -= startNano

	// 返回壁挂时间、单调时间和时区
	// 计算从1885-01-01开始到现在秒数
	// unixTointernal = 1970-01-01 00:00:00
	// minWall = 1885-01-01 00:00:00
	sec += unixToInternal - minWall
	
	// 如果有溢出,则不能用wall的second保存完整时间戳
	if uint64(sec)>>33 != 0 {
		
		// 返回自1970-01-01 00:00:00开始的秒数
		return Time{uint64(nsec), sec + minWall, Local}
	}
	return Time{hasMonotonic | uint64(sec)<<nsecShift | uint64(nsec), mono, Local}
}

// 有runtime.now
func now() (sec int64, nsec int32, mono int64
```

```
// func walltime() (sec int64, nsec int32)
TEXT runtime·walltime(SB),NOSPLIT,$0-12
	MOVQ	SP, BP	// Save old SP; BP unchanged by C code.

	get_tls(CX)
	MOVQ	g(CX), AX
	MOVQ	g_m(AX), BX // BX unchanged by C code.

	// Set vdsoPC and vdsoSP for SIGPROF traceback.
	MOVQ	0(SP), DX
	MOVQ	DX, m_vdsoPC(BX)
	LEAQ	sec+0(SP), DX
	MOVQ	DX, m_vdsoSP(BX)

	CMPQ	AX, m_curg(BX)	// Only switch if on curg.
	JNE	noswitch

	MOVQ	m_g0(BX), DX
	MOVQ	(g_sched+gobuf_sp)(DX), SP	// Set SP to g0 stack

    // 获取时间(秒, 毫秒)
noswitch:
	SUBQ	$16, SP		// Space for results
	ANDQ	$~15, SP	// Align for C code

	MOVQ	runtime·vdsoClockgettimeSym(SB), AX
	CMPQ	AX, $0
	JEQ	fallback
	MOVL	$0, DI // CLOCK_REALTIME
	LEAQ	0(SP), SI
	CALL	AX
	MOVQ	0(SP), AX	// sec
	MOVQ	8(SP), DX	// nsec
	MOVQ	BP, SP		// Restore real SP
	MOVQ	$0, m_vdsoSP(BX)
	MOVQ	AX, sec+0(FP)
	MOVL	DX, nsec+8(FP)
	RET
fallback:
	LEAQ	0(SP), DI
	MOVQ	$0, SI 
	MOVQ	runtime·vdsoGettimeofdaySym(SB), AX
	CALL	AX
	MOVQ	0(SP), AX	// sec
	MOVL	8(SP), DX	// usec
	IMULQ	$1000, DX
	MOVQ	BP, SP		// Restore real SP
	MOVQ	$0, m_vdsoSP(BX)
	MOVQ	AX, sec+0(FP)
	MOVL	DX, nsec+8(FP)
	RET
```

### 时区设置
```
// 设置时区
func (t Time) UTC() Time {
	t.setLoc(&utcLoc)
	return t
}

func (t *Time) setLoc(loc *Location) {
	if loc == &utcLoc {
		loc = nil
	}

	// second位已经不能够存下Wall time秒数,需要去掉单调时钟。
	// 通过ext字段进行存储
	t.stripMono()
	t.loc = loc
}

// stripMono 去除单调时钟
func (t *Time) stripMono() {
	if t.wall&hasMonotonic != 0 {
		t.ext = t.sec()
		t.wall &= nsecMask
	}
}
```

### 时间比较
After/Before/Equal用于表两个时间早晚。

#### After
time.After函数用于比较一个时间是否晚于另外一个时间
```
// 判断t时间是否晚于u时间
func (t Time) After(u Time) bool {
	// 判断t和u是否都有单调时钟
	if t.wall&u.wall&hasMonotonic != 0 {
		return t.ext > u.ext
	}
	
	// 否则
	// 需要从wall字段获取秒数
	ts := t.sec()
	us := u.sec()
	
	// 判断t的second是否大于u的second
	// 如果second相同,则比较nanosecond
	return ts > us || ts == us && t.nsec() > u.nsec()
}
```

#### Before
time.Before函数用于比较一个时间是否早于另外一个时间

```
func (t Time) Before(u Time) bool {
	if t.wall&u.wall&hasMonotonic != 0 {
		return t.ext < u.ext
	}
	return t.sec() < u.sec() || t.sec() == u.sec() && t.nsec() < u.nsec()
}
```

#### Equal
time.Equal函数用于比较一个时间是否和另外一个时间相等,不判断时区;如果判断时区,可以使用==符号进行.

```
func (t Time) Equal(u Time) bool {
	if t.wall&u.wall&hasMonotonic != 0 {
		return t.ext == u.ext
	}
	return t.sec() == u.sec() && t.nsec() == u.nsec()
}
```
