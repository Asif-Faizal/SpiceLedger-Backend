package market

import "time"

type Transaction struct {
	ID           string
	UserID       string
	SpiceGradeID string
	Type         string
	Quantity     float64
	Price        float64
	TradeDate    time.Time
	CreatedAt    time.Time
}

type BuyLot struct {
	ID            string
	TransactionID string
	UserID        string
	SpiceGradeID  string
	OriginalQty   float64
	RemainingQty  float64
	Price         float64
	TradeDate     time.Time
	CreatedAt     time.Time
}

type SellAllocation struct {
	ID                string
	SellTransactionID string
	BuyLotID          string
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

// PositionView extends Position with unrealised P&L computed at read time.
type PositionView struct {
	UserID        string
	SpiceGradeID  string
	TotalQty      float64
	TotalCost     float64
	AvgCost       float64 // total_cost / total_qty
	TodayPrice    float64 // from daily_price; 0 if not yet published
	RealizedPnL   float64
	UnrealizedPnL float64 // (today_price - avg_cost) × total_qty
	UpdatedAt     time.Time
}

// --- Merchant dashboard domain models ---

const (
	DefaultDashboardDays = 7
	MaxDashboardDays     = 90
)

type EnrichedHoldingRow struct {
	SpiceGradeID string
	ProductName  string
	GradeName    string
	TotalQty     float64
	TotalCost    float64
	RealizedPnL  float64
	TodayPrice   float64
}

type DailyRealizedPnLRow struct {
	Date             time.Time
	DailyRealizedPnL float64
}

type DailyActivityRow struct {
	Date     time.Time
	Type     string // BUY or SELL
	Quantity float64
	Count    int
}

// PeriodTradeStats aggregates buy/sell volume and trade count over a date window.
type PeriodTradeStats struct {
	TradesInPeriod     int
	BuyVolumeInPeriod  float64
	SellVolumeInPeriod float64
}

// PriceSnapshot holds today and previous daily_price for a held grade.
type PriceSnapshot struct {
	SpiceGradeID  string
	ProductName   string
	GradeName     string
	TodayPrice    float64
	PreviousPrice float64
}
