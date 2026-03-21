package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/Asif-Faizal/SpiceLedger-Backend/graphql"
	"github.com/Asif-Faizal/SpiceLedger-Backend/util"
	"strings"
)

func main() {
	// Load configuration
	config := util.LoadConfig()

	// Initialize Logger
	logger := util.NewLogger(config.LogLevel)

	controlURL := os.Getenv("ACCOUNT_GRPC_URL")
	if controlURL == "" {
		controlURL = "localhost:50051"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	// Initialize GraphQL server
	server, err := graphql.NewGraphQLServer(controlURL)
	if err != nil {
		logger.Service().Fatal().Err(err).Msg("failed to initialize GraphQL server")
	}
	defer func() {
		if err := server.Close(); err != nil {
			logger.Service().Error().Err(err).Msg("error closing server")
		}
	}()

	// Create HTTP server
	mux := http.NewServeMux()

	// GraphQL endpoint with Auth Middleware
	mux.Handle("/graphql", AuthMiddleware(handler.NewDefaultServer(server.ToExecutableSchema())))

	// Playground endpoint
	mux.Handle("/playground", playground.Handler("GraphQL Playground", "/graphql"))

	// Health check endpoint
	mux.HandleFunc("/health", healthCheckHandler)

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Service().Info().Str("addr", httpServer.Addr).Msg("Starting GraphQL server")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Service().Fatal().Err(err).Msg("server error")
		}
	}()

	// Graceful shutdown
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

// healthCheckHandler returns OK if the server is running
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"ok"}`)
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			next.ServeHTTP(w, r)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		ctx := context.WithValue(r.Context(), util.AccessTokenKey, tokenString)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
