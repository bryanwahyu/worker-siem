package observability

import (
	"fmt"
	"net/http"
	"sync/atomic"

	"github.com/bryanwahyu/worker-siem/internal/application/ports"
)

// MetricsAdapter records counters and exposes them in Prometheus text format.
type MetricsAdapter struct {
	ingested atomic.Int64
	uploaded atomic.Int64
	failed   atomic.Int64
	spool    atomic.Int64
}

func NewMetricsAdapter() *MetricsAdapter {
	return &MetricsAdapter{}
}

func (m *MetricsAdapter) IncEventsIngested() {
	m.ingested.Add(1)
}

func (m *MetricsAdapter) IncEventsUploaded(count int) {
	m.uploaded.Add(int64(count))
}

func (m *MetricsAdapter) IncEventsFailed(count int) {
	m.failed.Add(int64(count))
}

func (m *MetricsAdapter) SetSpoolQueueSize(size int) {
	m.spool.Store(int64(size))
}

// ServeHTTP outputs metrics in the Prometheus exposition format.
func (m *MetricsAdapter) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	fmt.Fprintf(w, "# TYPE mini_siem_worker_events_ingested_total counter\nmini_siem_worker_events_ingested_total %d\n", m.ingested.Load())
	fmt.Fprintf(w, "# TYPE mini_siem_worker_events_uploaded_total counter\nmini_siem_worker_events_uploaded_total %d\n", m.uploaded.Load())
	fmt.Fprintf(w, "# TYPE mini_siem_worker_events_failed_total counter\nmini_siem_worker_events_failed_total %d\n", m.failed.Load())
	fmt.Fprintf(w, "# TYPE mini_siem_worker_spool_queue_size gauge\nmini_siem_worker_spool_queue_size %d\n", m.spool.Load())
}

var _ ports.MetricsRecorder = (*MetricsAdapter)(nil)
