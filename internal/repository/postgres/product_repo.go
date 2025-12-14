package postgres

import (
    "context"

    "github.com/google/uuid"
    "github.com/saravanan/spice_backend/internal/domain"
    "gorm.io/gorm"
)

type productRepository struct {
    db *gorm.DB
}

func NewProductRepository(db *gorm.DB) domain.ProductRepository {
    return &productRepository{db: db}
}

func (r *productRepository) Create(ctx context.Context, product *domain.Product) error {
    if product.ID == uuid.Nil {
        product.ID = uuid.New()
    }
    return r.db.WithContext(ctx).Create(product).Error
}

func (r *productRepository) FindAll(ctx context.Context) ([]domain.Product, error) {
    var products []domain.Product
    if err := r.db.WithContext(ctx).Find(&products).Error; err != nil {
        return nil, err
    }
    return products, nil
}

func (r *productRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
    var product domain.Product
    if err := r.db.WithContext(ctx).First(&product, id).Error; err != nil {
        return nil, err
    }
    return &product, nil
}
