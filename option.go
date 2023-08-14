package goPool

import (
	"sync"
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
