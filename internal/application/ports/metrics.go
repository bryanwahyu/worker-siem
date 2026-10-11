package ports

// MetricsRecorder exposes observability counters needed by application use cases.
type MetricsRecorder interface {
	IncEventsIngested()
	IncEventsUploaded(count int)
	IncEventsFailed(count int)
	SetSpoolQueueSize(size int)
}
