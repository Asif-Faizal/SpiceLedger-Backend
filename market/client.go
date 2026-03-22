package market

import (
	"context"

	pb "github.com/Asif-Faizal/SpiceLedger-Backend/market/pb"
	"google.golang.org/grpc"
)

type MarketClient struct {
	connection *grpc.ClientConn
	client     pb.MarketServiceClient
}

func NewMarketClient(url string) (*MarketClient, error) {
	connection, err := grpc.Dial(url, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return &MarketClient{
		connection: connection,
		client:     pb.NewMarketServiceClient(connection),
	}, nil
}

func (c *MarketClient) Close() error {
	return c.connection.Close()
}

func (c *MarketClient) Buy(ctx context.Context, userID, spiceGradeID string, quantity, price float64, tradeDate string) (*pb.BuyResponse, error) {
	return c.client.Buy(ctx, &pb.BuyRequest{
		UserId:       userID,
		SpiceGradeId: spiceGradeID,
		Quantity:     quantity,
		Price:        price,
		TradeDate:    tradeDate,
	})
}

func (c *MarketClient) Sell(ctx context.Context, userID, spiceGradeID string, quantity, price float64, tradeDate string) (*pb.SellResponse, error) {
	return c.client.Sell(ctx, &pb.SellRequest{
		UserId:       userID,
		SpiceGradeId: spiceGradeID,
		Quantity:     quantity,
		Price:        price,
		TradeDate:    tradeDate,
	})
}

func (c *MarketClient) GetPosition(ctx context.Context, userID, spiceGradeID string) (*pb.GetPositionResponse, error) {
	return c.client.GetPosition(ctx, &pb.GetPositionRequest{
		UserId:       userID,
		SpiceGradeId: spiceGradeID,
	})
}

func (c *MarketClient) ListTransactions(ctx context.Context, userID, spiceGradeID string, skip, take uint32) (*pb.ListTransactionsResponse, error) {
	return c.client.ListTransactions(ctx, &pb.ListTransactionsRequest{
		UserId:       userID,
		SpiceGradeId: spiceGradeID,
		Skip:         skip,
		Take:         take,
	})
}
