package goPool

import (
	//"fmt"
	// "github.com/davecgh/go-spew/spew"
	"sync"
	"time"
)

//worker中要执行的方法
type task func()

// GoPool结构体, 代表整个 worker Pool
/*
* NewGoPool 方法会创建指定数量的 worker
* AddTask 方法会将任务添加到任务队列中
* Release 方法会关闭任务队列，并等待所有的 worker 完成当前的任务
 */
type goPool struct {
	workers     []*worker
	maxWorkers  int
	minWorkers  int
	workerStack []int
	taskQueue   chan task
	lock        sync.Locker
	cond        *sync.Cond
	timeout     time.Duration
}

func NewGoPool(maxWorkers int, opts ...Option) *goPool {
	pool := &goPool{
		maxWorkers:  maxWorkers,
		minWorkers:  maxWorkers, // Set minWorkers to maxWorkers by default
		workers:     make([]*worker, maxWorkers),
		workerStack: make([]int, maxWorkers),
		taskQueue:   make(chan task, 1e6),
		lock:        new(sync.Mutex),
		timeout:     0,
	}

	for _, opt := range opts {
		opt(pool)
	}
	if pool.cond == nil {
		pool.cond = sync.NewCond(pool.lock)
	}

	for i := 0; i < pool.minWorkers; i++ {
		worker := newWorker()
		pool.workers[i] = worker
		pool.workerStack[i] = i
		worker.start(pool, i)
	}
	go pool.dispatch()
	return pool
}

func (p *goPool) AddTask(t task) {
	p.taskQueue <- t
}

func (p *goPool) Release() {
	close(p.taskQueue)
	p.cond.L.Lock()
	for len(p.workerStack) != p.maxWorkers {
		p.cond.Wait()
	}
	p.cond.L.Unlock()
	for _, worker := range p.workers {
		close(worker.taskQueue)
	}
	p.workers = nil
	p.workerStack = nil
}

func (p *goPool) popWorker() int {
	p.cond.L.Lock()
	workerIndex := p.workerStack[len(p.workerStack)-1]
	p.workerStack = p.workerStack[:len(p.workerStack)-1]
	p.cond.L.Unlock()
	return workerIndex
}

func (p *goPool) pushWorker(workerIndex int) {
	p.cond.L.Lock()
	p.workerStack = append(p.workerStack, workerIndex)
	p.cond.L.Unlock()
	p.cond.Signal()
}

func (p *goPool) dispatch() {
	for t := range p.taskQueue {
		p.cond.L.Lock()
		for len(p.workerStack) == 0 {
			p.cond.Wait()
		}
		p.cond.L.Unlock()
		workerIndex := p.popWorker()
		p.workers[workerIndex].TaskQueue <- t
		// 动态调整
		if len(p.taskQueue) > (p.maxWorkers-p.minWorkers)/2+p.minWorkers && len(p.workerStack) < p.maxWorkers {
			worker := newWorker()
			p.workers = append(p.workersworkers, worker)
			p.workerStack = append(p.workerStack, len(p.workers)-1)
			worker.start(p, len(p.workers)-1)
		} else if len(p.taskQueue) < p.minWorkers && len(p.workerStack) > p.minWorkers {
			p.workers = p.workers[:len(p.workers)-1]
			p.workerStack = p.workerStack[:len(p.workerStack)-1]
		}
	}
}
