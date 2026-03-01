package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Asif-Faizal/SpiceLedger-Backend/rest"
	"github.com/Asif-Faizal/SpiceLedger-Backend/util"
	"github.com/kelseyhightower/envconfig"
)

type AppConfig struct {
	AccountGrpcURL string `envconfig:"ACCOUNT_GRPC_URL" default:"localhost:50051"`
	Port           int    `envconfig:"PORT" default:"8082"`
	Env            string `envconfig:"ENVIRONMENT" default:"development"`
	LogLevel       string `envconfig:"LOG_LEVEL" default:"info"`
	BasicAuthUser  string `envconfig:"BASIC_AUTH_USER" default:"admin"`
	BasicAuthPass  string `envconfig:"BASIC_AUTH_PASS" default:"admin"`
}

func main() {
	var cfg AppConfig
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	logger := util.NewLogger(cfg.LogLevel)

	server, err := rest.NewServer(cfg.AccountGrpcURL, cfg.BasicAuthUser, cfg.BasicAuthPass, logger)
	if err != nil {
		logger.Service().Fatal().Err(err).Msg("failed to initialize REST gateway")
	}
	defer func() {
		if err := server.Close(); err != nil {
			logger.Service().Error().Err(err).Msg("error closing server")
		}
	}()

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      rest.NewHandler(server),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Service().Info().Str("addr", httpServer.Addr).Str("env", cfg.Env).Msg("Starting REST gateway")
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
