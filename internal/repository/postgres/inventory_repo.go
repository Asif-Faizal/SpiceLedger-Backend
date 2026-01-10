package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/Asif-Faizal/SpiceLedger/internal/domain"
	"gorm.io/gorm"
)

type inventoryRepository struct {
	db *gorm.DB
}

func NewInventoryRepository(db *gorm.DB) domain.InventoryRepository {
	return &inventoryRepository{db: db}
}

func (r *inventoryRepository) CreateLot(ctx context.Context, lot *domain.PurchaseLot) error {
	lot.ID = uuid.New()
	return r.db.WithContext(ctx).Create(lot).Error
}

func (r *inventoryRepository) GetLots(ctx context.Context, userID uuid.UUID, filters map[string]interface{}) ([]domain.PurchaseLot, error) {
    var lots []domain.PurchaseLot
    query := r.db.WithContext(ctx).Where("user_id = ?", userID)

    if gid, ok := filters["grade_id"].(uuid.UUID); ok && gid != uuid.Nil {
        query = query.Where("grade_id = ?", gid)
    }
    if pid, ok := filters["product_id"].(uuid.UUID); ok && pid != uuid.Nil {
        query = query.Where("product_id = ?", pid)
    }

    err := query.Order("date asc").Preload("Grade").Preload("Product").Find(&lots).Error
    return lots, err
}

func (r *inventoryRepository) GetLotByID(ctx context.Context, id uuid.UUID) (*domain.PurchaseLot, error) {
	var lot domain.PurchaseLot
	if err := r.db.WithContext(ctx).First(&lot, id).Error; err != nil {
		return nil, err
	}
	return &lot, nil
}

func (r *inventoryRepository) CreateSale(ctx context.Context, sale *domain.SaleTransaction) error {
	sale.ID = uuid.New()
	return r.db.WithContext(ctx).Create(sale).Error
}

func (r *inventoryRepository) GetSales(ctx context.Context, userID uuid.UUID, filters map[string]interface{}) ([]domain.SaleTransaction, error) {
    var sales []domain.SaleTransaction
    query := r.db.WithContext(ctx).Where("user_id = ?", userID)

    if gid, ok := filters["grade_id"].(uuid.UUID); ok && gid != uuid.Nil {
        query = query.Where("grade_id = ?", gid)
    }
    if pid, ok := filters["product_id"].(uuid.UUID); ok && pid != uuid.Nil {
        query = query.Where("product_id = ?", pid)
    }

    err := query.Order("date asc").Preload("Grade").Preload("Product").Find(&sales).Error
    return sales, err
}

func (r *inventoryRepository) GetHistory(ctx context.Context, userID uuid.UUID, date time.Time) ([]domain.PurchaseLot, []domain.SaleTransaction, error) {
	var lots []domain.PurchaseLot
	var sales []domain.SaleTransaction

	// Get all lots <= date
    err := r.db.WithContext(ctx).
        Preload("Grade").
        Preload("Grade.Product").
        Preload("Product").
        Where("user_id = ? AND date <= ?", userID, date).
        Order("date asc, created_at asc").
        Find(&lots).Error
	if err != nil {
		return nil, nil, err
	}

	// Get all sales <= date
    err = r.db.WithContext(ctx).
        Preload("Grade").
        Preload("Grade.Product").
        Preload("Product").
        Where("user_id = ? AND date <= ?", userID, date).
        Order("date asc, created_at asc").
        Find(&sales).Error
	if err != nil {
		return nil, nil, err
	}

	return lots, sales, nil
}
