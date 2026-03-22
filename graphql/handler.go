package graphql

import (
	"context"
	"net/http"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/Asif-Faizal/SpiceLedger-Backend/util"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type contextKey string

const responseWriterKey contextKey = "responseWriter"

// NewHandler configures and returns the GraphQL HTTP handler with professional middleware
func NewHandler(s *Server) http.Handler {
	srv := handler.New(s.ToExecutableSchema())

	// Standard transports
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.MultipartForm{})

	srv.Use(extension.Introspection{})

	// Error Presenter translates gRPC errors to clean GraphQL errors
	srv.SetErrorPresenter(func(ctx context.Context, err error) *gqlerror.Error {
		gqlErr := graphql.DefaultErrorPresenter(ctx, err)
		if st, ok := status.FromError(err); ok {
			if gqlErr.Extensions == nil {
				gqlErr.Extensions = make(map[string]interface{})
			}
			gqlErr.Extensions["grpc_code"] = st.Code().String()
		}
		return gqlErr
	})

	// Response Interceptor for handling HTTP status codes (e.g., 401 for Auth errors)
	srv.AroundResponses(func(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
		resp := next(ctx)
		if len(resp.Errors) > 0 {
			for _, err := range resp.Errors {
				if st, ok := status.FromError(err.Unwrap()); ok && st.Code() == codes.Unauthenticated {
					if w, ok := ctx.Value(responseWriterKey).(http.ResponseWriter); ok {
						w.WriteHeader(http.StatusUnauthorized)
					}
					break
				}
			}
		}
		return resp
	})

	return wrapWithMiddleware(srv)
}

func wrapWithMiddleware(h http.Handler) http.Handler {
	return writerMiddleware(authMiddleware(h))
}

func authMiddleware(next http.Handler) http.Handler {
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

func writerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), responseWriterKey, w)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
