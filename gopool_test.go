package goPool

import (
	"testing"
	"time"
)

// go test -v -run TestGoPool *.go
func TestGoPool(t *testing.T) {
	pool := NewGoPool(100)
	for i := 0; i < 1000; i++ {
		pool.AddTask(func() {
			time.Sleep(10 * time.Millisecond)
		})
	}
	pool.Release()
}

// go test -v -bench=BenchmarkGoPool *.go
func BenchmarkGoPool(b *testing.B) {
	pool := NewGoPool(10000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pool.AddTask(func() {
			time.Sleep(10 * time.Millisecond)
		})
	}
	pool.Release()
}
