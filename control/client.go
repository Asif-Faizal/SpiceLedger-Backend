package control

import (
	"context"

	pb "github.com/Asif-Faizal/SpiceLedger-Backend/control/pb"
	"google.golang.org/grpc"
)

type ControlClient struct {
	connection *grpc.ClientConn
	client     pb.ControlServiceClient
}

func NewControlClient(url string) (*ControlClient, error) {
	connection, err := grpc.Dial(url, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return &ControlClient{
		connection: connection,
		client:     pb.NewControlServiceClient(connection),
	}, nil
}

func (client *ControlClient) Close() {
	client.connection.Close()
}

func (client *ControlClient) CheckEmailExists(ctx context.Context, email string) (*pb.CheckEmailExistsResponse, error) {
	response, err := client.client.CheckEmailExists(ctx, &pb.CheckEmailExistsRequest{
		Email: email,
	})
	if err != nil {
		return nil, err
	}
	return response, nil
}

// CreateOrUpdate Account
func (client *ControlClient) CreateOrUpdateAccount(ctx context.Context, id, name, userType, email, password string) (*pb.CreateOrUpdateAccountResponse, error) {
	response, err := client.client.CreateOrUpdateAccount(ctx, &pb.CreateOrUpdateAccountRequest{
		Id:       id,
		Name:     name,
		Usertype: userType,
		Email:    email,
		Password: password,
	})
	if err != nil {
		return nil, err
	}
	return response, nil
}

// Get Account by ID
func (client *ControlClient) GetAccountByID(ctx context.Context, id string) (*pb.GetAccountByIDResponse, error) {
	response, err := client.client.GetAccountByID(ctx, &pb.GetAccountByIDRequest{
		Id: id,
	})
	if err != nil {
		return nil, err
	}
	return response, nil
}

// List Accounts
func (client *ControlClient) ListAccounts(ctx context.Context, skip uint32, take uint32) (*pb.ListAccountsResponse, error) {
	response, err := client.client.ListAccounts(ctx, &pb.ListAccountsRequest{
		Skip: skip,
		Take: take,
	})
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (client *ControlClient) Login(ctx context.Context, email, password, deviceID string) (*pb.LoginResponse, error) {
	response, err := client.client.Login(ctx, &pb.LoginRequest{
		Email:    email,
		Password: password,
		DeviceId: deviceID,
	})
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (client *ControlClient) Logout(ctx context.Context, accessToken, deviceID string) (*pb.LogoutResponse, error) {
	response, err := client.client.Logout(ctx, &pb.LogoutRequest{
		AccessToken: accessToken,
		DeviceId:    deviceID,
	})
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (client *ControlClient) RefreshToken(ctx context.Context, refreshToken, deviceID string) (*pb.RefreshTokenResponse, error) {
	response, err := client.client.RefreshToken(ctx, &pb.RefreshTokenRequest{
		RefreshToken: refreshToken,
		DeviceId:     deviceID,
	})
	if err != nil {
		return nil, err
	}
	return response, nil
}
