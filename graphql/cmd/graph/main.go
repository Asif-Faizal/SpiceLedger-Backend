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

	"github.com/Asif-Faizal/SpiceLedger-Backend/graphql"
	"github.com/Asif-Faizal/SpiceLedger-Backend/util"
	"github.com/kelseyhightower/envconfig"
	"github.com/99designs/gqlgen/graphql/playground"
)

// AppConfig defines the orchestration parameters for the high-level service instance.
type AppConfig struct {
	AccountGrpcURL string `envconfig:"ACCOUNT_GRPC_URL" default:"localhost:50051"`
	Port           int    `envconfig:"PORT" default:"8081"`
	LogLevel       string `envconfig:"LOG_LEVEL" default:"info"`
}

func main() {
	// Professional configuration orchestration
	var cfg AppConfig
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatalf("Critical core configuration failure: %v", err)
	}
	
	logger := util.NewLogger(cfg.LogLevel)

	// Initialize the high-level GraphQL server instance
	appServer, err := graphql.NewServer(cfg.AccountGrpcURL, logger)
	if err != nil {
		logger.Service().Fatal().Err(err).Msg("Core service initialization failed")
	}
	defer appServer.Close()

	// High-level HTTP orchestration matching REST pattern
	mux := http.NewServeMux()
	mux.Handle("/graphql", graphql.NewHandler(appServer))
	mux.Handle("/playground", playground.Handler("GraphQL Interface", "/graphql"))
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"operational"}`)
	})

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Service execution in background
	go func() {
		logger.Service().Info().Str("addr", httpServer.Addr).Msg("Professional GraphQL entrypoint active")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Service().Fatal().Err(err).Msg("Fatal transport-layer error")
		}
	}()

	// Signal handling for graceful termination
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	logger.Service().Info().Msg("Initiating graceful shutdown sequence")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Service().Error().Err(err).Msg("Shutdown orchestration error")
	}
	logger.Service().Info().Msg("Service termination complete")
}
