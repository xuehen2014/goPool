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
				var result interface{}
				var err error

				if pool.timeout > 0 {
					ctx, cancel := context.WithTimeout()
					defer cancel()

					// create a channel to receive the result of the task
					done := make(chan struct{})

					go func() {
						result, err := t()
						close(done)
					}()

					select {
					case <-done: // 任务没有超时
						if err != nil && pool.errorCallback != nil {
							pool.errorCallback(err)
						} else if pool.resultCallback != nil {
							pool.resultCallback(result)
						}
					case <-ctx.Done(): // 任务超时
						if pool.errorCallback != nil {
							pool.errorCallback(fmt.Errorf("Task time out"))
						}
					}
				}
			} else { // 没有超时时间限制
				result, err = t()
				if err != nil && pool.errorCallback != nil {
					pool.errorCallback(err)
				} else if pool.resultCallback != nil {
					pool.resultCallback(result)
				}
			}
			pool.pushWorker(workerIndex)
		}
	}()
}
