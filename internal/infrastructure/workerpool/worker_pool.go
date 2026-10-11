package workerpool

import (
	"context"
	"sync"
	"time"

	"github.com/bryanwahyu/worker-siem/internal/application/usecase"
	"github.com/bryanwahyu/worker-siem/internal/domain/entities"
)

// WorkerPool orchestrates concurrent uploads with batching and spool flushing.
type WorkerPool struct {
	Events        <-chan entities.Event
	UploadUsecase usecase.UploadEventsUsecase
	FlushUsecase  usecase.FlushSpoolUsecase
	BatchSize     int
	BatchInterval time.Duration
	Concurrency   int
	FlushInterval time.Duration
}

// Run starts the worker pool and blocks until the context is cancelled.
func (p WorkerPool) Run(ctx context.Context) error {
	if p.BatchSize <= 0 {
		p.BatchSize = 100
	}
	if p.BatchInterval <= 0 {
		p.BatchInterval = 2 * time.Second
	}
	if p.Concurrency <= 0 {
		p.Concurrency = 4
	}
	if p.FlushInterval <= 0 {
		p.FlushInterval = 30 * time.Second
	}

	jobs := make(chan []entities.Event)
	var wg sync.WaitGroup

	// worker goroutines
	wg.Add(p.Concurrency)
	for i := 0; i < p.Concurrency; i++ {
		go func() {
			defer wg.Done()
			for batch := range jobs {
				_, _ = p.UploadUsecase.Execute(ctx, batch)
			}
		}()
	}

	aggregatorDone := make(chan struct{})
	go func() {
		defer close(aggregatorDone)
		ticker := time.NewTicker(p.BatchInterval)
		defer ticker.Stop()

		batch := make([]entities.Event, 0, p.BatchSize)

		flushBatch := func() {
			if len(batch) == 0 {
				return
			}
			copyBatch := make([]entities.Event, len(batch))
			copy(copyBatch, batch)
			batch = batch[:0]
			select {
			case jobs <- copyBatch:
			case <-ctx.Done():
			}
		}

		for {
			select {
			case <-ctx.Done():
				flushBatch()
				close(jobs)
				return
			case event, ok := <-p.Events:
				if !ok {
					flushBatch()
					close(jobs)
					return
				}
				batch = append(batch, event)
				if len(batch) >= p.BatchSize {
					flushBatch()
				}
			case <-ticker.C:
				flushBatch()
			}
		}
	}()

	flushCtx, cancelFlush := context.WithCancel(ctx)
	var flushWG sync.WaitGroup
	flushWG.Add(1)
	go func() {
		defer flushWG.Done()
		ticker := time.NewTicker(p.FlushInterval)
		defer ticker.Stop()
		for {
			select {
			case <-flushCtx.Done():
				return
			case <-ticker.C:
				_, _ = p.FlushUsecase.Execute(ctx, p.BatchSize)
			}
		}
	}()

	<-aggregatorDone
	wg.Wait()
	cancelFlush()
	flushWG.Wait()

	// Final flush to ensure spool drained.
	_, _ = p.FlushUsecase.Execute(context.Background(), p.BatchSize)

	return nil
}
