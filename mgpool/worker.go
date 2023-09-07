package mgpool

import (
	"log"
	"time"
)

type Worker struct {
	pool     *Pool
	task     chan func()
	lastTime time.Time // 执行任务的最后时间
}

func (w *Worker) run() {
	w.pool.incRunning()
	go w.running()
}

func (w *Worker) running() {
	defer func() {
		w.pool.PutWorker(w) // 任务运行完成，需要处理的事情
		if err := recover(); err != nil {
			if w.pool.PanicHandler != nil {
				w.pool.PanicHandler()
			} else {
				log.Println(err)
			}
		}
	}()
	for f := range w.task {
		if f == nil {
			return
		}
		f()
		return
	}
}
