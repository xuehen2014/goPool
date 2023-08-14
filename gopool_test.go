package goPool

import (
	"fmt"
	"github.com/daniel-hutao/spinlock"
	"sync"
	"testing"
	"time"
)

// go test -v -run TestGoPoolWithMutex
func TestGoPoolWithMutex(t *testing.T) {
	fmt.Println("aaaaaaaa")
	wg := &sync.WaitGroup{}
	pool := NewGoPool(100, WithLock(new(sync.Mutex)))
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		pool.AddTask(func() {
			time.Sleep(10 * time.Millisecond)
			wg.Done()
		})
	}
	wg.Wait()
	pool.Release()
}

// go test -v -run TestGoPoolWithSpinLock
func TestGoPoolWithSpinLock(t *testing.T) {
	wg := &sync.WaitGroup{}
	pool := NewGoPool(100, WithLock(new(spinlock.SpinLock)))
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		pool.AddTask(func() {
			time.Sleep(10 * time.Millisecond)
			wg.Done()
		})
	}
	wg.Wait()
	pool.Release()
}

// go test -benchmem -run=^$ -bench ^BenchmarkGoPoolWithMutex$ .
func BenchmarkGoPoolWithMutex(b *testing.B) {
	wg := &sync.WaitGroup{}
	var taskNum = int(1e6)
	pool := NewGoPool(5e4, WithLock(new(sync.Mutex)))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(taskNum)
		for num := 0; num < taskNum; num++ {
			pool.AddTask(func() {
				time.Sleep(10 * time.Millisecond)
				wg.Done()
			})
		}
	}
	wg.Wait()
	b.StopTimer()
	pool.Release()
}

// go test -benchmem -run=^$ -bench ^BenchmarkGoPoolWithSpinLock$ .
func BenchmarkGoPoolWithSpinLock(b *testing.B) {
	var wg sync.WaitGroup
	var taskNum = int(1e6)
	pool := NewGoPool(5e4, WithLock(new(spinlock.SpinLock)))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(taskNum)
		for num := 0; num < taskNum; num++ {
			pool.AddTask(func() {
				time.Sleep(10 * time.Millisecond)
				wg.Done()
			})
		}
	}
	wg.Wait()
	b.StopTimer()
	pool.Release()
}

// go test -benchmem -run=^$ -bench ^BenchmarkGoroutines$ .
func BenchmarkGoroutines(b *testing.B) {
	var wg sync.WaitGroup
	var taskNum = int(1e6)

	for i := 0; i < b.N; i++ {
		wg.Add(taskNum)
		for num := 0; num < taskNum; num++ {
			go func() {
				time.Sleep(10 * time.Millisecond)
				wg.Done()
			}()
		}
	}
}
