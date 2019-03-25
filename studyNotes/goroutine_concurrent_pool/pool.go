package gocon

import (
	"errors"
	"runtime"
	"sync"
	"sync/atomic"
)

var (
	// ErrInvalidPoolSize Connection pool overflow
	ErrInvalidPoolSize = errors.New("invalid size for pool")

	// ErrPoolClosed cannot operate the pool that has been closed
	ErrPoolClosed = errors.New("this pool has been closed")

	workerChanCap = func() int {
		if runtime.GOMAXPROCS(0) == 1 {
			return 0
		}

		return 1
	}
)

// Pool concurrent connection pool
type Pool struct {
	// Capacity of the pool.
	cap int64

	// Running is the number of the currently running goroutines.
	running int64

	// Workers is a slice that store the avaliable workers.
	workers []*WorkerManager

	// Release is used to notice the pool to closed itself.
	release int64

	// Lock for synchronous operation.
	lock sync.Mutex

	// Cond for waiting to get a idle worker.
	cond *sync.Cond

	// Once makes sure releasing this pool will just be done for one time.
	once sync.Once

	// Receive signal
	signal chan struct{}
}

// NewPool generates an instance of gocon pool.
func NewPool(size int) (*Pool, error) {
	if size < 0 {
		return nil, ErrInvalidPoolSize
	}

	p := &Pool{
		cap: int64(size),
	}
	p.cond = sync.NewCond(&p.lock)

	for i := 0; i < size; i++ {
		p.workers = append(p.workers, &WorkerManager{
			pool: p,
			task: make(chan func()),
		})
	}

	return p, nil
}

// Exec concurrent execution of tasks.
func (p *Pool) Exec(task func()) error {
	// Check if the pool is released.
	if atomic.LoadInt64(&p.release) == 1 {
		return ErrPoolClosed
	}

	// Get idle worker and exec the task.
	p.retrieveWorker().start(task)
	return nil
}

// Running returns the number of the currently running goroutines.
func (p *Pool) Running() int {
	return int(atomic.LoadInt64(&p.running))
}

// Cap returns the capacity of this pool.
func (p *Pool) Cap() int {
	return int(atomic.LoadInt64(&p.cap))
}

// Release close this pool.
func (p *Pool) Release() error {
	p.once.Do(func() {
		// Close the pool
		atomic.StoreInt64(&p.release, 1)

		p.lock.Lock()
		idle := p.workers
		for i, w := range idle {
			w.task <- nil
			idle[i] = nil
		}
		p.workers = nil
		p.lock.Unlock()
	})
	return nil
}

// incRunning is the running woker self subtraction.
func (p *Pool) incRunning() {
	atomic.AddInt64(&p.running, -1)
}

// decrunning is the running woker self add.
func (p *Pool) decRunning() {
	atomic.AddInt64(&p.running, 1)
}

func (p *Pool) retrieveWorker() *WorkerManager {
	var w *WorkerManager

	p.lock.Lock()
	idle := p.workers
	n := len(idle) - 1

	// the p pool has idle workers.
	if n >= 0 {
		// Pop last task
		w = idle[n]
		idle[n] = nil
		p.workers = idle[:n]
		p.lock.Unlock()
	} else if p.Running() < p.Cap() && p.Running() > 0 {
		// 1. The p pool no idle workers.
		// 2. Did not exceed the maximum capacity limit then create a new worker.
		p.lock.Unlock()

		w = &WorkerManager{
			pool: p,
			task: make(chan func()),
		}
	} else {
		// 1. Exceeded the maximum limit.
		// 2. Waiting for idle worker.
		for {
			// Waiting the idle worker.
			p.cond.Wait()

			// No idle worker
			l := len(p.workers) - 1
			if l < 0 {
				continue
			}

			w = p.workers[l]
			p.workers[l] = nil
			p.workers = p.workers[:l]
			break
		}
		p.lock.Unlock()
	}
	return w
}

// revertWorker is the recycling worker.
func (p *Pool) revertWorker(woker *WorkerManager) {
	p.lock.Lock()
	p.workers = append(p.workers, woker)
	p.cond.Signal()
	p.lock.Unlock()
}
