package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/terlyne/health-checker/internal/config"
	"github.com/terlyne/health-checker/internal/workerpool"
)

const (
	WORKERS_COUNT   = 3
	REQUEST_TIMEOUT = 3 * time.Second
	INTERVAL        = 20 * time.Second
)

func main() {
	// parse configPath from flag
	var configPath string
	flag.StringVar(&configPath, "config-path", "", "path to config file")
	flag.Parse()

	if configPath == "" {
		log.Fatal(`"-config-path" flag can not be empty`)
	}

	// load config
	cfg := config.MustLoad(configPath)

	stopCh := make(chan struct{})

	// init pool
	resultsCh := make(chan workerpool.Result, 100)
	pool := workerpool.New(WORKERS_COUNT, REQUEST_TIMEOUT, resultsCh)

	pool.Init()

	// graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// generating jobs
	go generateJobs(pool, cfg.Services, stopCh)

	// print results
	go processResults(resultsCh, stopCh)

	<-sigChan
	close(stopCh)
	pool.Stop()
}

func generateJobs(pool *workerpool.Pool, urls []string, stopCh chan struct{}) {
	ticker := time.NewTicker(INTERVAL)
	defer ticker.Stop()

	// for first time, launch jobs immediately
	for _, url := range urls {
		pool.Push(workerpool.Job{URL: url})
	}

	for {
		select {
		case <-stopCh:
			return
		case <-ticker.C:
			for _, url := range urls {
				pool.Push(workerpool.Job{URL: url})
			}
		}

	}

}

func processResults(resultCh chan workerpool.Result, stopCh chan struct{}) {
	for {
		select {
		case <-stopCh:
			return
		case result := <-resultCh:
			fmt.Println(result)
		}
	}
}
