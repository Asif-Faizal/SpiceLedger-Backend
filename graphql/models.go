package graphql

type ProductWithGradesAndPrice struct {
	ID          string            `json:"id" validate:"required,uuid4"`
	Name        string            `json:"name" validate:"required,min=3,max=255"`
	Category    string            `json:"category" validate:"required,oneof=spice others"`
	Description string            `json:"description" validate:"omitempty,min=3,max=255"`
	Status      string            `json:"status" validate:"required,oneof=active inactive"`
	Grades      []*GradeWithPrice `json:"grades,omitempty"`
}

type GradeWithPrice struct {
	ID          string  `json:"id" validate:"required,uuid4"`
	ProductID   string  `json:"product_id" validate:"required,uuid4"`
	Name        string  `json:"name" validate:"required,min=3,max=255"`
	Price       float64 `json:"price" validate:"required"`
	Description string  `json:"description" validate:"omitempty,min=3,max=255"`
	Status      string  `json:"status" validate:"required,oneof=active inactive"`
}

type Transaction struct {
	ID           string  `json:"id"`
	UserID       string  `json:"user_id"`
	SpiceGradeID string  `json:"spice_grade_id"`
	Type         string  `json:"type"`
	Quantity     float64 `json:"quantity"`
	Price        float64 `json:"price"`
	TradeDate    string  `json:"trade_date"`
	CreatedAt    string  `json:"created_at"`
}

type PositionView struct {
	UserID        string  `json:"user_id"`
	SpiceGradeID  string  `json:"spice_grade_id"`
	TotalQty      float64 `json:"total_qty"`
	TotalCost     float64 `json:"total_cost"`
	AvgCost       float64 `json:"avg_cost"`
	TodayPrice    float64 `json:"today_price"`
	RealizedPnL   float64 `json:"realized_pnl"`
	UnrealizedPnL float64 `json:"unrealized_pnl"`
	UpdatedAt     string  `json:"updated_at"`
}
