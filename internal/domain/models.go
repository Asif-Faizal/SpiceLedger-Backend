package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Email        string    `gorm:"uniqueIndex;not null" json:"email"`
	Name         string    `json:"name"`
	Role         string    `gorm:"type:varchar(20);default:'user'" json:"role"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Grade struct {
	Name        string    `gorm:"primaryKey;type:varchar(50)" json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type PurchaseLot struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	Date      time.Time `gorm:"type:date;not null;index" json:"date"`
	Quantity  float64   `gorm:"type:decimal(10,2);not null" json:"quantity_kg"`
	UnitCost  float64   `gorm:"type:decimal(10,2);not null" json:"unit_cost"`
	Grade     string    `gorm:"type:varchar(50);not null;index" json:"grade"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SaleTransaction struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	Date      time.Time `gorm:"type:date;not null;index" json:"date"`
	Quantity  float64   `gorm:"type:decimal(10,2);not null" json:"quantity_kg"`
	UnitPrice float64   `gorm:"type:decimal(10,2);not null" json:"unit_price"`
	Grade     string    `gorm:"type:varchar(50);not null;index" json:"grade"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DTOs for Logic & API

type DailyPrice struct {
	Date       string  `json:"date"`
	Grade      string  `json:"grade"`
	PricePerKg float64 `json:"price_per_kg"`
}

type InventorySnapshot struct {
	Grade          string  `json:"grade"`
	TotalQuantity  float64 `json:"total_quantity_kg"`
	AverageCost    float64 `json:"average_cost_per_kg"`
	TotalCostBasis float64 `json:"total_cost_basis"`
	MarketPrice    float64 `json:"market_price_per_kg,omitempty"`
	MarketValue    float64 `json:"market_value,omitempty"`
	UnrealizedPnL  float64 `json:"unrealized_pnl,omitempty"`
	RealizedPnL    float64 `json:"realized_pnl,omitempty"` // For history, maybe?
}

type OverallInventory struct {
	Snapshots     []InventorySnapshot `json:"grades"`
	TotalQuantity float64             `json:"total_quantity_kg"`
	TotalValue    float64             `json:"total_value"`
	TotalCost     float64             `json:"total_cost"`
	TotalPnL      float64             `json:"total_pnl"`
}
