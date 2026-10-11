package tailer

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/bryanwahyu/worker-siem/internal/application/dto"
)

// JSONLine represents an optional schema for structured logs.
type JSONLine struct {
	Timestamp  string         `json:"timestamp"`
	Message    string         `json:"message"`
	Source     string         `json:"source"`
	Severity   string         `json:"severity"`
	Host       string         `json:"host"`
	IP         string         `json:"ip"`
	StatusCode *int           `json:"status_code"`
	Metadata   map[string]any `json:"metadata"`
}

// NewDefaultParser returns a parser that handles JSON lines and plain text fallbacks.
func NewDefaultParser(defaultHost string) Parser {
	return func(line string) (dto.LogDTO, error) {
		line = strings.TrimSpace(line)
		if line == "" {
			return dto.LogDTO{}, nil
		}

		var jsonLine JSONLine
		if json.Unmarshal([]byte(line), &jsonLine) == nil && jsonLine.Message != "" {
			ts := time.Now().UTC()
			if jsonLine.Timestamp != "" {
				if parsed, err := time.Parse(time.RFC3339Nano, jsonLine.Timestamp); err == nil {
					ts = parsed
				}
			}
			status := 200
			if jsonLine.StatusCode != nil {
				status = *jsonLine.StatusCode
			}
			host := jsonLine.Host
			if host == "" {
				host = defaultHost
			}
			return dto.LogDTO{
				Timestamp:  ts,
				Message:    jsonLine.Message,
				Source:     fallback(jsonLine.Source, "file"),
				Severity:   fallback(jsonLine.Severity, "info"),
				Host:       host,
				IP:         fallback(jsonLine.IP, "127.0.0.1"),
				StatusCode: status,
				Metadata:   jsonLine.Metadata,
			}, nil
		}

		return dto.LogDTO{
			Timestamp:  time.Now().UTC(),
			Message:    line,
			Source:     "file",
			Severity:   "info",
			Host:       defaultHost,
			IP:         "127.0.0.1",
			StatusCode: 200,
			Metadata:   map[string]any{"raw": line},
		}, nil
	}
}

func fallback(value, def string) string {
	if value == "" {
		return def
	}
	return value
}
