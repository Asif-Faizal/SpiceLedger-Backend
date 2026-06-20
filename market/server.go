package market

import (
	"context"
	"fmt"
	"net"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/Asif-Faizal/SpiceLedger-Backend/internal/platform"
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
	platform.RegisterHealth(grpcServer, "market")

	return platform.RunGRPC(lis, grpcServer, logger, "market")
}

func (server *GrpcServer) GetMarketMetrics(ctx context.Context, req *pb.GetMarketMetricsRequest) (*pb.GetMarketMetricsResponse, error) {
	totalTx, totalVol, tops, err := server.marketService.GetMarketMetrics(ctx)
	if err != nil {
		return nil, err
	}

	var pbTops []*pb.GetMarketMetricsResponse_TopProduct
	for _, top := range tops {
		pbTops = append(pbTops, &pb.GetMarketMetricsResponse_TopProduct{
			ProductName: top.ProductName,
			GradeName:   top.GradeName,
			Volume:      top.Volume,
		})
	}

	return &pb.GetMarketMetricsResponse{
		TotalTransactions: totalTx,
		TotalVolume:       totalVol,
		TopProducts:       pbTops,
	}, nil
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
	listAll := false
	if userID == "" {
		if isAdmin, ok := ctx.Value(util.IsAdminKey).(bool); ok && isAdmin {
			listAll = true
		} else if id, ok := ctx.Value(util.AccountIDKey).(string); ok && id != "" {
			userID = id
		} else {
			return nil, fmt.Errorf("user_id is required")
		}
	}

	var txns []*Transaction
	var err error
	if listAll {
		txns, err = server.marketService.ListAllTransactions(ctx, uint(req.Skip), uint(req.Take))
	} else {
		txns, err = server.marketService.ListTransactions(ctx, userID, uint(req.Skip), uint(req.Take))
	}

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

func (server *GrpcServer) GetHoldings(ctx context.Context, req *pb.GetHoldingsRequest) (*pb.GetHoldingsResponse, error) {
	userID := req.GetUserId()
	if userID == "" {
		if id, ok := ctx.Value(util.AccountIDKey).(string); ok && id != "" {
			userID = id
		} else {
			return nil, fmt.Errorf("user_id is required")
		}
	}

	rows, err := server.marketService.GetEnrichedHoldings(ctx, userID)
	if err != nil {
		return nil, err
	}

	holdings := make([]*pb.EnrichedHolding, len(rows))
	for i, row := range rows {
		holdings[i] = &pb.EnrichedHolding{
			SpiceGradeId: row.SpiceGradeID,
			ProductName:  row.ProductName,
			GradeName:    row.GradeName,
			Quantity:     row.TotalQty,
			TotalCost:    row.TotalCost,
			RealizedPnl:  row.RealizedPnL,
			TodayPrice:   row.TodayPrice,
		}
	}

	return &pb.GetHoldingsResponse{Holdings: holdings}, nil
}

func (server *GrpcServer) GetRealizedPnLHistory(ctx context.Context, req *pb.GetRealizedPnLHistoryRequest) (*pb.GetRealizedPnLHistoryResponse, error) {
	userID := req.GetUserId()
	if userID == "" {
		if id, ok := ctx.Value(util.AccountIDKey).(string); ok && id != "" {
			userID = id
		} else {
			return nil, fmt.Errorf("user_id is required")
		}
	}

	rows, err := server.marketService.GetDailyRealizedPnLByUser(ctx, userID, uint(req.GetDays()))
	if err != nil {
		return nil, err
	}

	out := make([]*pb.RealizedPnLRow, len(rows))
	for i, row := range rows {
		out[i] = &pb.RealizedPnLRow{
			Date:   row.Date.Format("2006-01-02"),
			Amount: row.DailyRealizedPnL,
		}
	}

	return &pb.GetRealizedPnLHistoryResponse{Rows: out}, nil
}

func (server *GrpcServer) GetTradeActivity(ctx context.Context, req *pb.GetTradeActivityRequest) (*pb.GetTradeActivityResponse, error) {
	userID := req.GetUserId()
	if userID == "" {
		if id, ok := ctx.Value(util.AccountIDKey).(string); ok && id != "" {
			userID = id
		} else {
			return nil, fmt.Errorf("user_id is required")
		}
	}

	rows, err := server.marketService.GetDailyActivityByUser(ctx, userID, uint(req.GetDays()))
	if err != nil {
		return nil, err
	}

	out := make([]*pb.TradeActivityRow, len(rows))
	for i, row := range rows {
		out[i] = &pb.TradeActivityRow{
			Date:     row.Date.Format("2006-01-02"),
			Type:     row.Type,
			Quantity: row.Quantity,
			Count:    uint32(row.Count),
		}
	}

	return &pb.GetTradeActivityResponse{Rows: out}, nil
}

func (server *GrpcServer) GetTradeStats(ctx context.Context, req *pb.GetTradeStatsRequest) (*pb.GetTradeStatsResponse, error) {
	userID := req.GetUserId()
	if userID == "" {
		if id, ok := ctx.Value(util.AccountIDKey).(string); ok && id != "" {
			userID = id
		} else {
			return nil, fmt.Errorf("user_id is required")
		}
	}

	stats, err := server.marketService.GetPeriodTradeStats(ctx, userID, uint(req.GetDays()))
	if err != nil {
		return nil, err
	}

	return &pb.GetTradeStatsResponse{
		TradesInPeriod:     uint32(stats.TradesInPeriod),
		BuyVolumeInPeriod:  stats.BuyVolumeInPeriod,
		SellVolumeInPeriod: stats.SellVolumeInPeriod,
	}, nil
}

func (server *GrpcServer) GetPriceSnapshots(ctx context.Context, req *pb.GetPriceSnapshotsRequest) (*pb.GetPriceSnapshotsResponse, error) {
	userID := req.GetUserId()
	if userID == "" {
		if id, ok := ctx.Value(util.AccountIDKey).(string); ok && id != "" {
			userID = id
		} else {
			return nil, fmt.Errorf("user_id is required")
		}
	}

	snapshots, err := server.marketService.GetPriceSnapshotsForHoldings(ctx, userID)
	if err != nil {
		return nil, err
	}

	out := make([]*pb.PriceSnapshot, len(snapshots))
	for i, snap := range snapshots {
		out[i] = &pb.PriceSnapshot{
			SpiceGradeId:  snap.SpiceGradeID,
			ProductName:   snap.ProductName,
			GradeName:     snap.GradeName,
			TodayPrice:    snap.TodayPrice,
			PreviousPrice: snap.PreviousPrice,
		}
	}

	return &pb.GetPriceSnapshotsResponse{Snapshots: out}, nil
}
