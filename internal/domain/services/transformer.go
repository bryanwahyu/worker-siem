package services

import (
	"time"

	"github.com/bryanwahyu/worker-siem/internal/domain/entities"
	"github.com/bryanwahyu/worker-siem/internal/domain/valueobjects"
	id "github.com/bryanwahyu/worker-siem/internal/pkg/id"
)

// LogRecord represents a parsed log entry from a tailer before domain validation.
type LogRecord struct {
	Timestamp  time.Time
	Message    string
	Source     string
	Severity   string
	Host       string
	IP         string
	StatusCode int
	Metadata   map[string]any
}

// EventTransformer builds domain Events from incoming records.
type EventTransformer struct{}

func NewEventTransformer() EventTransformer {
	return EventTransformer{}
}

// Transform converts a log record into a domain Event, generating defaults as needed.
func (EventTransformer) Transform(record LogRecord) (entities.Event, error) {
	if record.Metadata == nil {
		record.Metadata = map[string]any{}
	}

	ip, err := valueobjects.NewIP(record.IP)
	if err != nil {
		return entities.Event{}, err
	}

	status, err := valueobjects.NewStatusCode(record.StatusCode)
	if err != nil {
		return entities.Event{}, err
	}

	timestamp := record.Timestamp
	if timestamp.IsZero() {
		timestamp = time.Now().UTC()
	}

	event := entities.NewEvent(
		id.NewID(),
		timestamp,
		record.Message,
		record.Source,
		record.Severity,
		record.Host,
		ip,
		status,
		record.Metadata,
	)

	return event, nil
}
