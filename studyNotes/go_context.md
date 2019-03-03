# Go context
Go中的context在与API和进程交互时可以派上用场，特别是在提供Web请求的生产级系统中。您可能想要通知所有goroutines停止工作并返回。
   

## 前言
在了解Context之前,需要先熟悉Go的goroutine和channel的概念。本篇不在过多阐述,如详细了解,可以看Goroutine和channel章节。

## Context
Go的Context包允许将"上下文"数据传递給你的程序.如上下文超时、截至时间或停止context等。

### 创建context
可以使用context包的Background()创建一个新的上下文。此函数返回空上下文, 仅用于高级别(根节点)。


```
ctx := context.Background()
```

### 实例
#### WitchTimeout
定义超时操作,当task在dealine之前未完成task,则超时.

```
package withtimeout

import (
    "context"
    "time"
)

var (
    SuccessCode = 0
    TimeoutCode = 1
)

func fnA(ctx context.Context) int {
    ctx, _ = context.WithTimeout(ctx, time.Duration(5) * time.Second)
    done := make(chan struct{})

    go func() {
        defer close(done)

        doSomething()
    }()

    select {
    case <-done:
        return SuccessCode
    case <-ctx.Done():
        return TimeoutCode
    }
}

func doSomething() {
    time.Sleep(time.Second * 3)
}
```

* 单元测试

```
package withtimeout

import (
    "testing"
    "context"
)

func Test_fnA(t *testing.T) {
    ctx := context.Background()
    code := fnA(ctx)
    if code == 1 {
        t.Fatal()
    }
}
```

#### withcancel
通过withcancel函数可以对goroutine进行控制。

```
package withcancel

import (
    "context"
    "time"
)

var (
    SuccessCode = 0
    TimeoutCode = 1
)

func Cancel() int {
    ctx, cancel := context.WithCancel(context.Background())

    go func() {
        defer cancel()

        go doSomething()
    }()

    select {
    case <-ctx.Done():
        return SuccessCode
    case <- time.After(time.Duration(1) * time.Second):
        return TimeoutCode
    }
}

func doSomething() {
    time.Sleep(time.Second * 2)
}
```

* 单元测试:

```
package withcancel

import (
    "testing"
)

func Test_Cancel(t *testing.T) {
    code := Cancel()
    if code == 1 {
        t.Fatal()
    }
}
```

## 源码解析

#### Context
context.Context是"上下文"的一个接口。用一个对象持有超时、取消状态,以及KV等上下文数据.使用该context的框架,检查这些状态,从而做出相应的处理机制.

创建一个上下文对象时,可指定一个parent,并从中”继承”某些状态.当向某个context发送指令时,所有childs都会收到同样的事件.


```
type Context interface {
   // 返回过期时间，ok == false 表示没有设置过期时间。
   // 可用来决定，剩余时间是否值得执行某个操作。
   Deadline() (deadline time.Time, ok bool)

   // 通过 channel closed 来判断是否已取消或过期。
   Done() <-chan struct{}

   // 取消或过期的具体原因。
   Err() error

   // 访问关联数据。
   Value(key interface{}) interface{}
}
```

### Background

Background总是返回同一个对象,一个什么也做不了的empty context。所以它适合作为root,用于将其他context组织起来。相关context内部都有专门字段存储parent,以此构成多级衍生关系.

首先是实现了Context接口的emptyCtx，所有方法都直接返回。

```
type emptyCtx int

var (
	background = new(emptyCtx)
	todo       = new(emptyCtx)
)

func Background() Context {
	return background
}
```

### Cancel
用匿名字段Context存储parent.

```
type cancelCtx struct {
	Context

	mu       sync.Mutex            
	done     chan struct{}         
	children map[canceler]struct{} 
	err      error
}

// newCancelCtx returns an initialized cancelCtx.
func newCancelCtx(parent Context) cancelCtx {
	return cancelCtx{Context: parent}
}
```

除创建context对象外,还需返回一个cancel函数,以便向context发出取消事件.

```
func WithCancel(parent Context) (ctx Context, cancel CancelFunc) {
	c := newCancelCtx(parent)
	propagateCancel(parent, &c)
	return &c, func() { c.cancel(true, Canceled) }
}
```

调用parentCancelCtx 返回内置类型parent,并添加到children字典构成层级关系.如果是用户自定义类型,则构不成层级关系,直接启动goroutine监控done channel.

