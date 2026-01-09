package postgres

import (
    "context"
    "time"

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

func (r *productRepository) Count(ctx context.Context) (int64, error) {
    var count int64
    if err := r.db.WithContext(ctx).Model(&domain.Product{}).Count(&count).Error; err != nil {
        return 0, err
    }
    return count, nil
}

func (r *productRepository) CountCreatedBetween(ctx context.Context, start time.Time, end time.Time) (int64, error) {
    var count int64
    err := r.db.WithContext(ctx).
        Model(&domain.Product{}).
        Where("created_at >= ? AND created_at < ?", start, end).
        Count(&count).Error
    if err != nil {
        return 0, err
    }
    return count, nil
}
