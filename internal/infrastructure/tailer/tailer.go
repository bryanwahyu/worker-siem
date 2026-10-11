package tailer

import (
	"context"

	"github.com/bryanwahyu/worker-siem/internal/application/dto"
)

// Parser converts a raw line into a DTO for ingestion.
type Parser func(line string) (dto.LogDTO, error)

// Tailer streams log entries into the ingestion channel.
type Tailer interface {
	Start(ctx context.Context, out chan<- dto.LogDTO) error
}
