package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/Asif-Faizal/SpiceLedger-Backend/graphql"
	"github.com/Asif-Faizal/SpiceLedger-Backend/util"
)

func main() {
	// Professional configuration orchestration
	config := util.LoadConfig()

	logger := util.NewLogger(config.LogLevel)

	// Initialize the high-level GraphQL server instance
	// Use CONTROL_GRPC_PORT for AccountGrpcURL if not explicitly provided
	accountGrpcURL := os.Getenv("ACCOUNT_GRPC_URL")
	if accountGrpcURL == "" {
		accountGrpcURL = fmt.Sprintf("127.0.0.1:%d", config.ControlGrpcPort)
	}

	marketGrpcURL := os.Getenv("MARKET_GRPC_URL")
	if marketGrpcURL == "" {
		marketGrpcURL = fmt.Sprintf("127.0.0.1:%d", config.MarketGrpcPort)
	}

	appServer, err := graphql.NewServer(accountGrpcURL, marketGrpcURL, logger)
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

	// Setup server with gracefull shutdown support
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", config.GraphqlPort),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
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
