package persistence_test

import (
	"context"
	"testing"
	"time"

	"github.com/bryanwahyu/worker-siem/internal/domain/entities"
	"github.com/bryanwahyu/worker-siem/internal/domain/valueobjects"
	"github.com/bryanwahyu/worker-siem/internal/infrastructure/persistence"
)

func TestDiskSpoolRepositoryLifecycle(t *testing.T) {
	dir := t.TempDir()
	repo := &persistence.DiskSpoolRepository{Directory: dir}

	ip, _ := valueobjects.NewIP("127.0.0.1")
	status, _ := valueobjects.NewStatusCode(500)
	event := entities.NewEvent("id-1", time.Now(), "message", "source", "warn", "host", ip, status, map[string]any{"key": "value"})

	item := entities.SpoolItem{Event: event, AttemptCount: 1, NextAttempt: time.Now().Add(-time.Minute)}

	if err := repo.Enqueue(context.Background(), item); err != nil {
		t.Fatalf("enqueue: %v", err)
	}

	count, err := repo.Count(context.Background())
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1, got %d", count)
	}

	items, err := repo.DequeueBatch(context.Background(), 10)
	if err != nil {
		t.Fatalf("dequeue: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Event.ID != event.ID {
		t.Fatalf("unexpected event id: %s", items[0].Event.ID)
	}

	if err := repo.Delete(context.Background(), []string{event.ID}); err != nil {
		t.Fatalf("delete: %v", err)
	}

	count, err = repo.Count(context.Background())
	if err != nil {
		t.Fatalf("count after delete: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0, got %d", count)
	}
}
