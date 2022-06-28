package main

import (
	"fmt"
	"os"

	"github.com/xuesongbj/gocon"
)

var (
	// DefaultPoolSize is the default goroutine concurrent pool size.
	DefaultPoolSize = 5
)

func initGoConcurrentPool() (*gocon.Pool, error) {
	P, err := gocon.NewPool(DefaultPoolSize)
	return P, err
}

// TestTask is the test tast.
func TestTask() {
	fmt.Println("Test task.")
}

func main() {
	p, err := initGoConcurrentPool()
	if err != nil {
		fmt.Printf("initGoConcurrentPool failed, err: %s", err.Error())
		os.Exit(99)
	}

	for i := 0; i < 10; i++ {
		// wg.Add(1)
		// defaultAntsPool.Exec(f1)
		p.Exec(TestTask)
	}

	// wg.Wait()

	fmt.Println("End...")
}
