package mgpool

import (
	"errors"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

const DefaultExpire = 5

var (
	ErrorInValidCap    = errors.New("pool cap can not <= 0")
	ErrorInValidExpire = errors.New("pool expire can not <= 0")
	ErrorPoolHasClosed = errors.New("pool has bean released")
)

type Pool struct {
	workers      []*Worker     // 空闲worker
	cap          int32         // pool max cap
	running      int32         // 正在运行的worker数量
	expire       time.Duration // 过期时间，空闲的worker超过这个时间回收掉
	release      chan sig      // 释放资源，pool就不能使用了
	lock         sync.Mutex    // 保护pool里相关资源的安全
	once         sync.Once     // 释放只能调用一次，不能多次调用
	workerCache  sync.Pool     //workerCache 缓存
	cond         *sync.Cond
	PanicHandler func()
}

type sig struct {
}

func NewPool(cap int) (*Pool, error) {
	return NewTimePool(cap, DefaultExpire)
}

func NewTimePool(cap int, expire int) (*Pool, error) {
	if cap <= 0 {
		return nil, ErrorInValidCap
	}
	if expire <= 0 {
		return nil, ErrorInValidExpire
	}
	p := &Pool{
		cap:     int32(cap),
		expire:  time.Duration(expire) * time.Second,
		release: make(chan sig, 1),
		lock:    sync.Mutex{},
	}
	p.workerCache.New = func() any {
		return &Worker{
			pool: p,
			task: make(chan func(), 1),
		}
	}
	p.cond = sync.NewCond(&p.lock)
	go p.clearExpireWorker() //定时清理过期的空闲worker
	return p, nil
}

func (p *Pool) incRunning() {
	atomic.AddInt32(&p.running, 1)
}

func (p *Pool) decRunning() {
	atomic.AddInt32(&p.running, -1)
}

func (p *Pool) PutWorker(w *Worker) {
	p.lock.Lock()
	w.lastTime = time.Now()
	w.pool.workerCache.Put(w)
	p.workers = append(p.workers, w)
	p.lock.Unlock()
	p.cond.Signal() // 通知goroutine已经有worker放回pool中
	w.pool.decRunning()
}

// Submit 提交任务
func (p *Pool) Submit(task func()) error {
	if len(p.release) > 0 {
		return ErrorPoolHasClosed
	}
	// 从pool中获取一个worker，然后执行任务
	w := p.GetWorker()
	w.task <- task
	w.run()
	return nil
}

func (p *Pool) GetWorker() (w *Worker) {
	p.lock.Lock()
	idleWorkers := p.workers
	n := len(idleWorkers) - 1
	if n >= 0 { // 说明存在空闲的worker
		w = idleWorkers[n]   // 取最后一个worker
		idleWorkers[n] = nil // 置空，防止内存泄漏
		p.workers = idleWorkers[:n]
		p.lock.Unlock()
		return
	}
	// 如果没有空闲的worker，要新建这个worker
	// 判断一下正在运行worker数量，如果小于cap容量，新建一个
	if p.running < p.cap {
		p.lock.Unlock()
		w = p.workerCache.Get().(*Worker)
		w.task = make(chan func(), 1)
		return
	}
	// 如果大于等于cap容量，阻塞等待worker释放
	p.lock.Unlock()
	return p.waitIdleWorker()
}

func (p *Pool) waitIdleWorker() *Worker {
	p.lock.Lock()
	p.cond.Wait() // 阻塞等待其他goroutine通知worker已经执行完，并放回池子
	idleWorkers := p.workers
	n := len(idleWorkers) - 1
	if n < 0 { // 说明不存在空闲worker
		panic("waitIdleWorker err")
	}
	w := idleWorkers[n]
	idleWorkers[n] = nil
	p.workers = idleWorkers[:n]
	p.lock.Unlock()
	return w
}

func (p *Pool) Release() {
	p.once.Do(func() {
		p.lock.Lock()
		workers := p.workers
		for i, w := range workers {
			if w == nil {
				continue
			}
			w.task = nil // todo 这里是否要将task执行完再
			w.pool = nil
			workers[i] = nil
		}
		p.workers = nil
		p.lock.Unlock()
		p.release <- sig{}
	})
}

func (p *Pool) IsClosed() bool {
	return len(p.release) > 0
}

func (p *Pool) RunningWorkerCount() int {
	return int(atomic.LoadInt32(&p.running))
}

func (p *Pool) FreeWorkerCount() int {
	return int(p.cap - p.running)
}

// worker中长时间没有任务，需要将worker清理掉，防止一直占用内存
func (p *Pool) clearExpireWorker() {
	ticker := time.NewTicker(p.expire)
	for {
		select {
		case <-ticker.C:
			if p.IsClosed() {
				return
			}
			//循环空闲的workers 如果当前时间和worker的最后运行任务的时间 差值大于expire 进行清理
			p.lock.Lock()
			idleWorkers := p.workers
			for i := len(idleWorkers) - 1; i >= 0; i-- {
				w := idleWorkers[i]
				diffTime := time.Now().Sub(w.lastTime)
				if diffTime < p.expire {
					continue
				}
				w.task = nil
				p.workers[i] = nil
				p.workers = append(p.workers[:i], p.workers[i+1:]...)
				log.Printf("worker%d超时%v，已被清理,running:%d, workers:%v \n", i, diffTime, p.running, p.workers)
			}
			log.Printf("workers:%v", p.workers)
			p.lock.Unlock()
		}
	}
}
