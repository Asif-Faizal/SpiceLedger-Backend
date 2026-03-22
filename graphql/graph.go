package graphql

import (
	"context"
	"fmt"

	"github.com/99designs/gqlgen/graphql"
	"github.com/Asif-Faizal/SpiceLedger-Backend/control/pb"
	"github.com/Asif-Faizal/SpiceLedger-Backend/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// Server represents the high-level GraphQL service instance, managing its own dependencies.
type Server struct {
	controlClient pb.ControlServiceClient
	conn          *grpc.ClientConn
	logger        util.Logger
}

// NewServer initializes gRPC connections and returns a professional-grade Server instance.
func NewServer(controlURL string, logger util.Logger) (*Server, error) {
	if controlURL == "" {
		return nil, fmt.Errorf("ACCOUNT_GRPC_URL must be provided for service connectivity")
	}

	// High-level gRPC interceptor for seamless JWT propagation from GraphQL context.
	authInterceptor := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if accessToken, ok := ctx.Value(util.AccessTokenKey).(string); ok && accessToken != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+accessToken)
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}

	conn, err := grpc.Dial(controlURL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(authInterceptor),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to establish gRPC backbone connection: %w", err)
	}

	return &Server{
		controlClient: pb.NewControlServiceClient(conn),
		conn:          conn,
		logger:        logger,
	}, nil
}

// Close ensures the service terminates its outbound connections gracefully.
func (s *Server) Close() error {
	if s.conn != nil {
		return s.conn.Close()
	}
	return nil
}

// plumbing for gqlgen ResolverRoot interface

func (s *Server) Mutation() MutationResolver {
	return &mutationResolver{server: s}
}

func (s *Server) Query() QueryResolver {
	return &queryResolver{server: s}
}

func (s *Server) ToExecutableSchema() graphql.ExecutableSchema {
	return NewExecutableSchema(Config{Resolvers: s})
}

type mutationResolver struct{ server *Server }
type queryResolver struct{ server *Server }
