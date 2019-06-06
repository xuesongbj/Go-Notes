package syncTest 

import (
    "sync"
)

func adduseMutex(c *int, wg *sync.WaitGroup, m *sync.Mutex) {
    defer wg.Done() 

    m.Lock()
    defer m.Unlock()

    *c += 1
}

func UseMutex() {
    var (
        wg sync.WaitGroup
        m  sync.Mutex
        c int
    )

    for i := 0; i < 5; i++ {
        wg.Add(1)
        go adduseMutex(&c, &wg, &m)        
    }

    wg.Wait()
}
