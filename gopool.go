package goPool

type Task func()

// GoPool结构体, 代表整个 Worker Pool
/*
* NewGoPool 方法会创建指定数量的 Worker，并启动它们
* AddTask 方法会将任务添加到任务队列中
* Release 方法会关闭任务队列，并等待所有的 Worker 完成当前的任务
 */
type GoPool struct {
	TaskQueue  chan Task
	MaxWorkers int
	Workers    []*Worker
}

func NewGoPool(maxWorkers int) *GoPool {
	pool := &GoPool{
		TaskQueue:  make(chan Task),
		MaxWorkers: maxWorkers,
		Workers:    make([]*Worker, maxWorkers),
	}
	for i := 0; i < maxWorkers; i++ {
		worker := newWorker(pool.TaskQueue)
		pool.Workers[i] = worker
		worker.start()
	}
	return pool
}

func (p *GoPool) AddTask(task Task) {
	p.TaskQueue <- task
}

func (p *GoPool) Release() {
	close(p.TaskQueue)
	for _, worker := range p.Workers {
		<-worker.TaskQueue
	}
}
