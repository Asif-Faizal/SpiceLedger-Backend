package market

import "time"

type Transaction struct {
	ID           int64
	UserID       string
	SpiceGradeID string
	Type         string
	Quantity     float64
	Price        float64
	TradeDate    time.Time
	CreatedAt    time.Time
}

type BuyLot struct {
	ID            int64
	TransactionID int64
	UserID        string
	SpiceGradeID  string
	OriginalQty   float64
	RemainingQty  float64
	Price         float64
	TradeDate     time.Time
	CreatedAt     time.Time
}

type SellAllocation struct {
	ID                int64
	SellTransactionID int64
	BuyLotID          int64
	Quantity          float64
	BuyPrice          float64
	SellPrice         float64
	RealizedPnL       float64
	CreatedAt         time.Time
}

type Position struct {
	UserID       string
	SpiceGradeID string
	TotalQty     float64
	TotalCost    float64
	RealizedPnL  float64
	UpdatedAt    time.Time
}
