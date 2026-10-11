package dto

import "time"

// LogDTO carries data from tailers into the application layer.
type LogDTO struct {
	Timestamp  time.Time
	Message    string
	Source     string
	Severity   string
	Host       string
	IP         string
	StatusCode int
	Metadata   map[string]any
}
