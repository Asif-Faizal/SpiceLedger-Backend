package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Asif-Faizal/SpiceLedger-Backend/rest"
	"github.com/Asif-Faizal/SpiceLedger-Backend/util"
)

func main() {
	config := util.LoadConfig()
	logger := util.NewLogger(config.LogLevel)

	// Use CONTROL_GRPC_PORT for AccountGrpcURL if not explicitly provided
	accountGrpcURL := os.Getenv("ACCOUNT_GRPC_URL")
	if accountGrpcURL == "" {
		accountGrpcURL = fmt.Sprintf("localhost:%d", config.ControlGrpcPort)
	}

	server, err := rest.NewServer(accountGrpcURL, config.BasicAuthUser, config.BasicAuthPass, logger)
	if err != nil {
		logger.Service().Fatal().Err(err).Msg("failed to initialize REST gateway")
	}
	defer func() {
		if err := server.Close(); err != nil {
			logger.Service().Error().Err(err).Msg("error closing server")
		}
	}()

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", config.RestPort),
		Handler:      rest.NewHandler(server),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Service().Info().Str("addr", httpServer.Addr).Msg("Starting REST gateway")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Service().Fatal().Err(err).Msg("server error")
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Service().Info().Msg("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Service().Fatal().Err(err).Msg("shutdown error")
	}

	logger.Service().Info().Msg("server stopped")
}