```
func propagateCancel(parent Context, child canceler) {
	if parent.Done() == nil {
		return // parent is never canceled
	}
	if p, ok := parentCancelCtx(parent); ok {
		p.mu.Lock()
		if p.err != nil {
			// parent has already been canceled
			child.cancel(false, p.err)
		} else {
			if p.children == nil {
				p.children = make(map[canceler]struct{})
			}
			p.children[child] = struct{}{}
		}
		p.mu.Unlock()
	} else {
		go func() {
			select {
			case <-parent.Done():
				child.cancel(false, parent.Err())
			case <-child.Done():
			}
		}()
	}
}
 

func parentCancelCtx(parent Context) (*cancelCtx, bool) {
	for {
		switch c := parent.(type) {
		case *cancelCtx:
			return c, true
		case *timerCtx:
			return &c.cancelCtx, true
		case *valueCtx:
			parent = c.Context
		default:
			return nil, false
		}
	}
}
```

当调用WithCancel返回的cancel函数时:

	- 关闭done
	- 向所有child发出通知(递归调用)
	- 该context已经无用,移除层级关系
	
```
func (c *cancelCtx) cancel(removeFromParent bool, err error) {
	if err == nil {
		panic("context: internal error: missing cancel error")
	}
	c.mu.Lock()
	if c.err != nil {
		c.mu.Unlock()
		return // already canceled
	}
	
	// 设置 Err，并 close(done) 发出通知。
	c.err = err
	if c.done == nil {
		c.done = closedchan
	} else {
		close(c.done)
	}
	
	// 遍历所有 child，执行 cancel 操作。
	for child := range c.children {
		// NOTE: acquiring the child's lock while holding parent's lock.
		child.cancel(false, err)
	}
	
	// 移除所有 child。
	c.children = nil
	c.mu.Unlock()

	// 从 parent 中移除。
	if removeFromParent {
		removeChild(c.Context, c)
	}
}
 
func removeChild(parent Context, child canceler) {
	p, ok := parentCancelCtx(parent)
	if !ok {
		return
	}
	p.mu.Lock()
	if p.children != nil {
		delete(p.children, child)
	}
	p.mu.Unlock()
}
```

#### timerCtx
先基于cancelCtx实现记时方式的上下文.

```
type timerCtx struct {
	cancelCtx
	timer *time.Timer // Under cancelCtx.mu.

	deadline time.Time
}

func (c *timerCtx) Deadline() (deadline time.Time, ok bool) {
	return c.deadline, true
}
```

Deadline无非是额外设置的计数器,以便主动引发超时取消事件.至于Timeout 和 Deadline无非计时参数类型不同而已.


** 注意: 会重新调用newCancelCtx,因此不会和parent共享同一 done channel. **

```
func WithDeadline(parent Context, deadline time.Time) (Context, CancelFunc) {
    // 如果 parent 更早过期，无需设置。
    if cur, ok := parent.Deadline(); ok && cur.Before(deadline) {
        return WithCancel(parent)
    }

    c := &timerCtx{
        cancelCtx: newCancelCtx(parent),
        deadline:  deadline,
    }
    propagateCancel(parent, c)

    // 计算超时时间。
    d := deadline.Sub(time.Now())
    if d <= 0 {
        c.cancel(true, DeadlineExceeded) // deadline has already passed
        return c, func() { c.cancel(true, Canceled) }
    }

    c.mu.Lock()
    defer c.mu.Unlock()

    // 设定计时器，以便引发超时取消操作。
    if c.err == nil {
        c.timer = time.AfterFunc(d, func() {
            c.cancel(true, DeadlineExceeded)
        })
    }

    return c, func() { c.cancel(true, Canceled) }
}


func WithTimeout(parent Context, timeout time.Duration) (Context, CancelFunc) {
    return WithDeadline(parent, time.Now().Add(timeout))
}
```

#### Value
保存上下文数据

```
type valueCtx struct {
	Context
	key, val interface{}
}
```

key 需支持比较操作.按此做法,Value Context 可以构建多级结构.

```
func WithValue(parent Context, key, val interface{}) Context {
	if key == nil {
		panic("nil key")
	}
	if !reflect.TypeOf(key).Comparable() {
		panic("key is not comparable")
	}
	return &valueCtx{parent, key, val}
}
```

访问时,会依次尝试local、parent（递归）

```
func (c *valueCtx) Value(key interface{}) interface{} {
	if c.key == key {
		return c.val
	}
	return c.Context.Value(key)
}
```