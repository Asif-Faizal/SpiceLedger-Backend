package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
	Count(ctx context.Context) (int64, error)
}

type GradeRepository interface {
	Create(ctx context.Context, grade *Grade) error
	FindAll(ctx context.Context) ([]Grade, error)
}

type InventoryRepository interface {
	CreateLot(ctx context.Context, lot *PurchaseLot) error
	GetLots(ctx context.Context, userID uuid.UUID, filters map[string]interface{}) ([]PurchaseLot, error)
	GetLotByID(ctx context.Context, id uuid.UUID) (*PurchaseLot, error)

	CreateSale(ctx context.Context, sale *SaleTransaction) error
	GetSales(ctx context.Context, userID uuid.UUID, filters map[string]interface{}) ([]SaleTransaction, error)

	// GetHistory returns all lots and sales up to a certain date (inclusive) for calculation.
	// Can filter by grade if specific grade needed, otherwise returns all.
	GetHistory(ctx context.Context, userID uuid.UUID, date time.Time) ([]PurchaseLot, []SaleTransaction, error)
}

type PriceRepository interface {
	SetPrice(ctx context.Context, date string, grade string, price float64) error
	GetPrice(ctx context.Context, date string, grade string) (float64, error)
	GetPricesForDate(ctx context.Context, date string) ([]DailyPrice, error)
}
