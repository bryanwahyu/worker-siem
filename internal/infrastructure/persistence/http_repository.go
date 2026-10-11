package persistence

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/bryanwahyu/worker-siem/internal/domain/entities"
	"github.com/bryanwahyu/worker-siem/internal/domain/repositories"
)

// HTTPEventRepository persists events to the Mini SIEM HTTP API.
type HTTPEventRepository struct {
	Client  *http.Client
	BaseURL string
	APIKey  string
	HostID  string
}

var _ repositories.EventRepository = (*HTTPEventRepository)(nil)

// UploadBatch implements the EventRepository interface.
func (r *HTTPEventRepository) UploadBatch(ctx context.Context, events []entities.Event) error {
	if len(events) == 0 {
		return nil
	}

	payloadEvents := make([]map[string]any, 0, len(events))
	for _, event := range events {
		host := event.Host
		if host == "" {
			host = r.HostID
		}
		payloadEvents = append(payloadEvents, map[string]any{
			"id":          event.ID,
			"timestamp":   event.Timestamp.Format(time.RFC3339Nano),
			"message":     event.Message,
			"source":      event.Source,
			"severity":    event.Severity,
			"host":        host,
			"ip":          event.IP.String(),
			"status_code": event.StatusCode.Int(),
			"metadata":    event.Metadata,
		})
	}

	bodyBytes, err := json.Marshal(map[string]any{"events": payloadEvents})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/events", r.BaseURL), bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", r.APIKey)

	resp, err := r.httpClient().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	return fmt.Errorf("api responded with status %d", resp.StatusCode)
}

func (r *HTTPEventRepository) httpClient() *http.Client {
	if r.Client != nil {
		return r.Client
	}
	return http.DefaultClient
}
