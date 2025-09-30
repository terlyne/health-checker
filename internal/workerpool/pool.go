package workerpool

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

type Job struct {
	URL string
}

type Result struct {
	URL          string
	StatusCode   int
	ResponseTime time.Duration
	Error        error
}

func (r Result) String() string {
	if r.Error != nil {
		return fmt.Sprintf("[ERROR] - [%s] Error Message: %s", r.URL, r.Error.Error())
	}
	return fmt.Sprintf("[SUCCESS] - [%s] Status Code: %d | Response Time: %v", r.URL, r.StatusCode, r.ResponseTime)
}

type Pool struct {
	worker       *Worker
	workersCount int

	jobsCh    chan Job
	resultsCh chan Result

	wg     *sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

func New(workersCount int, timeout time.Duration, resultsCh chan Result) *Pool {
	ctx, cancel := context.WithCancel(context.Background())
	return &Pool{
		worker:       newWorker(timeout),
		workersCount: workersCount,

		jobsCh:    make(chan Job, 100), // max count of services being checked at one time
		resultsCh: resultsCh,

		wg:     new(sync.WaitGroup),
		ctx:    ctx,
		cancel: cancel,
	}
}

func (p *Pool) initWorker(id int) {
	defer func() {
		fmt.Printf("[FINISH] - [Worker %d] finished processing\n", id)
		p.wg.Done()
	}()

	for {
		select {
		case <-p.ctx.Done():
			return
		case job, ok := <-p.jobsCh:
			if !ok {
				return
			}
			p.resultsCh <- p.worker.process(job)
		}
	}

}

func (p *Pool) Init() {
	for i := 0; i < p.workersCount; i++ {
		go p.initWorker(i + 1)
	}
}

func (p *Pool) Push(j Job) {
	select {
	case <-p.ctx.Done():
		return
	case p.jobsCh <- j:
		p.wg.Add(1)
	default:
		log.Printf("Job queue is full, dropping job: %s", j.URL)
	}
}

func (p *Pool) Stop() {
	p.cancel()
	close(p.jobsCh)
	p.wg.Wait()
	close(p.resultsCh)
}
