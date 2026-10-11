package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bryanwahyu/worker-siem/internal/application/ports"
	"github.com/bryanwahyu/worker-siem/internal/application/usecase"
	"github.com/bryanwahyu/worker-siem/internal/domain/entities"
	"github.com/bryanwahyu/worker-siem/internal/domain/valueobjects"
)

type stubEventRepo struct {
	fail bool
}

func (s *stubEventRepo) UploadBatch(ctx context.Context, events []entities.Event) error {
	if s.fail {
		return errors.New("upload failed")
	}
	return nil
}

type stubSpoolRepo struct {
	enqueued []entities.SpoolItem
}

func (s *stubSpoolRepo) Enqueue(ctx context.Context, item entities.SpoolItem) error {
	s.enqueued = append(s.enqueued, item)
	return nil
}

func (s *stubSpoolRepo) DequeueBatch(ctx context.Context, limit int) ([]entities.SpoolItem, error) {
	return nil, nil
}

func (s *stubSpoolRepo) Delete(ctx context.Context, ids []string) error {
	return nil
}

func (s *stubSpoolRepo) Count(ctx context.Context) (int, error) {
	return len(s.enqueued), nil
}

type noopMetrics struct{}

func (noopMetrics) IncEventsIngested()          {}
func (noopMetrics) IncEventsUploaded(count int) {}
func (noopMetrics) IncEventsFailed(count int)   {}
func (noopMetrics) SetSpoolQueueSize(size int)  {}

var _ ports.MetricsRecorder = noopMetrics{}

type noopLogger struct{}

func (noopLogger) Info(context.Context, string, map[string]any)         {}
func (noopLogger) Error(context.Context, string, error, map[string]any) {}

func makeEvent(t *testing.T) entities.Event {
	t.Helper()
	ip, err := valueobjects.NewIP("127.0.0.1")
	if err != nil {
		t.Fatalf("ip: %v", err)
	}
	status, err := valueobjects.NewStatusCode(200)
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	return entities.NewEvent("id", time.Now(), "msg", "src", "info", "host", ip, status, nil)
}

func TestUploadEventsSuccess(t *testing.T) {
	spool := &stubSpoolRepo{}
	uc := usecase.NewUploadEventsUsecase(&stubEventRepo{}, spool, noopMetrics{}, noopLogger{})

	res, err := uc.Execute(context.Background(), []entities.Event{makeEvent(t)})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if res.Uploaded != 1 {
		t.Fatalf("expected uploaded=1, got %+v", res)
	}
	if len(spool.enqueued) != 0 {
		t.Fatalf("expected no spool entries")
	}
}

func TestUploadEventsFailureSpools(t *testing.T) {
	spool := &stubSpoolRepo{}
	uc := usecase.NewUploadEventsUsecase(&stubEventRepo{fail: true}, spool, noopMetrics{}, noopLogger{})

	res, err := uc.Execute(context.Background(), []entities.Event{makeEvent(t)})
	if err == nil {
		t.Fatalf("expected error")
	}
	if res.Spooled != 1 {
		t.Fatalf("expected spooled=1, got %+v", res)
	}
	if len(spool.enqueued) != 1 {
		t.Fatalf("expected spool item")
	}
}
