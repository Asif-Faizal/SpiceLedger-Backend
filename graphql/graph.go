package graphql

import (
	"context"
	"fmt"
	"log"

	"github.com/99designs/gqlgen/graphql"
	"github.com/Asif-Faizal/SpiceLedger-Backend/control/pb"
	"github.com/Asif-Faizal/SpiceLedger-Backend/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type Server struct {
	controlClient pb.ControlServiceClient
	conn          *grpc.ClientConn
}

// NewGraphQLServer initializes and returns a new GraphQL server with all microservice clients
func NewGraphQLServer(controlURL string) (*Server, error) {
	// Validate URLs
	if controlURL == "" {
		return nil, fmt.Errorf("control service URL must be provided")
	}

	log.Printf("Connecting to control service at %s", controlURL)

	// Initialize Control Client with Auth Interceptor
	conn, err := grpc.Dial(controlURL, 
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
			accessToken, _ := ctx.Value(util.AccessTokenKey).(string)
			if accessToken != "" {
				ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+accessToken)
			}
			return invoker(ctx, method, req, reply, cc, opts...)
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to control service: %w", err)
	}

	controlClient := pb.NewControlServiceClient(conn)

	return &Server{
		controlClient: controlClient,
		conn:          conn,
	}, nil
}

// Close gracefully closes all client connections
func (s *Server) Close() error {
	if s.conn != nil {
		s.conn.Close()
	}
	return nil
}

func (s *Server) Mutation() MutationResolver {
	return &mutationResolver{
		server: s,
	}
}

func (s *Server) Query() QueryResolver {
	return &queryResolver{
		server: s,
	}
}

func (s *Server) ToExecutableSchema() graphql.ExecutableSchema {
	return NewExecutableSchema(Config{
		Resolvers: s,
	})
}

type mutationResolver struct{ server *Server }
type queryResolver struct{ server *Server }
