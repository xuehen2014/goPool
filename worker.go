package goPool

import (
	"context"
	"fmt"
)

// 代表一个工作协成
type worker struct {
	taskQueue chan task
}

func newWorker() *worker {
	return &worker{
		taskQueue: make(chan task, 1),
	}
}

func (w *worker) start(pool *goPool, workerIndex int) {
	go func() {
		for t := range w.taskQueue {
			if t != nil {
				if pool.timeout > 0 {
					ctx, cancel := context.WithTimeout()
					defer cancel()

					// create a channel to receive the result of the task
					done := make(chan struct{})

					go func() {
						t()
						close(done)
					}()

					select {
					case <-done:
					case <-ctx.Done():
						fmt.Println("Task time out")
					}
				}
			} else { // 没有超时时间限制
				t()
			}
			pool.pushWorker(workerIndex)
		}
	}()
}
