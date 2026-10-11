package entities

import (
	"time"

	"github.com/bryanwahyu/worker-siem/internal/domain/valueobjects"
)

// Event captures a normalized log record that will be delivered to the Mini SIEM API.
type Event struct {
	ID         string
	Timestamp  time.Time
	Message    string
	Source     string
	Severity   string
	Host       string
	IP         valueobjects.IP
	StatusCode valueobjects.StatusCode
	Metadata   map[string]any
}

// NewEvent constructs an Event with default metadata map when nil.
func NewEvent(id string, ts time.Time, message, source, severity, host string, ip valueobjects.IP, status valueobjects.StatusCode, metadata map[string]any) Event {
	if metadata == nil {
		metadata = make(map[string]any)
	}
	return Event{
		ID:         id,
		Timestamp:  ts,
		Message:    message,
		Source:     source,
		Severity:   severity,
		Host:       host,
		IP:         ip,
		StatusCode: status,
		Metadata:   metadata,
	}
}

// Clone returns a deep-ish copy of the event to avoid shared map mutations.
func (e Event) Clone() Event {
	clone := e
	if e.Metadata != nil {
		clone.Metadata = make(map[string]any, len(e.Metadata))
		for k, v := range e.Metadata {
			clone.Metadata[k] = v
		}
	}
	return clone
}
