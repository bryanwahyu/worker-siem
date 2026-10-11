package cli

import (
	"net/http"

	"github.com/bryanwahyu/worker-siem/internal/application/usecase"
	entities "github.com/bryanwahyu/worker-siem/internal/domain/entities"
	"github.com/bryanwahyu/worker-siem/internal/domain/services"
	"github.com/bryanwahyu/worker-siem/internal/infrastructure/config"
	"github.com/bryanwahyu/worker-siem/internal/infrastructure/observability"
	"github.com/bryanwahyu/worker-siem/internal/infrastructure/persistence"
	"github.com/bryanwahyu/worker-siem/internal/infrastructure/tailer"
	"github.com/bryanwahyu/worker-siem/internal/infrastructure/workerpool"
)

type runtimeResources struct {
	Config        config.Config
	Logger        observability.LoggerAdapter
	Metrics       *observability.MetricsAdapter
	SpoolRepo     *persistence.DiskSpoolRepository
	EventRepo     *persistence.HTTPEventRepository
	UploadUsecase usecase.UploadEventsUsecase
	FlushUsecase  usecase.FlushSpoolUsecase
	IngestUsecase usecase.IngestLogsUsecase
	WorkerPool    workerpool.WorkerPool
	EventsChan    chan entities.Event
	Tailers       []tailer.Tailer
	DryRun        bool
}

func loadRuntime(configPath string, dryRunFlag bool) (runtimeResources, error) {
	cfg, err := config.Load(configPath)
	if err != nil {
		return runtimeResources{}, err
	}

	dryRun := cfg.DryRun || dryRunFlag

	logger := observability.NewLoggerAdapter()
	metrics := observability.NewMetricsAdapter()

	httpClient := &http.Client{Timeout: cfg.API.Timeout}
	eventRepo := &persistence.HTTPEventRepository{Client: httpClient, BaseURL: cfg.API.BaseURL, APIKey: cfg.API.APIKey, HostID: cfg.HostID}
	spoolRepo := &persistence.DiskSpoolRepository{Directory: cfg.Spool.Directory}

	eventsChan := make(chan entities.Event, cfg.Worker.BatchSize*2)

	uploadUC := usecase.NewUploadEventsUsecase(eventRepo, spoolRepo, metrics, logger)
	flushUC := usecase.NewFlushSpoolUsecase(eventRepo, spoolRepo, metrics, logger)
	ingestUC := usecase.NewIngestLogsUsecase(
		services.NewEventTransformer(),
		services.NewEventValidator(),
		eventsChan,
		spoolRepo,
		metrics,
		logger,
		dryRun,
	)

	wp := workerpool.WorkerPool{
		Events:        eventsChan,
		UploadUsecase: uploadUC,
		FlushUsecase:  flushUC,
		BatchSize:     cfg.Worker.BatchSize,
		BatchInterval: cfg.Worker.BatchInterval,
		Concurrency:   cfg.Worker.Concurrency,
		FlushInterval: cfg.Spool.FlushInterval,
	}

	tailers := make([]tailer.Tailer, 0, len(cfg.Tailers.Files)+1)
	parser := tailer.NewDefaultParser(cfg.HostID)
	for _, file := range cfg.Tailers.Files {
		tailers = append(tailers, tailer.FileTailer{Path: file, Parser: parser, Logger: logger})
	}
	if cfg.Tailers.JournaldUnit != "" {
		tailers = append(tailers, tailer.JournaldTailer{Unit: cfg.Tailers.JournaldUnit, Parser: parser, Logger: logger})
	}

	return runtimeResources{
		Config:        cfg,
		Logger:        logger,
		Metrics:       metrics,
		SpoolRepo:     spoolRepo,
		EventRepo:     eventRepo,
		UploadUsecase: uploadUC,
		FlushUsecase:  flushUC,
		IngestUsecase: ingestUC,
		WorkerPool:    wp,
		EventsChan:    eventsChan,
		Tailers:       tailers,
		DryRun:        dryRun,
	}, nil
}
