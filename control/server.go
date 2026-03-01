package control

import (
	"context"
	"fmt"
	"net"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "github.com/Asif-Faizal/SpiceLedger-Backend/control/pb"
	"github.com/Asif-Faizal/SpiceLedger-Backend/util"
)

type GrpcServer struct {
	accountService Service
	logger         util.Logger
	pb.UnimplementedControlServiceServer
}

func ListenGrpcServer(service Service, logger util.Logger, port int, jwtSecret string) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			util.UnaryServerInterceptor(logger),
			util.AuthInterceptor(jwtSecret),
		)),
	)

	server := &GrpcServer{
		accountService: service,
		logger:         logger,
	}
	pb.RegisterControlServiceServer(grpcServer, server)
	reflection.Register(grpcServer)

	logger.Transport().Info().Int("port", port).Msg("gRPC server listening")
	return grpcServer.Serve(lis)
}

func (server *GrpcServer) CheckEmailExists(ctx context.Context, request *pb.CheckEmailExistsRequest) (*pb.CheckEmailExistsResponse, error) {
	exists, err := server.accountService.CheckEmailExists(ctx, request.Email)
	if err != nil {
		return nil, err
	}
	return &pb.CheckEmailExistsResponse{Exists: exists}, nil
}

func (server *GrpcServer) CreateOrUpdateAccount(ctx context.Context, request *pb.CreateOrUpdateAccountRequest) (*pb.CreateOrUpdateAccountResponse, error) {
	if request.Usertype == "" {
		return nil, fmt.Errorf("usertype is required")
	}
	if request.Email == "" {
		return nil, fmt.Errorf("email is required")
	}

	account, err := server.accountService.CreateOrUpdateAccount(ctx, &Account{
		ID:       request.Id,
		Name:     request.Name,
		UserType: request.Usertype,
		Email:    request.Email,
		Password: request.Password,
	})
	if err != nil {
		return nil, err
	}
	return &pb.CreateOrUpdateAccountResponse{
		Account: &pb.Account{
			Id:       account.ID,
			Name:     account.Name,
			Usertype: account.UserType,
			Email:    account.Email,
		},
	}, nil
}

func (server *GrpcServer) GetAccountByID(ctx context.Context, request *pb.GetAccountByIDRequest) (*pb.GetAccountByIDResponse, error) {
	account, err := server.accountService.GetAccountByID(ctx, request.Id)
	if err != nil {
		return nil, err
	}
	return &pb.GetAccountByIDResponse{
		Account: &pb.Account{
			Id:       account.ID,
			Name:     account.Name,
			Usertype: account.UserType,
			Email:    account.Email,
		},
	}, nil
}

func (server *GrpcServer) ListAccounts(ctx context.Context, request *pb.ListAccountsRequest) (*pb.ListAccountsResponse, error) {
	domainAccounts, err := server.accountService.ListAccounts(ctx, uint(request.Skip), uint(request.Take))
	if err != nil {
		return nil, err
	}
	accounts := []*pb.Account{}
	for _, account := range domainAccounts {
		accounts = append(accounts, &pb.Account{
			Id:       account.ID,
			Name:     account.Name,
			Usertype: account.UserType,
			Email:    account.Email,
		})
	}
	return &pb.ListAccountsResponse{Accounts: accounts}, nil
}

func (server *GrpcServer) Login(ctx context.Context, request *pb.LoginRequest) (*pb.LoginResponse, error) {
	resp, err := server.accountService.Login(ctx, request.Email, request.Password, request.DeviceId)
	if err != nil {
		return nil, err
	}

	var account *pb.Account
	if resp.Account != nil {
		account = &pb.Account{
			Id:       resp.Account.ID,
			Name:     resp.Account.Name,
			Usertype: resp.Account.UserType,
			Email:    resp.Account.Email,
		}
	}

	return &pb.LoginResponse{
		Account:      account,
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
	}, nil
}

func (server *GrpcServer) Logout(ctx context.Context, request *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	if err := server.accountService.Logout(ctx, request.AccessToken, request.DeviceId); err != nil {
		return nil, err
	}
	return &pb.LogoutResponse{Success: true}, nil
}

func (server *GrpcServer) RefreshToken(ctx context.Context, request *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	resp, err := server.accountService.RefreshToken(ctx, request.RefreshToken, request.DeviceId)
	if err != nil {
		return nil, err
	}

	var account *pb.Account
	if resp.Account != nil {
		account = &pb.Account{
			Id:       resp.Account.ID,
			Name:     resp.Account.Name,
			Usertype: resp.Account.UserType,
			Email:    resp.Account.Email,
		}
	}

	return &pb.RefreshTokenResponse{
		Account:      account,
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
	}, nil
}
