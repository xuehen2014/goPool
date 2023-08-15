package goPool

import (
	"errors"
	"github.com/daniel-hutao/spinlock"
	"sync"
	"testing"
	"time"
)

// go test -v -run TestGoPoolWithMutex
func TestGoPoolWithMutex(t *testing.T) {
	wg := &sync.WaitGroup{}
	pool := NewGoPool(100, WithLock(new(sync.Mutex)))
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		pool.AddTask(func() (interface{}, error) {
			time.Sleep(10 * time.Millisecond)
			return nil, nil
		})
	}
	wg.Wait()
	pool.Release()
}

// go test -v -run TestGoPoolWithSpinLock
func TestGoPoolWithSpinLock(t *testing.T) {
	pool := NewGoPool(100, WithLock(new(spinlock.SpinLock)))
	for i := 0; i < 1000; i++ {
		pool.AddTask(func() (interface{}, error) {
			time.Sleep(10 * time.Millisecond)
			return nil, nil
		})
	}
	pool.Release()
}

// go test -benchmem -run=^$ -bench ^BenchmarkGoPoolWithMutex$ .
func BenchmarkGoPoolWithMutex(b *testing.B) {
	var wg sync.WaitGroup
	var taskNum = int(1e6)
	pool := NewGoPool(1e4, WithLock(new(sync.Mutex)))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(taskNum)
		for num := 0; num < taskNum; num++ {
			pool.AddTask(func() (interface{}, error) {
				time.Sleep(10 * time.Millisecond)
				wg.Done()
				return nil, nil
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
	pool := NewGoPool(1e4, WithLock(new(spinlock.SpinLock)))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(taskNum)
		for num := 0; num < taskNum; num++ {
			pool.AddTask(func() (interface{}, error) {
				time.Sleep(10 * time.Millisecond)
				wg.Done()
				return nil, nil
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
			go func() (interface{}, error) {
				time.Sleep(10 * time.Millisecond)
				wg.Done()
				return nil, nil
			}()
		}
	}
}

func TestGoPoolWithError(t *testing.T) {
	var errTaskError = errors.New("task error")
	pool := NewGoPool(100, WithErrorCallback(func(err error) {
		if err != errTaskError {
			t.Errorf("Expected error %v, but got %v", errTaskError, err)
		}
	}))
	for i := 0; i < 1000; i++ {
		pool.AddTask(func() (interface{}, error) {
			return nil, errTaskError
		})
	}
	pool.Release()
}

func TestGoPoolWithResult(t *testing.T) {
	var expectedResult = "task result"
	pool := NewGoPool(100, WithResultCallback(func(result interface{}) {
		if result != expectedResult {
			t.Errorf("Expected result %v, but got %v", expectedResult, result)
		}
	}))
	for i := 0; i < 1000; i++ {
		pool.AddTask(func() (interface{}, error) {
			return expectedResult, nil
		})
	}
	pool.Release()
}
