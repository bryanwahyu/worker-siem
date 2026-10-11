package ports

import "context"

// Logger abstracts structured logging used by application use cases.
type Logger interface {
	Info(ctx context.Context, msg string, fields map[string]any)
	Error(ctx context.Context, msg string, err error, fields map[string]any)
}
