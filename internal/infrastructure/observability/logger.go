package observability

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/bryanwahyu/worker-siem/internal/application/ports"
)

// LoggerAdapter implements the application Logger port using the standard library.
type LoggerAdapter struct {
	mu  *sync.Mutex
	out io.Writer
}

func NewLoggerAdapter() LoggerAdapter {
	return LoggerAdapter{mu: &sync.Mutex{}, out: os.Stdout}
}

func (l LoggerAdapter) Info(_ context.Context, msg string, fields map[string]any) {
	l.write("info", msg, nil, fields)
}

func (l LoggerAdapter) Error(_ context.Context, msg string, err error, fields map[string]any) {
	l.write("error", msg, err, fields)
}

func (l LoggerAdapter) write(level, msg string, err error, fields map[string]any) {
	entry := map[string]any{
		"ts":    time.Now().UTC().Format(time.RFC3339Nano),
		"level": level,
		"msg":   msg,
	}
	if err != nil {
		entry["error"] = err.Error()
	}
	for k, v := range fields {
		entry[k] = v
	}

	data, marshalErr := json.Marshal(entry)
	l.mu.Lock()
	defer l.mu.Unlock()
	if marshalErr != nil {
		fmt.Fprintf(l.out, "{\"level\":\"error\",\"msg\":\"logger marshal error\",\"err\":%q}\n", marshalErr)
		return
	}
	fmt.Fprintf(l.out, "%s\n", data)
}

var _ ports.Logger = LoggerAdapter{}
