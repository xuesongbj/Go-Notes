package syncTest 

import (
    "sync"
)

type Lock struct{
    ch chan struct{}
}

func NewLock() *Lock {
    return &Lock{
        ch: make(chan struct{}, 1),
    }
}

func (t *Lock) Lock() {
    <-t.ch
}

func (t *Lock) Unlock() {
    t.ch <- struct{}{}
}

func addUseChan(c *int) {
    *c += 1
}

func UseChan() {
    var (
        c int
        wg sync.WaitGroup
    )

    var t = NewLock()
    t.ch <- struct{}{}

    for i := 0; i < 5; i++ {
        wg.Add(1)
        go func(c *int) {
            defer wg.Done()       

            t.Lock()
            defer t.Unlock()
            // add(c)
            addUseChan(c)
        }(&c)
    }
    wg.Wait()
}
