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
	DefaultRecentTxnTake = 5
)

const (
	InsightKindWinner        = "WINNER"
	InsightKindLoser         = "LOSER"
	InsightKindStreak        = "STREAK"
	InsightKindIdle          = "IDLE"
	InsightKindMilestone     = "MILESTONE"
	InsightKindConcentration = "CONCENTRATION"
)

const (
	InsightSeverityInfo    = "info"
	InsightSeveritySuccess = "success"
	InsightSeverityWarning = "warning"
)

const (
	PriceDirectionUp   = "UP"
	PriceDirectionDown = "DOWN"
	PriceDirectionFlat = "FLAT"
)

// MerchantDashboard is the top-level aggregate returned by GetMerchantDashboard.
// All slices are scoped to a single merchant account_id (from JWT, never from client input).
type MerchantDashboard struct {
	Summary            MerchantSummary
	Holdings           []MerchantHolding
	PortfolioMix       []PortfolioSlice
	PnLTrend           []PnLPoint
	ActivityTrend      []ActivityDay
	RecentTransactions []Transaction
	Insights           []MerchantInsight
	PriceMovers        []PriceMover
}

// MerchantSummary holds headline portfolio metrics for the dashboard hero cards.
type MerchantSummary struct {
	PortfolioValue     float64
	TotalCost          float64
	TotalRealizedPnL   float64
	TotalUnrealizedPnL float64
	NetPnL             float64
	OpenPositions      int
	TotalQuantityKg    float64
	TradesInPeriod     int
	BuyVolumeInPeriod  float64
	SellVolumeInPeriod float64
}

// MerchantHolding is an enriched open position with catalog names and computed market values.
type MerchantHolding struct {
	SpiceGradeID         string
	ProductName          string
	GradeName            string
	Quantity             float64
	AvgCost              float64
	TodayPrice           float64
	MarketValue          float64
	CostBasis            float64
	UnrealizedPnL        float64
	UnrealizedPnLPercent float64
	RealizedPnL          float64
	WeightPercent        float64
}

// EnrichedHoldingRow is the repository row before service-layer computed fields.
type EnrichedHoldingRow struct {
	SpiceGradeID string
	ProductName  string
	GradeName    string
	TotalQty     float64
	TotalCost    float64
	RealizedPnL  float64
	TodayPrice   float64
}

// PortfolioSlice represents one segment of the portfolio mix chart (by product/grade label).
type PortfolioSlice struct {
	Label    string
	Value    float64
	Quantity float64
}

// PnLPoint is one day on the cumulative realized P&L trend chart.
type PnLPoint struct {
	Date                  string // YYYY-MM-DD
	DailyRealizedPnL      float64
	CumulativeRealizedPnL float64
}

// ActivityDay is one day on the buy vs sell activity chart.
type ActivityDay struct {
	Date         string // YYYY-MM-DD
	BuyQuantity  float64
	SellQuantity float64
	BuyCount     int
	SellCount    int
}

// MerchantInsight is a rule-generated card shown on the dashboard.
type MerchantInsight struct {
	Kind         string
	Title        string
	Body         string
	SpiceGradeID string
	Severity     string
}

// PriceMover tracks day-over-day price change for a grade the merchant holds.
type PriceMover struct {
	SpiceGradeID  string
	ProductName   string
	GradeName     string
	TodayPrice    float64
	PreviousPrice float64
	ChangePercent float64
	Direction     string
}

// DailyRealizedPnLRow is a repository row for building the P&L trend series.
type DailyRealizedPnLRow struct {
	Date             time.Time
	DailyRealizedPnL float64
}

// DailyActivityRow is a repository row for building the activity trend series.
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
