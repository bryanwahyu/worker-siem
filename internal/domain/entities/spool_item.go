package entities

import "time"

// SpoolItem represents an event that has been persisted for retry.
type SpoolItem struct {
	Event        Event
	AttemptCount int
	NextAttempt  time.Time
	LastError    string
}

// IncrementAttempt updates attempt metadata using exponential backoff.
func (s *SpoolItem) IncrementAttempt(backoff time.Duration) {
	s.AttemptCount++
	s.NextAttempt = time.Now().Add(backoff)
}
