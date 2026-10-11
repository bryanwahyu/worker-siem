package usecase

import (
	"context"
	"time"

	"github.com/bryanwahyu/worker-siem/internal/application/dto"
	"github.com/bryanwahyu/worker-siem/internal/application/ports"
	"github.com/bryanwahyu/worker-siem/internal/domain/entities"
	"github.com/bryanwahyu/worker-siem/internal/domain/repositories"
)

// UploadEventsUsecase attempts to upload a batch of events and spools failures.
type UploadEventsUsecase struct {
	repo    repositories.EventRepository
	spool   repositories.SpoolRepository
	metrics ports.MetricsRecorder
	logger  ports.Logger
}

func NewUploadEventsUsecase(repo repositories.EventRepository, spool repositories.SpoolRepository, metrics ports.MetricsRecorder, logger ports.Logger) UploadEventsUsecase {
	return UploadEventsUsecase{repo: repo, spool: spool, metrics: metrics, logger: logger}
}

func (u UploadEventsUsecase) Execute(ctx context.Context, events []entities.Event) (dto.UploadResult, error) {
	if len(events) == 0 {
		return dto.UploadResult{}, nil
	}

	err := u.repo.UploadBatch(ctx, events)
	if err == nil {
		u.metrics.IncEventsUploaded(len(events))
		u.logger.Info(ctx, "uploaded batch", map[string]any{"count": len(events)})
		return dto.UploadResult{Uploaded: len(events)}, nil
	}

	u.logger.Error(ctx, "upload batch failed", err, map[string]any{"count": len(events)})

	var spooled int
	for _, event := range events {
		item := entities.SpoolItem{
			Event:        event.Clone(),
			AttemptCount: 1,
			NextAttempt:  time.Now().Add(2 * time.Second),
			LastError:    err.Error(),
		}
		if spoolErr := u.spool.Enqueue(ctx, item); spoolErr != nil {
			u.logger.Error(ctx, "failed to spool event", spoolErr, map[string]any{"event_id": event.ID})
			continue
		}
		spooled++
	}

	u.metrics.IncEventsFailed(len(events))

	count, countErr := u.spool.Count(ctx)
	if countErr == nil {
		u.metrics.SetSpoolQueueSize(count)
	}

	return dto.UploadResult{Failed: len(events), Spooled: spooled}, err
}
