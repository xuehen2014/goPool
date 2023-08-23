package goPool

import (
	"sync"
	"time"
)

type Option func(*goPool)

func WithLock(lock sync.Locker) Option {
	return func(p *goPool) {
		p.lock = lock
		p.cond = sync.NewCond(p.lock)
	}
}

func WithMinWorkers(minWorkers int) Option {
	return func(p *goPool) {
		p.minWorkers = minWorkers
	}
}

// 设置超时时间
func WithTimeout(timeout time.Duration) Option {
	return func(p *goPool) {
		p.timeout = timeout
	}
}

// 结果回调方法
func WithResultCallback(callback func(interface{})) Option {
	return func(p *goPool) {
		p.resultCallback = callback
	}
}

// 错误回调方法
func WithErrorCallback(callback func(error)) Option {
	return func(p *goPool) {
		p.errorCallback = callback
	}
}

// 设置充值次数
func WithRetryCount(retryCount int) Option {
	return func(p *goPool) {
		p.retryCount = retryCount
	}
}

// 设置任务通道缓冲区长度
func WithTaskQueueSize(size int) Option {
	return func(p *goPool) {
		p.taskQueueSize = size
	}
}
