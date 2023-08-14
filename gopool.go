package goPool

import (
	//"fmt"
	// "github.com/davecgh/go-spew/spew"
	"sync"
)

type Task func()

// GoPool结构体, 代表整个 Worker Pool
/*
* NewGoPool 方法会创建指定数量的 Worker，并启动它们
* AddTask 方法会将任务添加到任务队列中
* Release 方法会关闭任务队列，并等待所有的 Worker 完成当前的任务
 */
type GoPool struct {
	MaxWorkers  int
	MinWorkers  int
	Workers     []*Worker
	workerStack []int
	taskQueue   chan Task
	lock        sync.Locker
	cond        *sync.Cond
}

type Option func(*GoPool)

func WithLock(lock sync.Locker) Option {
	return func(p *GoPool) {
		p.lock = lock
		p.cond = sync.NewCond(p.lock)
	}
}

func WithMinWorkers(minWorkers int) Option {
	return func(p *GoPool) {
		p.MinWorkers = minWorkers
	}
}

func NewGoPool(maxWorkers int, opts ...Option) *GoPool {
	pool := &GoPool{
		MaxWorkers:  maxWorkers,
		MinWorkers:  maxWorkers, // Set MinWorkers to MaxWorkers by default
		Workers:     make([]*Worker, maxWorkers),
		workerStack: make([]int, maxWorkers),
		taskQueue:   make(chan Task, 1e6),
		lock:        new(sync.Mutex),
	}

	for _, opt := range opts {
		opt(pool)
	}
	if pool.cond == nil {
		pool.cond = sync.NewCond(pool.lock)
	}

	for i := 0; i < pool.MinWorkers; i++ {
		worker := newWorker()
		pool.Workers[i] = worker
		pool.workerStack[i] = i
		worker.start(pool, i)
	}
	go pool.dispatch()
	return pool
}

func (p *GoPool) AddTask(task Task) {
	p.taskQueue <- task
}

func (p *GoPool) Release() {
	close(p.taskQueue)
	p.cond.L.Lock()
	for len(p.workerStack) != p.MaxWorkers {
		p.cond.Wait()
	}
	p.cond.L.Unlock()
	for _, worker := range p.Workers {
		close(worker.TaskQueue)
	}
	p.Workers = nil
	p.workerStack = nil
}

func (p *GoPool) popWorker() int {
	p.cond.L.Lock()
	workerIndex := p.workerStack[len(p.workerStack)-1]
	p.workerStack = p.workerStack[:len(p.workerStack)-1]
	p.cond.L.Unlock()
	return workerIndex
}

func (p *GoPool) pushWorker(workerIndex int) {
	p.cond.L.Lock()
	p.workerStack = append(p.workerStack, workerIndex)
	p.cond.L.Unlock()
	p.cond.Signal()
}

func (p *GoPool) dispatch() {
	for task := range p.taskQueue {
		p.cond.L.Lock()
		for len(p.workerStack) == 0 {
			p.cond.Wait()
		}
		p.cond.L.Unlock()
		workerIndex := p.popWorker()
		p.Workers[workerIndex].TaskQueue <- task
		// 动态调整
		if len(p.taskQueue) > (p.MaxWorkers-p.MinWorkers)/2+p.MinWorkers && len(p.workerStack) < p.MaxWorkers {
			worker := newWorker()
			p.Workers = append(p.Workers, worker)
			p.workerStack = append(p.workerStack, len(p.Workers)-1)
			worker.start(p, len(p.Workers)-1)
		} else if len(p.taskQueue) < p.MinWorkers && len(p.workerStack) > p.MinWorkers {
			p.Workers = p.Workers[:len(p.Workers)-1]
			p.workerStack = p.workerStack[:len(p.workerStack)-1]
		}
	}
}
