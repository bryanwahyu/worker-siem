package tailer

import (
	"bufio"
	"context"
	"errors"
	"io"
	"os/exec"
	"runtime"
	"strings"

	"github.com/bryanwahyu/worker-siem/internal/application/dto"
	"github.com/bryanwahyu/worker-siem/internal/application/ports"
)

// JournaldTailer streams systemd journal entries using journalctl.
type JournaldTailer struct {
	Unit   string
	Parser Parser
	Logger ports.Logger
}

func (t JournaldTailer) Start(ctx context.Context, out chan<- dto.LogDTO) error {
	if t.Parser == nil {
		return errors.New("parser is required")
	}

	if runtime.GOOS != "linux" {
		t.Logger.Info(ctx, "journald tailer not supported on this platform", map[string]any{"unit": t.Unit})
		<-ctx.Done()
		return nil
	}

	args := []string{"-f", "-n", "0", "-o", "cat"}
	if t.Unit != "" {
		args = append([]string{"-u", t.Unit}, args...)
	}

	cmd := exec.CommandContext(ctx, "journalctl", args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	errCh := make(chan error, 1)

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			dtoValue, err := t.Parser(line)
			if err != nil {
				t.Logger.Error(ctx, "failed to parse journald line", err, map[string]any{"unit": t.Unit})
				continue
			}
			if dtoValue.Message == "" {
				continue
			}
			select {
			case <-ctx.Done():
				errCh <- ctx.Err()
				return
			case out <- dtoValue:
			}
		}
		if scanErr := scanner.Err(); scanErr != nil && !errors.Is(scanErr, io.EOF) {
			errCh <- scanErr
			return
		}
		errCh <- nil
	}()

	go func() {
		bytes, _ := io.ReadAll(stderr)
		if len(bytes) > 0 {
			t.Logger.Error(ctx, "journald stderr", errors.New(strings.TrimSpace(string(bytes))), map[string]any{"unit": t.Unit})
		}
	}()

	select {
	case <-ctx.Done():
		_ = cmd.Process.Kill()
		return ctx.Err()
	case err := <-errCh:
		if err != nil {
			return err
		}
	}

	return cmd.Wait()
}
