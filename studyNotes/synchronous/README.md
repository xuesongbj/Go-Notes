# 信号量、互斥锁、条件变量

### 信号量、互斥锁概念
信号量用在多线程或多任务同步的,一个线程完成了某一个动作就通过信号量告诉别的线程,别的线程再进行某些动作，别的线程再进行某些动作(大家都在semtake的时候,就会阻塞在那里)。

互斥锁是用在多线程多任务互斥的,一个线程占用了某一个资源,那么别的线程就无法访问,直到这个线程unlock,其它的线程才可以利用这个资源。比如,对于全局变量的访问,有时要加锁,操作完了,再解锁。

有的时候锁和信号量会同时使用的。也就是说，信号量不一定是锁定某一个资源,而是流程上的概念。比如:有A、B两个线程，B线程要等A线程完成某一个任务以后再进行自己下面的步骤,这个任务并不一定是锁住某一个资源,还可以是进行一些计算或者数据处理之类;而线程互斥锁则是“锁住某一资源”的概念,在锁定期内,其它线程无法对被保护的数据进行操作。


## 信号量、互斥锁区别

### 作用域
* 信号量: 进程间或线程间(Linux仅线程间)
* 互斥锁: 线程间

### 上锁
* 信号量: 只要信号量的value大于0,其它线程就可以sem\_wait成功,成功后信号量的value减1。若value值不大于0,则sem_wait阻塞,直到sem_post释放后value值加1。
* 互斥锁: 只要被锁住,其它任何线程都不可以访问被保护的资源。

### 概念区别
信号量(灯)与互斥锁、条件变量的主要不同在于"灯"的概念,灯亮则意味着不可用。如果说后两种同步方式侧重于"等待"操作，即资源不可用的话，信号灯机制则偏重于点灯，即告诉资源可用;没有等待线程的解锁或激发条件都没有意义的，而没有等待灯亮线程的点灯操作则有效，且能保持灯亮状态。当然，这样的操作原语也意味着更多开销。


## 实例

### 互斥锁
在Go语言中可使用sync.Metux实现互斥锁.

```
import "sync"

func Exmaple() {
	var m sync.Mutex
	var wg sync.WaitGroup
	
	data = make([]int, 0, 10)
	
	for i := 0; i < 10; i++{
		wg.Add(1)
		go func(wg *sync.WaitGroup, m *sync.Mutex, data *[]int, i int) {
			defer wg.Done()
			
			// 互斥锁
			// 同时仅允许一个线程访问data资源
			m.Lock()
			*data = append(*data, i)
			m.Unlock()
		}(&wg, &m, *data, i)
	}
}
```

### 信号量
在Go语言中可以使用chan实现信号通知。

```
func Example(maxWorker int, data []string) {
	ch := make([]struct{}, maxWorker)
	
	// 初始化,开灯
	for i := 0; i < maxWorker; i++ {
		ch <- struct{}{}
	}
	
	done := make(chan bool)
	waitAllJobs := make(chan bool)
	
	go func(){
		// data具体消费的任务
		for i := 0; i < len(data); i++ {
			<-done
			
			// 开灯
			ch <- struct{}{}
		}
		waitAllJobs <- true
	}()
	
	for i := 0; i < len(data); i++ {
		// 关灯
		<-ch
		
		go func() {
			// TODO
			
			// 任务完成之后,发送完成信号
			// 开灯
			done <- true
		}()
	}
	
	<-waitAllJobs
}
```

#### semaphore
除了使用chan实现信号通知,也可以使用golang扩展库提供的"golang.org/x/sync/semaphore"库进行实现。

```
package main

import (
    "context"
    "fmt"
    "log"
    "runtime"

    "golang.org/x/sync/semaphore"
)

// Example_workerPool demonstrates how to use a semaphore to limit the number of
// goroutines working on parallel tasks.
//
// This use of a semaphore mimics a typical “worker pool” pattern, but without
// the need to explicitly shut down idle workers when the work is done.
func main() {
    ctx := context.TODO()

    var (
        maxWorkers = runtime.GOMAXPROCS(0)
        sem        = semaphore.NewWeighted(int64(maxWorkers))
        out        = make([]int, 32)
    )

    // Compute the output using up to maxWorkers goroutines at a time.
    for i := range out {
        // When maxWorkers goroutines are in flight, Acquire blocks until one of the
        // workers finishes.
        if err := sem.Acquire(ctx, 1); err != nil {
            log.Printf("Failed to acquire semaphore: %v", err)
            break
        }

        go func(i int) {
            defer sem.Release(1)
            out[i] = collatzSteps(i + 1)
        }(i)
    }

    // Acquire all of the tokens to wait for any remaining workers to finish.
    //
    // If you are already waiting for the workers by some other means (such as an
    // errgroup.Group), you can omit this final Acquire call.
    if err := sem.Acquire(ctx, int64(maxWorkers)); err != nil {
        log.Printf("Failed to acquire semaphore: %v", err)
    }

    fmt.Println(out)

}

// collatzSteps computes the number of steps to reach 1 under the Collatz
// conjecture. (See https://en.wikipedia.org/wiki/Collatz_conjecture.)
func collatzSteps(n int) (steps int) {
    if n <= 0 {
        panic("nonpositive input")
    }

    for ; n > 1; steps++ {
        if steps < 0 {
            panic("too many steps")
        }

        if n%2 == 0 {
            n /= 2
            continue
        }

        const maxInt = int(^uint(0) >> 1)
        if n > (maxInt-1)/3 {
            panic("overflow")
        }
        n = 3*n + 1
    }

    return steps
}
```
