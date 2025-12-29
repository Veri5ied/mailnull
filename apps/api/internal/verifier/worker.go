package verifier

import (
	"sync"
)

type Job struct {
	Email string

	ResultChan chan Result
}

type WorkerPool struct {
	verifier *Verifier
	jobs     chan Job
	wg       sync.WaitGroup
	workers  int
}

func NewWorkerPool(v *Verifier, workers int, bufferSize int) *WorkerPool {
	return &WorkerPool{
		verifier: v,
		jobs:     make(chan Job, bufferSize),
		workers:  workers,
	}
}

func (wp *WorkerPool) Start() {
	for i := 0; i < wp.workers; i++ {
		wp.wg.Add(1)
		go wp.worker()
	}
}

func (wp *WorkerPool) worker() {
	defer wp.wg.Done()
	for job := range wp.jobs {
		res := wp.verifier.Verify(job.Email)
		job.ResultChan <- res
	}
}

func (wp *WorkerPool) Submit(email string, resultChan chan Result) {
	wp.jobs <- Job{Email: email, ResultChan: resultChan}
}

func (wp *WorkerPool) Wait() {
	close(wp.jobs)
	wp.wg.Wait()
}
