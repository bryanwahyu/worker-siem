package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/bryanwahyu/worker-siem/internal/application/dto"
	"github.com/bryanwahyu/worker-siem/internal/application/usecase"
	"github.com/bryanwahyu/worker-siem/internal/domain/entities"
	"github.com/bryanwahyu/worker-siem/internal/domain/services"
)

type captureSpool struct {
	items []entities.SpoolItem
}

func (c *captureSpool) Enqueue(ctx context.Context, item entities.SpoolItem) error {
	c.items = append(c.items, item)
	return nil
}

func (c *captureSpool) DequeueBatch(ctx context.Context, limit int) ([]entities.SpoolItem, error) {
	return nil, nil
}
func (c *captureSpool) Delete(ctx context.Context, ids []string) error { return nil }
func (c *captureSpool) Count(ctx context.Context) (int, error)         { return len(c.items), nil }

func TestIngestDryRunSpools(t *testing.T) {
	eventsChan := make(chan entities.Event, 1)
	spool := &captureSpool{}
	uc := usecase.NewIngestLogsUsecase(
		services.NewEventTransformer(),
		services.NewEventValidator(),
		eventsChan,
		spool,
		noopMetrics{},
		noopLogger{},
		true,
	)

	err := uc.Execute(context.Background(), dto.LogDTO{
		Timestamp:  time.Now(),
		Message:    "msg",
		Source:     "src",
		Severity:   "info",
		Host:       "host",
		IP:         "127.0.0.1",
		StatusCode: 200,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(spool.items) != 1 {
		t.Fatalf("expected item in spool")
	}
	if len(eventsChan) != 0 {
		t.Fatalf("expected channel empty in dry-run")
	}
}

func TestIngestNormalEnqueuesEvent(t *testing.T) {
	eventsChan := make(chan entities.Event, 1)
	spool := &captureSpool{}
	uc := usecase.NewIngestLogsUsecase(
		services.NewEventTransformer(),
		services.NewEventValidator(),
		eventsChan,
		spool,
		noopMetrics{},
		noopLogger{},
		false,
	)

	err := uc.Execute(context.Background(), dto.LogDTO{
		Timestamp:  time.Now(),
		Message:    "msg",
		Source:     "src",
		Severity:   "info",
		Host:       "host",
		IP:         "127.0.0.1",
		StatusCode: 200,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(eventsChan) != 1 {
		t.Fatalf("expected event enqueued")
	}
	if len(spool.items) != 0 {
		t.Fatalf("unexpected spool items")
	}
}
