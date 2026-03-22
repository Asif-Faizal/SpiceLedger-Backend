package graphql

import (
	"context"
	"fmt"

	"github.com/99designs/gqlgen/graphql"
	controlpb "github.com/Asif-Faizal/SpiceLedger-Backend/control/pb"
	marketpb "github.com/Asif-Faizal/SpiceLedger-Backend/market/pb"
	"github.com/Asif-Faizal/SpiceLedger-Backend/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// Server represents the high-level GraphQL service instance, managing its own dependencies.
type Server struct {
	controlClient controlpb.ControlServiceClient
	marketClient  marketpb.MarketServiceClient
	controlConn   *grpc.ClientConn
	marketConn    *grpc.ClientConn
	logger        util.Logger
}

// NewServer initializes gRPC connections and returns a professional-grade Server instance.
func NewServer(controlURL, marketURL string, logger util.Logger) (*Server, error) {
	if controlURL == "" {
		return nil, fmt.Errorf("CONTROL_GRPC_URL must be provided for service connectivity")
	}
	if marketURL == "" {
		return nil, fmt.Errorf("MARKET_GRPC_URL must be provided for service connectivity")
	}

	// High-level gRPC interceptor for seamless JWT propagation from GraphQL context.
	authInterceptor := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if accessToken, ok := ctx.Value(util.AccessTokenKey).(string); ok && accessToken != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+accessToken)
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}

	controlConn, err := grpc.Dial(controlURL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(authInterceptor),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to establish gRPC backbone connection to control: %w", err)
	}

	marketConn, err := grpc.Dial(marketURL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(authInterceptor),
	)
	if err != nil {
		controlConn.Close()
		return nil, fmt.Errorf("failed to establish gRPC backbone connection to market: %w", err)
	}

	return &Server{
		controlClient: controlpb.NewControlServiceClient(controlConn),
		marketClient:  marketpb.NewMarketServiceClient(marketConn),
		controlConn:   controlConn,
		marketConn:    marketConn,
		logger:        logger,
	}, nil
}

// Close ensures the service terminates its outbound connections gracefully.
func (s *Server) Close() error {
	if s.controlConn != nil {
		s.controlConn.Close()
	}
	if s.marketConn != nil {
		s.marketConn.Close()
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

func (s *Server) __InputValue() __InputValueResolver {
	return nil
}

func (s *Server) __Type() __TypeResolver {
	return nil
}

func (s *Server) ToExecutableSchema() graphql.ExecutableSchema {
	return NewExecutableSchema(Config{Resolvers: s})
}

type mutationResolver struct{ server *Server }
type queryResolver struct{ server *Server }
