package service

import (
    "context"
    "time"

    "github.com/google/uuid"
    "github.com/saravanan/spice_backend/internal/domain"
)

type ProductService struct {
    productRepo domain.ProductRepository
}

func NewProductService(productRepo domain.ProductRepository) *ProductService {
    return &ProductService{productRepo: productRepo}
}

func (s *ProductService) CreateProduct(ctx context.Context, name, description string) error {
    p := &domain.Product{
        ID:          uuid.New(),
        Name:        name,
        Description: description,
        CreatedAt:   time.Now(),
    }
    return s.productRepo.Create(ctx, p)
}

func (s *ProductService) ListProducts(ctx context.Context) ([]domain.Product, error) {
    return s.productRepo.FindAll(ctx)
}
