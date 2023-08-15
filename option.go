package goPool

import (
	"sync"
	"time"
)

type Option func(*GoPool)

func WithLock(lock sync.Locker) Option {
	return func(p *GoPool) {
		p.lock = lock
		p.cond = sync.NewCond(p.lock)
	}
}

func WithMinWorkers(minWorkers int) Option {
	return func(p *GoPool) {
		p.minWorkers = minWorkers
	}
}

// 设置超时时间
func WithTimeout(timeout time.Duration) Option {
	return func(p *goPool) {
		p.timeout = timeout
	}
}
