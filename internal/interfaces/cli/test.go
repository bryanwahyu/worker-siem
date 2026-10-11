package cli

import (
	"context"
	"time"
)

// Diagnostics runs lightweight checks against configuration and spool.
func Diagnostics(ctx context.Context, configPath string) error {
	runtime, err := loadRuntime(configPath, false)
	if err != nil {
		return err
	}

	diagCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	count, err := runtime.SpoolRepo.Count(diagCtx)
	if err != nil {
		runtime.Logger.Error(ctx, "spool count failed", err, nil)
		return err
	}

	runtime.Logger.Info(ctx, "diagnostics", map[string]any{
		"spool_directory": runtime.Config.Spool.Directory,
		"spool_count":     count,
		"dry_run":         runtime.DryRun,
	})
	return nil
}
