package services

import (
	"errors"
	"fmt"
	"github.com/bryanwahyu/worker-siem/internal/domain/entities"
	"strings"
)

// EventValidator enforces business invariants on Events.
type EventValidator struct{}

func NewEventValidator() EventValidator {
	return EventValidator{}
}

// Validate returns an error describing all violations found.
func (EventValidator) Validate(event entities.Event) error {
	var problems []string

	if event.ID == "" {
		problems = append(problems, "id is required")
	}
	if event.Timestamp.IsZero() {
		problems = append(problems, "timestamp is required")
	}
	if strings.TrimSpace(event.Message) == "" {
		problems = append(problems, "message is required")
	}
	if strings.TrimSpace(event.Source) == "" {
		problems = append(problems, "source is required")
	}
	if strings.TrimSpace(event.Severity) == "" {
		problems = append(problems, "severity is required")
	}
	if strings.TrimSpace(event.Host) == "" {
		problems = append(problems, "host identifier is required")
	}
	if event.IP.IsZero() {
		problems = append(problems, "ip must be valid")
	}
	if event.StatusCode.IsZero() {
		problems = append(problems, "status code must be valid")
	}

	if len(problems) == 0 {
		return nil
	}

	return errors.New("event validation failed: " + strings.Join(problems, "; "))
}

// MustValidate panics when the event is invalid; useful in tests.
func (v EventValidator) MustValidate(event entities.Event) {
	if err := v.Validate(event); err != nil {
		panic(fmt.Sprintf("event validation failed: %v", err))
	}
}
