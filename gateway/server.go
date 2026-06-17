package gateway

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/Asif-Faizal/SpiceLedger-Backend/graphql"
	"github.com/Asif-Faizal/SpiceLedger-Backend/rest"
	"github.com/Asif-Faizal/SpiceLedger-Backend/util"
)

// Dependencies holds live connections owned by the API gateway process.
type Dependencies struct {
	REST     *rest.Server
	GraphQL  *graphql.Server
	closers  []func() error
}

// Close releases outbound gRPC connections.
func (d *Dependencies) Close() error {
	var first error
	for _, closeFn := range d.closers {
		if err := closeFn(); err != nil && first == nil {
			first = err
		}
	}
	return first
}

// NewDependencies wires REST and GraphQL gateways to upstream gRPC services.
func NewDependencies(cfg *util.Config, logger util.Logger) (*Dependencies, error) {
	restServer, err := rest.NewServer(
		cfg.ResolveAccountGrpcURL(),
		cfg.BasicAuthUser,
		cfg.BasicAuthPass,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("rest gateway: %w", err)
	}

	gqlServer, err := graphql.NewServer(
		cfg.ResolveAccountGrpcURL(),
		cfg.ResolveMarketGrpcURL(),
		logger,
	)
	if err != nil {
		_ = restServer.Close()
		return nil, fmt.Errorf("graphql gateway: %w", err)
	}

	return &Dependencies{
		REST:    restServer,
		GraphQL: gqlServer,
		closers: []func() error{restServer.Close, gqlServer.Close},
	}, nil
}

// NewHandler returns the unified edge HTTP handler for REST + GraphQL on one port.
func NewHandler(deps *Dependencies) http.Handler {
	mux := http.NewServeMux()

	restHandler := rest.NewHandler(deps.REST)
	mux.Handle("/rest/", http.StripPrefix("/rest", restHandler))
	mux.Handle("/graphql", graphql.NewHandler(deps.GraphQL))
	mux.Handle("/playground", playground.Handler("SpiceLedger GraphQL", "/graphql"))

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":true,"message":"","data":{"service":"gateway","status":"operational"}}`))
	})

	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":true,"message":"","data":{"ready":true}}`))
	})

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/playground", http.StatusFound)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/rest") ||
			strings.HasPrefix(r.URL.Path, "/graphql") ||
			r.URL.Path == "/playground" ||
			r.URL.Path == "/health" ||
			r.URL.Path == "/ready" {
			mux.ServeHTTP(w, r)
			return
		}
		util.WriteJSONResponse(w, http.StatusNotFound, false, "route not found", nil)
	})
}
