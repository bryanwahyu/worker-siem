package cli

import (
	"context"
	"os/signal"
	"sync"
	"syscall"

	"github.com/bryanwahyu/worker-siem/internal/application/dto"
	"github.com/bryanwahyu/worker-siem/internal/infrastructure/observability"
)

// Run executes the main worker lifecycle.
func Run(ctx context.Context, configPath string, dryRun bool) error {
	runtime, err := loadRuntime(configPath, dryRun)
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logChan := make(chan dto.LogDTO, runtime.Config.Worker.BatchSize*2)

	var tailerWG sync.WaitGroup
	for _, t := range runtime.Tailers {
		tailerWG.Add(1)
		tailer := t
		go func() {
			defer tailerWG.Done()
			if err := tailer.Start(ctx, logChan); err != nil {
				runtime.Logger.Error(ctx, "tailer exited", err, nil)
			}
		}()
	}

	go func() {
		tailerWG.Wait()
		close(logChan)
	}()

	var ingestWG sync.WaitGroup
	ingestWG.Add(1)
	go func() {
		defer ingestWG.Done()
		defer close(runtime.EventsChan)
		for logItem := range logChan {
			if logItem.Message == "" {
				continue
			}
			if err := runtime.IngestUsecase.Execute(ctx, logItem); err != nil {
				runtime.Logger.Error(ctx, "ingest failed", err, map[string]any{"source": logItem.Source})
			}
		}
	}()

	go func() {
		if err := observability.StartMetricsServer(ctx, runtime.Config.Metrics.ListenAddr, runtime.Metrics, runtime.Logger); err != nil {
			runtime.Logger.Error(ctx, "metrics server failure", err, nil)
		}
	}()

	if err := runtime.WorkerPool.Run(ctx); err != nil {
		runtime.Logger.Error(ctx, "worker pool exited", err, nil)
	}

	stop()
	tailerWG.Wait()
	ingestWG.Wait()

	return nil
}
