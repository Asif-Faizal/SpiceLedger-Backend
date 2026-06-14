package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Asif-Faizal/SpiceLedger-Backend/util"
)

func main() {
	config := util.LoadConfig()
	logger := util.NewLogger(config.LogLevel)

	accountServiceURL := config.ResolveAccountServiceURL()
	graphqlGatewayURL := config.ResolveGraphqlGatewayURL()

	logger.Service().Info().
		Str("AccountServiceURL", accountServiceURL).
		Str("GraphQLGatewayURL", graphqlGatewayURL).
		Msg("Proxy Starting")

	authTarget, err := url.Parse(accountServiceURL)
	if err != nil {
		logger.Service().Fatal().Err(err).Msg("Failed to parse account URL")
	}

	gqlTarget, err := url.Parse(graphqlGatewayURL)
	if err != nil {
		logger.Service().Fatal().Err(err).Msg("Failed to parse graphql URL")
	}

	authProxy := httputil.NewSingleHostReverseProxy(authTarget)
	gqlProxy := httputil.NewSingleHostReverseProxy(gqlTarget)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"service":"proxy","status":"operational"}`)
	})
	mux.HandleFunc("/rest/", func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/rest")
		authProxy.ServeHTTP(w, r)
	})
	mux.HandleFunc("/graphql", func(w http.ResponseWriter, r *http.Request) {
		gqlProxy.ServeHTTP(w, r)
	})
	mux.HandleFunc("/playground", func(w http.ResponseWriter, r *http.Request) {
		gqlProxy.ServeHTTP(w, r)
	})

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", config.ProxyPort),
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	go func() {
		logger.Service().Info().Str("addr", server.Addr).Msg("Starting reverse proxy")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Service().Fatal().Err(err).Msg("Proxy exit")
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	logger.Service().Info().Msg("shutting down proxy")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Service().Fatal().Err(err).Msg("proxy shutdown error")
	}
}
