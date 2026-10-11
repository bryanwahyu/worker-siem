package repositories

import (
	"context"

	"github.com/bryanwahyu/worker-siem/internal/domain/entities"
)

// EventRepository abstracts persistence of events to Mini SIEM API.
type EventRepository interface {
	UploadBatch(ctx context.Context, events []entities.Event) error
}

// SpoolRepository represents durable storage for failed uploads.
type SpoolRepository interface {
	Enqueue(ctx context.Context, item entities.SpoolItem) error
	DequeueBatch(ctx context.Context, limit int) ([]entities.SpoolItem, error)
	Delete(ctx context.Context, ids []string) error
	Count(ctx context.Context) (int, error)
}
