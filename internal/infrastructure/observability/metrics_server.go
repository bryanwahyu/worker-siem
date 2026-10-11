package observability

import (
	"context"
	"net/http"
	"time"

	"github.com/bryanwahyu/worker-siem/internal/application/ports"
)

// StartMetricsServer launches a metrics HTTP server until context cancellation.
func StartMetricsServer(ctx context.Context, addr string, metrics *MetricsAdapter, logger ports.Logger) error {
	mux := http.NewServeMux()
	mux.Handle("/metrics", metrics)

	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	logger.Info(ctx, "starting metrics server", map[string]any{"addr": addr})
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error(ctx, "metrics server error", err, nil)
		return err
	}
	return nil
}
