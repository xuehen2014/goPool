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
				result, err := w.executeTask(t, pool) // 执行任务
				w.handleResult(result, err, pool)     // 处理任务执行后返回的结果或者错误
			}
			//执行完成后, 将worker归位
			pool.pushWorker(workerIndex)
		}
	}()
}

func (w *worker) executeTask(t task, pool *goPool) (interface{}, error) {
	var result interface{}
	var err error
	for i := 0; i <= pool.retryCount; i++ {
		if pool.timeout > 0 { //如果有超时时间限制
			result, err = w.executeTaskWithTimeout(t, pool)
		} else {
			result, err = w.executeTaskWithoutTimeout(t, pool)
		}
		if err == nil || i == pool.retryCount {
			return result, err
		}
	}
	return result, err
}

// 有超时时间限制
func (w *worker) executeTaskWithTimeout(t task, pool *goPool) (interface{}, error) {
	var result interface{}
	var err error
	// 设置一个超时时间
	ctx, cancel := context.WithTimeout(context.Background(), pool.timeout)
	defer cancel()

	// 创建一个channel, 用来接收任务运行结果
	done := make(chan struct{})

	// 执行任务
	go func() {
		result, err = t()
		// r任务执行完成, 关闭done这个channel
		close(done)
	}()

	// 等待任务完成或者超时
	select {
	case <-done:
		// 收到任务完成信号, 返回任务执行的结果
		return result, err
	case <-ctx.Done():
		// 任务超时
		return nil, fmt.Errorf("Task timed out")
	}
}

// 没有超时时间限制
func (w *worker) executeTaskWithoutTimeout(t task, pool *goPool) (interface{}, error) {
	var result interface{}
	var err error
	result, err = t()
	return result, err
}

// 执行结果或者错误处理回调方法
func (w *worker) handleResult(result interface{}, err error, pool *goPool) {
	if err != nil && pool.errorCallback != nil {
		pool.errorCallback(err)
	} else if pool.resultCallback != nil {
		pool.resultCallback(result)
	}
}
