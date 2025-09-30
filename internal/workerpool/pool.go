package workerpool

import (
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

func (r Result) Info() string {
	if r.Error != nil {
		return fmt.Sprintf("[ERROR] %s | error message: %s", r.URL, r.Error.Error())
	}
	return fmt.Sprintf("[OK] %s | status code: %d | response time: %v", r.URL, r.StatusCode, r.ResponseTime)
}

type Pool struct {
	worker       *Worker
	workersCount int

	jobsCh    chan Job
	resultsCh chan Result

	wg        *sync.WaitGroup
	stoppedCh chan struct{}
}

func New(workersCount int, timeout time.Duration, resultsCh chan Result) *Pool {
	return &Pool{
		worker:       newWorker(timeout),
		workersCount: workersCount,

		jobsCh:    make(chan Job, 100), // max count of services being checked at one time
		resultsCh: resultsCh,

		wg:        new(sync.WaitGroup),
		stoppedCh: make(chan struct{}),
	}
}

func (p *Pool) initWorker(id int) {
	defer func() {
		log.Printf("[FINISH] worker %d | finished processing", id)
		p.wg.Done()
	}()

	for {
		select {
		case <-p.stoppedCh:
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
	p.wg.Add(p.workersCount)
	for i := 0; i < p.workersCount; i++ {
		go p.initWorker(i + 1)
	}
}

func (p *Pool) Push(j Job) {
	select {
	case <-p.stoppedCh:
		return
	case p.jobsCh <- j:

	default:
		log.Printf("Job queue is full, dropping job: %s", j.URL)
	}
}

func (p *Pool) Stop() {
	close(p.stoppedCh)
	close(p.jobsCh)

	p.wg.Wait()

	close(p.resultsCh)
}
