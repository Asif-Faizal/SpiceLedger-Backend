package graphql

import (
	"context"

	"github.com/Asif-Faizal/SpiceLedger-Backend/control/pb"
)

// CreateProduct is the resolver for the createProduct field.
func (r *mutationResolver) CreateProduct(ctx context.Context, input CreateProductInput) (*ProductWithGradesAndPrice, error) {
	desc := ""
	if input.Description != nil {
		desc = *input.Description
	}
	status := ""
	if input.Status != nil {
		status = *input.Status
	}
	if input.ID == "" && status == "" {
		status = "active"
	}

	resp, err := r.server.controlClient.CreateOrUpdateProduct(ctx, &pb.CreateOrUpdateProductRequest{
		Id:          input.ID,
		Name:        input.Name,
		Category:    input.Category,
		Description: desc,
		Status:      status,
	})
	if err != nil {
		return nil, err
	}
	return &ProductWithGradesAndPrice{
		ID:          resp.Product.Id,
		Name:        resp.Product.Name,
		Category:    resp.Product.Category,
		Description: resp.Product.Description,
		Status:      resp.Product.Status,
	}, nil
}

// CreateGrade is the resolver for the createGrade field.
func (r *mutationResolver) CreateGrade(ctx context.Context, input CreateGradeInput) (*GradeWithPrice, error) {
	desc := ""
	if input.Description != nil {
		desc = *input.Description
	}
	status := ""
	if input.Status != nil {
		status = *input.Status
	}
	if input.ID == "" && status == "" {
		status = "active"
	}

	resp, err := r.server.controlClient.CreateOrUpdateGrade(ctx, &pb.CreateOrUpdateGradeRequest{
		Id:          input.ID,
		ProductId:   input.ProductID,
		Name:        input.Name,
		Description: desc,
		Status:      status,
	})
	if err != nil {
		return nil, err
	}
	return &GradeWithPrice{
		ID:          resp.Grade.Id,
		ProductID:   resp.Grade.ProductId,
		Name:        resp.Grade.Name,
		Description: resp.Grade.Description,
		Status:      resp.Grade.Status,
	}, nil
}

// CreateDailyPrice is the resolver for the createDailyPrice field.
func (r *mutationResolver) CreateDailyPrice(ctx context.Context, input CreateDailyPriceInput) (*DailyPrice, error) {
	resp, err := r.server.controlClient.CreateOrUpdateDailyPrice(ctx, &pb.CreateOrUpdateDailyPriceRequest{
		Id:        input.ID,
		ProductId: input.ProductID,
		GradeId:   input.GradeID,
		Price:     input.Price,
		Date:      input.Date,
		Time:      input.Time,
	})
	if err != nil {
		return nil, err
	}
	return &DailyPrice{
		ID:        resp.DailyPrice.Id,
		ProductID: resp.DailyPrice.ProductId,
		GradeID:   resp.DailyPrice.GradeId,
		Price:     resp.DailyPrice.Price,
		Date:      resp.DailyPrice.Date,
		Time:      resp.DailyPrice.Time,
	}, nil
}
