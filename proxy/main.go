package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/Asif-Faizal/SpiceLedger-Backend/util"
)

func main() {
	config := util.LoadConfig()
	logger := util.NewLogger(config.LogLevel)

	// Determine upstream URLs
	accountServiceURL := os.Getenv("ACCOUNT_SERVICE_URL")
	if accountServiceURL == "" {
		accountServiceURL = fmt.Sprintf("http://localhost:%d", config.RestPort)
	}

	graphqlGatewayURL := os.Getenv("GRAPHQL_GATEWAY_URL")
	if graphqlGatewayURL == "" {
		graphqlGatewayURL = fmt.Sprintf("http://localhost:%d", config.GraphqlPort)
	}

	logger.Service().Info().
		Str("AccountServiceURL", accountServiceURL).
		Str("GraphQLGatewayURL", graphqlGatewayURL).
		Msg("Proxy Starting")

	// Parse upstream targets
	authTarget, err := url.Parse(accountServiceURL)
	if err != nil {
		logger.Service().Fatal().Err(err).Msg("Failed to parse account URL")
	}

	gqlTarget, err := url.Parse(graphqlGatewayURL)
	if err != nil {
		logger.Service().Fatal().Err(err).Msg("Failed to parse graphql URL")
	}

	// Reverse proxies
	authProxy := httputil.NewSingleHostReverseProxy(authTarget)
	gqlProxy := httputil.NewSingleHostReverseProxy(gqlTarget)

	http.HandleFunc("/rest/", func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/rest")
		authProxy.ServeHTTP(w, r)
	})

	http.HandleFunc("/graphql", func(w http.ResponseWriter, r *http.Request) {
		gqlProxy.ServeHTTP(w, r)
	})

	http.HandleFunc("/playground", func(w http.ResponseWriter, r *http.Request) {
		gqlProxy.ServeHTTP(w, r)
	})

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", config.ProxyPort),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	logger.Service().Info().Str("addr", server.Addr).Msg("Starting reverse proxy")
	logger.Service().Fatal().Err(server.ListenAndServe()).Msg("Proxy exit")
}
