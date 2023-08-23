package goPool

import (
	//"fmt"
	// "github.com/davecgh/go-spew/spew"
	"context"
	"sync"
	"time"
)

type GoPool interface {
	// 添加任务到池子里
	AddTask(t task)
	// 等待任务分发完成
	Wait()
	// 释放资源
	Release()
	// 返回正在运行中的worker数量
	Running() int
	// 返回全部worker的数量
	GetWorkerCount() int
	// 返回通道缓冲区长度
	GetTaskQueueSize() int
}

//worker中要执行的方法
type task func() (interface{}, error)

// GoPool结构体, 代表整个 worker Pool
/*
* NewGoPool 方法会创建指定数量的 worker
* AddTask 方法会将任务添加到任务队列中
* Release 方法会关闭任务队列，并等待所有的 worker 完成当前的任务
 */
type goPool struct {
	workers        []*worker // 所有已创建的worker列表
	workerStack    []int     // 所有已创建的worker的编号列表
	maxWorkers     int       // 最大worker数量
	minWorkers     int       // 最小worker数量, 可以被WithMinWorkers() 指定, 默认等于maxWorkers
	taskQueue      chan task // 任务先添加到这个channel, 然后通过dispatch() 方法分发到每个worker自己的channel中
	taskQueueSize  int       // taskQueue缓冲区大小, 可以被WithTaskQueueSize() 指定
	retryCount     int       // 任务重试次数
	lock           sync.Locker
	cond           *sync.Cond
	timeout        time.Duration     // task执行超时时间, 可以被WithTimeout()指定
	resultCallback func(interface{}) //task执行后结果的回调处理方法
	errorCallback  func(error)       // task执行后返回错误的回调处理方法
	adjustInterval time.Duration     // 调整worker数量的时间间隔
	ctx            context.Context
	cancel         context.CancelFunc
}

func NewGoPool(maxWorkers int, opts ...Option) *goPool {
	ctx, cancel := context.WithCancel(context.Background())
	pool := &goPool{
		maxWorkers:     maxWorkers,
		minWorkers:     maxWorkers,
		workers:        nil,
		workerStack:    nil,
		taskQueue:      nil,
		taskQueueSize:  10000,
		retryCount:     0,
		lock:           new(sync.Mutex),
		timeout:        0,
		adjustInterval: 1 * time.Second,
		ctx:            ctx,
		cancel:         cancel,
	}

	for _, opt := range opts {
		opt(pool)
	}

	pool.taskQueue = make(chan task, pool.taskQueueSize)
	pool.workers = make([]*worker, pool.minWorkers)
	pool.workerStack = make([]int, pool.minWorkers)

	if pool.cond == nil {
		pool.cond = sync.NewCond(pool.lock)
	}

	//创建worker
	for i := 0; i < pool.minWorkers; i++ {
		worker := newWorker()
		pool.workers[i] = worker
		pool.workerStack[i] = i
		worker.start(pool, i)
	}
	go pool.adjustWorkers()
	go pool.dispatch()
	return pool
}

func (p *goPool) AddTask(t task) {
	p.taskQueue <- t
}

// 等待所有任务发布以及执行完成
func (p *goPool) Wait() {
	for {
		p.lock.Lock()
		workerStackLen := len(p.workerStack)
		p.lock.Unlock()
		if len(p.taskQueue) == 0 && workerStackLen == len(p.workers) {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func (p *goPool) Release() {
	close(p.taskQueue)
	p.cancel()
	p.cond.L.Lock()
	for len(p.workerStack) != p.minWorkers { //等待worker全部空闲
		p.cond.Wait()
	}
	p.cond.L.Unlock()
	// 关闭worker里的channel
	for _, worker := range p.workers {
		close(worker.taskQueue)
	}
	p.workers = nil
	p.workerStack = nil
}

// 从空闲worker列表里取出一个, 用于dispatch任务
func (p *goPool) popWorker() int {
	p.cond.L.Lock()
	workerIndex := p.workerStack[len(p.workerStack)-1]
	p.workerStack = p.workerStack[:len(p.workerStack)-1]
	p.cond.L.Unlock()
	return workerIndex
}

// worker的任务执行完成后, 归还worker到空闲列表
func (p *goPool) pushWorker(workerIndex int) {
	p.cond.L.Lock()
	p.workerStack = append(p.workerStack, workerIndex)
	p.cond.L.Unlock()
	p.cond.Signal()
}

// 从worker空闲列表中取出worker, 派发任务
func (p *goPool) dispatch() {
	for t := range p.taskQueue {
		p.cond.L.Lock()
		for len(p.workerStack) == 0 {
			p.cond.Wait()
		}
		p.cond.L.Unlock()
		workerIndex := p.popWorker()
		p.workers[workerIndex].taskQueue <- t
	}
}

// 动态调整worker的数量
func (p *goPool) adjustWorkers() {
	ticker := time.NewTicker(p.adjustInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.cond.L.Lock()
			if len(p.taskQueue) > len(p.workers)*3/4 && len(p.workers) < p.maxWorkers {
				newWorkers := min(len(p.workers)*2, p.maxWorkers) - len(p.workers)
				for i := 0; i < newWorkers; i++ {
					worker := newWorker()
					p.workers = append(p.workers, worker)
					p.workerStack = append(p.workerStack, len(p.workers)-1)
					worker.start(p, len(p.workers)-1)
				}
			} else if len(p.taskQueue) == 0 && len(p.workers) > p.minWorkers {
				removeWorkers := max((len(p.workers)-p.minWorkers)/2, p.minWorkers)
				p.workers = p.workers[:len(p.workers)-removeWorkers]
				p.workerStack = p.workerStack[:len(p.workerStack)-removeWorkers]
			}
			p.cond.L.Unlock()
		case <-p.ctx.Done():
			return
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
