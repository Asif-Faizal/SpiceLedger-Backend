package market

import (
	"context"
	"fmt"
	"net"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "github.com/Asif-Faizal/SpiceLedger-Backend/market/pb"
	"github.com/Asif-Faizal/SpiceLedger-Backend/util"
)

type GrpcServer struct {
	marketService Service
	logger        util.Logger
	config        *util.Config
	pb.UnimplementedMarketServiceServer
}

func ListenGrpcServer(service Service, logger util.Logger, config *util.Config) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", config.MarketGrpcPort))
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			util.UnaryServerInterceptor(logger),
			util.AuthInterceptor(config.JWTSecret, config.BasicAuthUser, config.BasicAuthPass),
		)),
	)

	server := &GrpcServer{
		marketService: service,
		logger:        logger,
		config:        config,
	}
	pb.RegisterMarketServiceServer(grpcServer, server)
	reflection.Register(grpcServer)

	logger.Transport().Info().Int("port", config.MarketGrpcPort).Msg("gRPC server listening")
	return grpcServer.Serve(lis)
}

func (server *GrpcServer) Buy(ctx context.Context, req *pb.BuyRequest) (*pb.BuyResponse, error) {
	tradeDate, err := time.Parse("2006-01-02", req.TradeDate)
	if err != nil {
		tradeDate = time.Now()
	}

	userID := req.UserId
	if userID == "" {
		if id, ok := ctx.Value(util.AccountIDKey).(string); ok {
			userID = id
		}
	}

	txn, err := server.marketService.Buy(ctx, userID, req.SpiceGradeId, req.Quantity, req.Price, tradeDate)
	if err != nil {
		return nil, err
	}

	return &pb.BuyResponse{
		Transaction: &pb.Transaction{
			Id:           txn.ID,
			UserId:       txn.UserID,
			SpiceGradeId: txn.SpiceGradeID,
			Type:         txn.Type,
			Quantity:     txn.Quantity,
			Price:        txn.Price,
			TradeDate:    txn.TradeDate.Format("2006-01-02"),
			CreatedAt:    txn.CreatedAt.Format("2006-01-02 15:04:05"),
		},
	}, nil
}

func (server *GrpcServer) Sell(ctx context.Context, req *pb.SellRequest) (*pb.SellResponse, error) {
	tradeDate, err := time.Parse("2006-01-02", req.TradeDate)
	if err != nil {
		tradeDate = time.Now()
	}

	userID := req.UserId
	if userID == "" {
		if id, ok := ctx.Value(util.AccountIDKey).(string); ok {
			userID = id
		}
	}

	txn, err := server.marketService.Sell(ctx, userID, req.SpiceGradeId, req.Quantity, req.Price, tradeDate)
	if err != nil {
		return nil, err
	}

	return &pb.SellResponse{
		Transaction: &pb.Transaction{
			Id:           txn.ID,
			UserId:       txn.UserID,
			SpiceGradeId: txn.SpiceGradeID,
			Type:         txn.Type,
			Quantity:     txn.Quantity,
			Price:        txn.Price,
			TradeDate:    txn.TradeDate.Format("2006-01-02"),
			CreatedAt:    txn.CreatedAt.Format("2006-01-02 15:04:05"),
		},
	}, nil
}

func (server *GrpcServer) GetGradePosition(ctx context.Context, req *pb.GetGradePositionRequest) (*pb.GetGradePositionResponse, error) {
	userID := req.UserId
	if userID == "" {
		if id, ok := ctx.Value(util.AccountIDKey).(string); ok {
			userID = id
		}
	}

	pos, err := server.marketService.GetGradePosition(ctx, userID, req.SpiceGradeId)
	if err != nil {
		return nil, err
	}

	return &pb.GetGradePositionResponse{
		Position: &pb.PositionView{
			UserId:        pos.UserID,
			SpiceGradeId:  pos.SpiceGradeID,
			TotalQty:      pos.TotalQty,
			TotalCost:     pos.TotalCost,
			AvgCost:       pos.AvgCost,
			TodayPrice:    pos.TodayPrice,
			RealizedPnl:   pos.RealizedPnL,
			UnrealizedPnl: pos.UnrealizedPnL,
			UpdatedAt:     pos.UpdatedAt.Format("2006-01-02 15:04:05"),
		},
	}, nil
}

func (server *GrpcServer) GetPositions(ctx context.Context, req *pb.GetPositionsRequest) (*pb.GetPositionsResponse, error) {
	userID := req.UserId
	if userID == "" {
		if id, ok := ctx.Value(util.AccountIDKey).(string); ok {
			userID = id
		}
	}

	positions, err := server.marketService.GetPositions(ctx, userID)
	if err != nil {
		return nil, err
	}

	var pbPositions []*pb.PositionView
	for _, pos := range positions {
		pbPositions = append(pbPositions, &pb.PositionView{
			UserId:        pos.UserID,
			SpiceGradeId:  pos.SpiceGradeID,
			TotalQty:      pos.TotalQty,
			TotalCost:     pos.TotalCost,
			AvgCost:       pos.AvgCost,
			TodayPrice:    pos.TodayPrice,
			RealizedPnl:   pos.RealizedPnL,
			UnrealizedPnl: pos.UnrealizedPnL,
			UpdatedAt:     pos.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return &pb.GetPositionsResponse{
		Positions: pbPositions,
	}, nil
}

func (server *GrpcServer) ListGradeTransactions(ctx context.Context, req *pb.ListGradeTransactionsRequest) (*pb.ListGradeTransactionsResponse, error) {
	userID := req.UserId
	if userID == "" {
		if id, ok := ctx.Value(util.AccountIDKey).(string); ok {
			userID = id
		}
	}

	txns, err := server.marketService.ListGradeTransactions(ctx, userID, req.SpiceGradeId, uint(req.Skip), uint(req.Take))
	if err != nil {
		return nil, err
	}

	var protoTxns []*pb.Transaction
	for _, txn := range txns {
		protoTxns = append(protoTxns, &pb.Transaction{
			Id:           txn.ID,
			UserId:       txn.UserID,
			SpiceGradeId: txn.SpiceGradeID,
			Type:         txn.Type,
			Quantity:     txn.Quantity,
			Price:        txn.Price,
			TradeDate:    txn.TradeDate.Format("2006-01-02"),
			CreatedAt:    txn.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return &pb.ListGradeTransactionsResponse{
		Transactions: protoTxns,
	}, nil
}

func (server *GrpcServer) ListTransactions(ctx context.Context, req *pb.ListTransactionsRequest) (*pb.ListTransactionsResponse, error) {
	userID := req.UserId
	if userID == "" {
		if id, ok := ctx.Value(util.AccountIDKey).(string); ok {
			userID = id
		}
	}

	txns, err := server.marketService.ListTransactions(ctx, userID, uint(req.Skip), uint(req.Take))
	if err != nil {
		return nil, err
	}

	var protoTxns []*pb.Transaction
	for _, txn := range txns {
		protoTxns = append(protoTxns, &pb.Transaction{
			Id:           txn.ID,
			UserId:       txn.UserID,
			SpiceGradeId: txn.SpiceGradeID,
			Type:         txn.Type,
			Quantity:     txn.Quantity,
			Price:        txn.Price,
			TradeDate:    txn.TradeDate.Format("2006-01-02"),
			CreatedAt:    txn.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return &pb.ListTransactionsResponse{
		Transactions: protoTxns,
	}, nil
}
