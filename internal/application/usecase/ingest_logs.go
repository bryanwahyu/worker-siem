package usecase

import (
	"context"
	"time"

	"github.com/bryanwahyu/worker-siem/internal/application/dto"
	"github.com/bryanwahyu/worker-siem/internal/application/ports"
	"github.com/bryanwahyu/worker-siem/internal/domain/entities"
	"github.com/bryanwahyu/worker-siem/internal/domain/repositories"
	"github.com/bryanwahyu/worker-siem/internal/domain/services"
)

// IngestLogsUsecase transforms inbound logs into domain events and dispatches them for upload.
type IngestLogsUsecase struct {
	transformer services.EventTransformer
	validator   services.EventValidator
	eventsChan  chan<- entities.Event
	spool       repositories.SpoolRepository
	metrics     ports.MetricsRecorder
	logger      ports.Logger
	dryRun      bool
}

func NewIngestLogsUsecase(
	transformer services.EventTransformer,
	validator services.EventValidator,
	eventsChan chan<- entities.Event,
	spool repositories.SpoolRepository,
	metrics ports.MetricsRecorder,
	logger ports.Logger,
	dryRun bool,
) IngestLogsUsecase {
	return IngestLogsUsecase{
		transformer: transformer,
		validator:   validator,
		eventsChan:  eventsChan,
		spool:       spool,
		metrics:     metrics,
		logger:      logger,
		dryRun:      dryRun,
	}
}

func (u IngestLogsUsecase) Execute(ctx context.Context, input dto.LogDTO) error {
	record := services.LogRecord{
		Timestamp:  input.Timestamp,
		Message:    input.Message,
		Source:     input.Source,
		Severity:   input.Severity,
		Host:       input.Host,
		IP:         input.IP,
		StatusCode: input.StatusCode,
		Metadata:   input.Metadata,
	}

	event, err := u.transformer.Transform(record)
	if err != nil {
		u.logger.Error(ctx, "failed to transform log", err, map[string]any{"source": input.Source})
		return err
	}

	if err := u.validator.Validate(event); err != nil {
		u.logger.Error(ctx, "invalid event", err, map[string]any{"source": input.Source})
		return err
	}

	if u.dryRun {
		spoolItem := entities.SpoolItem{Event: event.Clone(), NextAttempt: time.Now(), AttemptCount: 0}
		if err := u.spool.Enqueue(ctx, spoolItem); err != nil {
			u.logger.Error(ctx, "failed to spool event in dry-run", err, map[string]any{"event_id": event.ID})
			return err
		}
		u.metrics.IncEventsIngested()
		return nil
	}

	select {
	case <-ctx.Done():
		spoolItem := entities.SpoolItem{Event: event.Clone(), NextAttempt: time.Now()}
		_ = u.spool.Enqueue(context.Background(), spoolItem)
		return ctx.Err()
	case u.eventsChan <- event:
		u.metrics.IncEventsIngested()
		u.logger.Info(ctx, "event enqueued", map[string]any{"event_id": event.ID})
		return nil
	}
}
