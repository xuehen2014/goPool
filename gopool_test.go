package goPool

import (
	"github.com/daniel-hutao/spinlock"
	"sync"
	"testing"
	"time"
)

// go test -v -run TestGoPoolWithMutex
func TestGoPoolWithMutex(t *testing.T) {
	pool := NewGoPool(100, WithLock(new(sync.Mutex)))
	for i := 0; i < 1000; i++ {
		pool.AddTask(func() {
			time.Sleep(10 * time.Millisecond)
		})
	}

	pool.Release()
}

func TestGoPoolWithSpinLock(t *testing.T) {
	pool := NewGoPool(100, WithLock(new(spinlock.SpinLock)))
	for i := 0; i < 1000; i++ {
		pool.AddTask(func() {
			time.Sleep(10 * time.Millisecond)
		})
	}
	pool.Release()
}
