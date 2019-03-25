package gocon

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

// Worker concurrent thread control
type WorkerManager struct {
	// Pool who owns this worker.
	pool *Pool

	// Task is a job should be done.
	task chan func()

	Q chan os.Signal
}

// start is the start a thread.
func (w *WorkerManager) start(task func()) {
	go func() {
		w.run()
	}()

	w.task <- task
}

// run is the execute the task.
func (w *WorkerManager) run() {
	w.pool.incRunning()
	go func() {
		defer func() {
			if p := recover(); p != nil {
				w.pool.decRunning()
				log.Printf("Worker exits from a panic: %v", p)
			}
		}()

		for f := range w.task {
			if f == nil {
				w.pool.decRunning()
				return
			}

			f()

			// Revert worker to pool
			w.pool.revertWorker(w)
		}
	}()
}

func (w *WorkerManager) MakeRecvSignal() os.Signal {
	w.MakeSignal()

	// Block until a signal is received.
	return w.RecvSignal()
}

func (w *WorkerManager) RecvSignal() os.Signal {
	select {
	case s := <-w.Q:
		fmt.Println("custom recv signal: ", s)
		return s
	}
}

// MakeSignal Signal sent to c
func (w *WorkerManager) MakeSignal() {
	signal.Notify(w.Q,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		os.Interrupt,
		os.Kill,
	)
}
