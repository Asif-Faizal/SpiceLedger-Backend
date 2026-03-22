package graphql

import (
	"context"
	"time"

	"github.com/Asif-Faizal/SpiceLedger-Backend/control/pb"
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
	panic("not implemented")
}

// GetPositions is the resolver for the getPositions field.
func (r *queryResolver) GetPositions(ctx context.Context) ([]*PositionView, error) {
	panic("not implemented")
}

// ListGradeTransactions is the resolver for the listGradeTransactions field.
func (r *queryResolver) ListGradeTransactions(ctx context.Context, spiceGradeID string, skip *int, take *int) ([]*Transaction, error) {
	panic("not implemented")
}

// ListTransactions is the resolver for the listTransactions field.
func (r *queryResolver) ListTransactions(ctx context.Context, skip *int, take *int) ([]*Transaction, error) {
	panic("not implemented")
}
