package usecase

import (
	"context"
	"time"

	"github.com/bryanwahyu/worker-siem/internal/application/dto"
	"github.com/bryanwahyu/worker-siem/internal/application/ports"
	"github.com/bryanwahyu/worker-siem/internal/domain/entities"
	"github.com/bryanwahyu/worker-siem/internal/domain/repositories"
)

// FlushSpoolUsecase replays spooled events back to the Mini SIEM API.
type FlushSpoolUsecase struct {
	repo    repositories.EventRepository
	spool   repositories.SpoolRepository
	metrics ports.MetricsRecorder
	logger  ports.Logger
}

func NewFlushSpoolUsecase(repo repositories.EventRepository, spool repositories.SpoolRepository, metrics ports.MetricsRecorder, logger ports.Logger) FlushSpoolUsecase {
	return FlushSpoolUsecase{repo: repo, spool: spool, metrics: metrics, logger: logger}
}

func (u FlushSpoolUsecase) Execute(ctx context.Context, limit int) (dto.UploadResult, error) {
	if limit <= 0 {
		limit = 100
	}

	items, err := u.spool.DequeueBatch(ctx, limit)
	if err != nil {
		u.logger.Error(ctx, "failed to dequeue spool items", err, nil)
		return dto.UploadResult{}, err
	}

	if len(items) == 0 {
		return dto.UploadResult{}, nil
	}

	events := make([]entities.Event, 0, len(items))
	ids := make([]string, 0, len(items))
	for _, item := range items {
		events = append(events, item.Event)
		ids = append(ids, item.Event.ID)
	}

	err = u.repo.UploadBatch(ctx, events)
	if err == nil {
		if delErr := u.spool.Delete(ctx, ids); delErr != nil {
			u.logger.Error(ctx, "failed to delete flushed items", delErr, map[string]any{"count": len(ids)})
		}
		u.metrics.IncEventsUploaded(len(events))
		u.logger.Info(ctx, "flushed spool batch", map[string]any{"count": len(events)})
		count, countErr := u.spool.Count(ctx)
		if countErr == nil {
			u.metrics.SetSpoolQueueSize(count)
		}
		return dto.UploadResult{Uploaded: len(events)}, nil
	}

	u.logger.Error(ctx, "failed to flush spool batch", err, map[string]any{"count": len(events)})
	u.metrics.IncEventsFailed(len(events))

	for _, item := range items {
		item.AttemptCount++
		item.LastError = err.Error()
		backoff := time.Duration(item.AttemptCount*item.AttemptCount) * time.Second
		item.NextAttempt = time.Now().Add(backoff)
		if enqueueErr := u.spool.Enqueue(ctx, item); enqueueErr != nil {
			u.logger.Error(ctx, "failed to re-enqueue spool item", enqueueErr, map[string]any{"event_id": item.Event.ID})
		}
	}

	count, countErr := u.spool.Count(ctx)
	if countErr == nil {
		u.metrics.SetSpoolQueueSize(count)
	}

	return dto.UploadResult{Failed: len(events)}, err
}
