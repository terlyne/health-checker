package workerpool

import (
	"net/http"
	"time"
)

type Worker struct {
	client *http.Client
}

func newWorker(timeout time.Duration) *Worker {
	return &Worker{
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (w *Worker) process(j Job) Result {
	result := Result{URL: j.URL}

	start := time.Now()

	resp, err := w.client.Get(j.URL)
	result.ResponseTime = time.Since(start)
	if err != nil {
		result.Error = err
		return result
	}

	result.StatusCode = resp.StatusCode
	return result
}
