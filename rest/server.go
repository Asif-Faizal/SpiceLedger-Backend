package rest

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/Asif-Faizal/SpiceLedger-Backend/control"
	"github.com/Asif-Faizal/SpiceLedger-Backend/util"
	"google.golang.org/grpc/metadata"
)

type Server struct {
	accountClient *control.ControlClient
	logger        util.Logger
	basicUser     string
	basicPass     string
}

func NewServer(accountGrpcURL string, basicUser, basicPass string, logger util.Logger) (*Server, error) {
	if accountGrpcURL == "" {
		return nil, fmt.Errorf("ACCOUNT_GRPC_URL must be provided")
	}

	accountClient, err := control.NewControlClient(accountGrpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to account service: %w", err)
	}

	return &Server{
		accountClient: accountClient,
		logger:        logger,
		basicUser:     basicUser,
		basicPass:     basicPass,
	}, nil
}

func (s *Server) authCtx(ctx context.Context) context.Context {
	auth := s.basicUser + ":" + s.basicPass
	enc := base64.StdEncoding.EncodeToString([]byte(auth))
	return metadata.AppendToOutgoingContext(ctx, "authorization", "Basic "+enc)
}

func (s *Server) Close() error {
	if s.accountClient != nil {
		s.accountClient.Close()
	}
	return nil
}
