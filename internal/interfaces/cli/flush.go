package cli

import "context"

// Flush triggers a one-off spool flush.
func Flush(ctx context.Context, configPath string) error {
	runtime, err := loadRuntime(configPath, false)
	if err != nil {
		return err
	}

	result, err := runtime.FlushUsecase.Execute(ctx, runtime.Config.Worker.BatchSize)
	if err != nil {
		runtime.Logger.Error(ctx, "flush failed", err, map[string]any{"failed": result.Failed})
		return err
	}
	runtime.Logger.Info(ctx, "flush completed", map[string]any{"uploaded": result.Uploaded})
	return nil
}
