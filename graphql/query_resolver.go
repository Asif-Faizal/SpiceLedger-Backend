package graphql

import (
	"context"
	"time"

	"github.com/Asif-Faizal/SpiceLedger-Backend/control/pb"
	marketpb "github.com/Asif-Faizal/SpiceLedger-Backend/market/pb"
)

// Products is the resolver for the products field.
func (r *queryResolver) Products(ctx context.Context, date *string, search *string) ([]*ProductWithGradesAndPrice, error) {
	dateStr := ""
	if date != nil {
		dateStr = *date
	} else {
		dateStr = time.Now().Format("2006-01-02")
	}

	searchStr := ""
	if search != nil {
		searchStr = *search
	}

	resp, err := r.server.controlClient.GetProductsWithGradesAndPrices(ctx, &pb.GetProductsWithGradesAndPricesRequest{
		Date:   dateStr,
		Search: searchStr,
	})
	if err != nil {
		return nil, err
	}

	products := make([]*ProductWithGradesAndPrice, len(resp.Products))
	for i, p := range resp.Products {
		grades := make([]*GradeWithPrice, len(p.Grades))
		for j, g := range p.Grades {
			grades[j] = &GradeWithPrice{
				ID:          g.Id,
				ProductID:   g.ProductId,
				Name:        g.Name,
				Description: g.Description,
				Status:      g.Status,
				Price:       g.Price,
			}
		}

		products[i] = &ProductWithGradesAndPrice{
			ID:          p.Id,
			Name:        p.Name,
			Category:    p.Category,
			Description: p.Description,
			Status:      p.Status,
			Grades:      grades,
		}
	}
	return products, nil
}

// GetGradePosition is the resolver for the getGradePosition field.
func (r *queryResolver) GetGradePosition(ctx context.Context, spiceGradeID string) (*PositionView, error) {
	resp, err := r.server.marketClient.GetGradePosition(ctx, &marketpb.GetGradePositionRequest{
		SpiceGradeId: spiceGradeID,
	})
	if err != nil {
		return nil, err
	}
	return &PositionView{
		UserID:        resp.Position.UserId,
		SpiceGradeID:  resp.Position.SpiceGradeId,
		TotalQty:      resp.Position.TotalQty,
		TotalCost:     resp.Position.TotalCost,
		AvgCost:       resp.Position.AvgCost,
		TodayPrice:    resp.Position.TodayPrice,
		RealizedPnL:   resp.Position.RealizedPnl,
		UnrealizedPnL: resp.Position.UnrealizedPnl,
		UpdatedAt:     resp.Position.UpdatedAt,
	}, nil
}

// GetPositions is the resolver for the getPositions field.
func (r *queryResolver) GetPositions(ctx context.Context) ([]*PositionView, error) {
	resp, err := r.server.marketClient.GetPositions(ctx, &marketpb.GetPositionsRequest{})
	if err != nil {
		return nil, err
	}
	positions := make([]*PositionView, len(resp.Positions))
	for i, p := range resp.Positions {
		positions[i] = &PositionView{
			UserID:        p.UserId,
			SpiceGradeID:  p.SpiceGradeId,
			TotalQty:      p.TotalQty,
			TotalCost:     p.TotalCost,
			AvgCost:       p.AvgCost,
			TodayPrice:    p.TodayPrice,
			RealizedPnL:   p.RealizedPnl,
			UnrealizedPnL: p.UnrealizedPnl,
			UpdatedAt:     p.UpdatedAt,
		}
	}
	return positions, nil
}

// ListGradeTransactions is the resolver for the listGradeTransactions field.
func (r *queryResolver) ListGradeTransactions(ctx context.Context, spiceGradeID string, skip *int, take *int) ([]*Transaction, error) {
	var skip32, take32 uint32
	if skip != nil {
		skip32 = uint32(*skip)
	}
	if take != nil {
		take32 = uint32(*take)
	}
	resp, err := r.server.marketClient.ListGradeTransactions(ctx, &marketpb.ListGradeTransactionsRequest{
		SpiceGradeId: spiceGradeID,
		Skip:         skip32,
		Take:         take32,
	})
	if err != nil {
		return nil, err
	}
	transactions := make([]*Transaction, len(resp.Transactions))
	for i, t := range resp.Transactions {
		transactions[i] = &Transaction{
			ID:           t.Id,
			UserID:       t.UserId,
			SpiceGradeID: t.SpiceGradeId,
			Type:         t.Type,
			Quantity:     t.Quantity,
			Price:        t.Price,
			TradeDate:    t.TradeDate,
			CreatedAt:    t.CreatedAt,
		}
	}
	return transactions, nil
}

// ListTransactions is the resolver for the listTransactions field.
func (r *queryResolver) ListTransactions(ctx context.Context, skip *int, take *int) ([]*Transaction, error) {
	var skip32, take32 uint32
	if skip != nil {
		skip32 = uint32(*skip)
	}
	if take != nil {
		take32 = uint32(*take)
	}
	resp, err := r.server.marketClient.ListTransactions(ctx, &marketpb.ListTransactionsRequest{
		Skip: skip32,
		Take: take32,
	})
	if err != nil {
		return nil, err
	}
	transactions := make([]*Transaction, len(resp.Transactions))
	for i, t := range resp.Transactions {
		transactions[i] = &Transaction{
			ID:           t.Id,
			UserID:       t.UserId,
			SpiceGradeID: t.SpiceGradeId,
			Type:         t.Type,
			Quantity:     t.Quantity,
			Price:        t.Price,
			TradeDate:    t.TradeDate,
			CreatedAt:    t.CreatedAt,
		}
	}
	return transactions, nil
}
