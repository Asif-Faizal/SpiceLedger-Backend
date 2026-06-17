package platform

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Asif-Faizal/SpiceLedger-Backend/util"
)

// HTTPConfig holds settings for a production HTTP server.
type HTTPConfig struct {
	Name         string
	Addr         string
	Handler      http.Handler
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	ShutdownWait time.Duration
}

// RunHTTP starts an HTTP server and blocks until SIGINT/SIGTERM, then shuts down gracefully.
func RunHTTP(cfg HTTPConfig, logger util.Logger) error {
	if cfg.Name == "" {
		cfg.Name = "http"
	}
	if cfg.ReadTimeout == 0 {
		cfg.ReadTimeout = 15 * time.Second
	}
	if cfg.WriteTimeout == 0 {
		cfg.WriteTimeout = 15 * time.Second
	}
	if cfg.IdleTimeout == 0 {
		cfg.IdleTimeout = 60 * time.Second
	}
	if cfg.ShutdownWait == 0 {
		cfg.ShutdownWait = 30 * time.Second
	}

	server := &http.Server{
		Addr:         cfg.Addr,
		Handler:      cfg.Handler,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Service().Info().Str("service", cfg.Name).Str("addr", cfg.Addr).Msg("http server listening")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return fmt.Errorf("%s server: %w", cfg.Name, err)
	case sig := <-sigCh:
		logger.Service().Info().Str("service", cfg.Name).Str("signal", sig.String()).Msg("shutdown initiated")
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownWait)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("%s shutdown: %w", cfg.Name, err)
	}

	logger.Service().Info().Str("service", cfg.Name).Msg("http server stopped")
	return nil
}
