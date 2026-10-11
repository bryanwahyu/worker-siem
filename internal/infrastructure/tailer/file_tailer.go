package tailer

import (
	"bufio"
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"time"

	"github.com/bryanwahyu/worker-siem/internal/application/dto"
	"github.com/bryanwahyu/worker-siem/internal/application/ports"
)

// FileTailer streams a single file using a polling approach.
type FileTailer struct {
	Path   string
	Parser Parser
	Logger ports.Logger
}

func (t FileTailer) Start(ctx context.Context, out chan<- dto.LogDTO) error {
	if t.Parser == nil {
		return errors.New("parser is required")
	}

	var (
		file   *os.File
		reader *bufio.Reader
		offset int64
	)

	openFile := func() error {
		if file != nil {
			file.Close()
		}
		f, err := os.Open(t.Path)
		if err != nil {
			return err
		}
		if _, err := f.Seek(0, io.SeekEnd); err != nil {
			f.Close()
			return err
		}
		file = f
		reader = bufio.NewReader(f)
		offset = 0
		return nil
	}

	if err := openFile(); err != nil {
		return err
	}
	defer func() {
		if file != nil {
			file.Close()
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				time.Sleep(200 * time.Millisecond)
				stat, statErr := file.Stat()
				if statErr == nil {
					if stat.Size() < offset {
						if err := openFile(); err != nil {
							t.Logger.Error(ctx, "reopen failed", err, map[string]any{"path": t.Path})
							time.Sleep(time.Second)
						}
					}
				}
				continue
			}
			t.Logger.Error(ctx, "tail read error", err, map[string]any{"path": t.Path})
			reader.Reset(file)
			continue
		}

		offset += int64(len(line))
		line = strings.TrimRight(line, "\r\n")
		dtoValue, err := t.Parser(line)
		if err != nil {
			t.Logger.Error(ctx, "failed to parse line", err, map[string]any{"path": t.Path})
			continue
		}
		if dtoValue.Message == "" {
			continue
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case out <- dtoValue:
		}
	}
}
