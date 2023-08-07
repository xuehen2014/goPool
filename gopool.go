package goPool

type Task func()

// GoPool结构体, 代表整个 Worker Pool
/*
* NewGoPool 方法会创建指定数量的 Worker，并启动它们
* AddTask 方法会将任务添加到任务队列中
* Release 方法会关闭任务队列，并等待所有的 Worker 完成当前的任务
 */
type GoPool struct {
	MaxWorkers  int
	Workers     []*Worker
	workerStack []int
}

func NewGoPool(maxWorkers int) *GoPool {
	pool := &GoPool{
		MaxWorkers:  maxWorkers,
		Workers:     make([]*Worker, maxWorkers),
		workerStack: make([]int, maxWorkers),
	}
	for i := 0; i < maxWorkers; i++ {
		worker := newWorker()
		pool.Workers[i] = worker
		pool.workerStack[i] = i
		worker.start(pool, i)
	}
	return pool
}

func (p *GoPool) AddTask(task Task) {
	workerIndex := p.popWorker()
	p.Workers[workerIndex].TaskQueue <- task
}

func (p *GoPool) Release() {
	for _, worker := range p.Workers {
		close(worker.TaskQueue)
	}
}

func (p *GoPool) popWorker() int {
	workerIndex := p.workerStack[len(p.workerStack)-1]
	p.workerStack = p.workerStack[:len(p.workerStack)-1]
	return workerIndex
}

func (p *GoPool) pushWorker(workerIndex int) {
	p.workerStack = append(p.workerStack, workerIndex)
}
