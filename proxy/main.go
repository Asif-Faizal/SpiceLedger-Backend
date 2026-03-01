package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/Asif-Faizal/SpiceLedger-Backend/util"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	AccountServiceURL   string        `envconfig:"ACCOUNT_SERVICE_URL" required:"true"`
	GraphQLGatewayURL   string        `envconfig:"GRAPHQL_GATEWAY_URL" required:"true"`
	MaxIdleConns        int           `envconfig:"PROXY_MAX_IDLE_CONNS" default:"100"`
	MaxIdleConnsPerHost int           `envconfig:"PROXY_MAX_IDLE_CONNS_PER_HOST" default:"100"`
	IdleConnTimeout     time.Duration `envconfig:"PROXY_IDLE_CONN_TIMEOUT" default:"90s"`
	RequestTimeout      time.Duration `envconfig:"PROXY_REQUEST_TIMEOUT" default:"30s"`
	Port                int           `envconfig:"PROXY_PORT" default:"80"`
	LogLevel            string        `envconfig:"LOG_LEVEL" default:"info"`
}

func main() {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatalf("[PROXY] Failed to process config: %v", err)
	}
	logger := util.NewLogger(cfg.LogLevel)

	logger.Service().Info().
		Str("AccountServiceURL", cfg.AccountServiceURL).
		Str("GraphQLGatewayURL", cfg.GraphQLGatewayURL).
		Int("MaxIdleConns", cfg.MaxIdleConns).
		Int("MaxIdleConnsPerHost", cfg.MaxIdleConnsPerHost).
		Dur("IdleConnTimeout", cfg.IdleConnTimeout).
		Dur("RequestTimeout", cfg.RequestTimeout).
		Msg("Proxy Starting")

	// Parse upstream targets
	authTarget, err := url.Parse(cfg.AccountServiceURL)
	if err != nil {
		logger.Service().Fatal().Err(err).Msg("Failed to parse account URL")
	}

	gqlTarget, err := url.Parse(cfg.GraphQLGatewayURL)
	if err != nil {
		logger.Service().Fatal().Err(err).Msg("Failed to parse graphql URL")
	}

	// Configure transport with connection pooling
	transport := &http.Transport{
		MaxIdleConns:        cfg.MaxIdleConns,
		MaxIdleConnsPerHost: cfg.MaxIdleConnsPerHost,
		IdleConnTimeout:     cfg.IdleConnTimeout,
	}

	// Create reverse proxies
	restProxy := httputil.NewSingleHostReverseProxy(authTarget)
	restProxy.Transport = transport
	restProxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		logger.Transport().Error().Err(err).Str("method", r.Method).Str("path", r.URL.Path).Msg("Proxy Error")
		http.Error(w, "Service temporarily unavailable", http.StatusServiceUnavailable)
	}

	gqlProxy := httputil.NewSingleHostReverseProxy(gqlTarget)
	gqlProxy.Transport = transport
	gqlProxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		logger.Transport().Error().Err(err).Str("method", r.Method).Str("path", r.URL.Path).Msg("Proxy Error")
		http.Error(w, "Service temporarily unavailable", http.StatusServiceUnavailable)
	}

	// Route handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Log incoming request
		logger.Transport().Info().Str("method", r.Method).Str("path", r.URL.Path).Msg("Proxying request")

		// Route based on path
		if strings.HasPrefix(r.URL.Path, "/rest") {
			r.URL.Path = strings.TrimPrefix(r.URL.Path, "/rest")
			restProxy.ServeHTTP(w, r)
		} else {
			// Everything else goes to GraphQL
			gqlProxy.ServeHTTP(w, r)
		}
	})

	// Add timeout wrapper
	timeoutHandler := http.TimeoutHandler(handler, cfg.RequestTimeout, "Request timeout")

	addr := fmt.Sprintf(":%d", cfg.Port)
	logger.Service().Info().Str("addr", addr).Msg("Starting reverse proxy")
	logger.Service().Fatal().Err(http.ListenAndServe(addr, timeoutHandler)).Msg("Proxy exit")
}
