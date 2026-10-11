package persistence

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/bryanwahyu/worker-siem/internal/domain/entities"
	"github.com/bryanwahyu/worker-siem/internal/domain/repositories"
	"github.com/bryanwahyu/worker-siem/internal/domain/valueobjects"
)

// DiskSpoolRepository stores spool items on disk as JSON files.
type DiskSpoolRepository struct {
	Directory string
	mu        sync.Mutex
}

var _ repositories.SpoolRepository = (*DiskSpoolRepository)(nil)

// spoolFile is the on-disk representation of a spool item.
type spoolFile struct {
	Event        spoolEvent `json:"event"`
	AttemptCount int        `json:"attempt_count"`
	NextAttempt  time.Time  `json:"next_attempt"`
	LastError    string     `json:"last_error"`
}

type spoolEvent struct {
	ID         string         `json:"id"`
	Timestamp  time.Time      `json:"timestamp"`
	Message    string         `json:"message"`
	Source     string         `json:"source"`
	Severity   string         `json:"severity"`
	Host       string         `json:"host"`
	IP         string         `json:"ip"`
	StatusCode int            `json:"status_code"`
	Metadata   map[string]any `json:"metadata"`
}

// Enqueue writes/overwrites the spool file for a given event.
func (r *DiskSpoolRepository) Enqueue(ctx context.Context, item entities.SpoolItem) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if err := os.MkdirAll(r.Directory, 0o700); err != nil {
		return err
	}

	file := spoolFile{
		Event: spoolEvent{
			ID:         item.Event.ID,
			Timestamp:  item.Event.Timestamp,
			Message:    item.Event.Message,
			Source:     item.Event.Source,
			Severity:   item.Event.Severity,
			Host:       item.Event.Host,
			IP:         item.Event.IP.String(),
			StatusCode: item.Event.StatusCode.Int(),
			Metadata:   item.Event.Metadata,
		},
		AttemptCount: item.AttemptCount,
		NextAttempt:  item.NextAttempt,
		LastError:    item.LastError,
	}

	data, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return err
	}

	path := r.pathFor(item.Event.ID)
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o600); err != nil {
		return err
	}

	return os.Rename(tmpPath, path)
}

// DequeueBatch returns up to limit ready spool items sorted by NextAttempt ascending.
func (r *DiskSpoolRepository) DequeueBatch(ctx context.Context, limit int) ([]entities.SpoolItem, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	entries := make([]entities.SpoolItem, 0)

	if _, statErr := os.Stat(r.Directory); errors.Is(statErr, os.ErrNotExist) {
		return nil, nil
	} else if statErr != nil {
		return nil, statErr
	}

	err := filepath.WalkDir(r.Directory, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".json" {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		bytes, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var stored spoolFile
		if err := json.Unmarshal(bytes, &stored); err != nil {
			return err
		}
		if stored.NextAttempt.After(time.Now()) {
			return nil
		}
		event, err := restoreEvent(stored.Event)
		if err != nil {
			return err
		}
		entries = append(entries, entities.SpoolItem{
			Event:        event,
			AttemptCount: stored.AttemptCount,
			NextAttempt:  stored.NextAttempt,
			LastError:    stored.LastError,
		})
		return nil
	})
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].NextAttempt.Before(entries[j].NextAttempt)
	})

	if limit > 0 && len(entries) > limit {
		entries = entries[:limit]
	}

	return entries, nil
}

// Delete removes the spool files for the given event IDs.
func (r *DiskSpoolRepository) Delete(ctx context.Context, ids []string) error {
	for _, id := range ids {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if err := os.Remove(r.pathFor(id)); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	return nil
}

// Count returns the number of .json files in the spool directory.
func (r *DiskSpoolRepository) Count(ctx context.Context) (int, error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
	}

	entries, err := os.ReadDir(r.Directory)
	if errors.Is(err, os.ErrNotExist) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	count := 0
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			count++
		}
	}
	return count, nil
}

func (r *DiskSpoolRepository) pathFor(id string) string {
	return filepath.Join(r.Directory, fmt.Sprintf("%s.json", id))
}

func restoreEvent(event spoolEvent) (entities.Event, error) {
	ip, err := valueobjects.NewIP(event.IP)
	if err != nil {
		return entities.Event{}, err
	}
	status, err := valueobjects.NewStatusCode(event.StatusCode)
	if err != nil {
		return entities.Event{}, err
	}
	return entities.NewEvent(event.ID, event.Timestamp, event.Message, event.Source, event.Severity, event.Host, ip, status, event.Metadata), nil
}
