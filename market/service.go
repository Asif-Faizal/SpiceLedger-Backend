package market

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Asif-Faizal/SpiceLedger-Backend/util"
)

type Service interface {
	Buy(ctx context.Context, userID string, spiceGradeID string, quantity float64, price float64, tradeDate time.Time) (*Transaction, error)
	Sell(ctx context.Context, userID string, spiceGradeID string, quantity float64, price float64, tradeDate time.Time) (*Transaction, error)
	GetGradePosition(ctx context.Context, userID string, spiceGradeID string) (*PositionView, error)
	ListGradeTransactions(ctx context.Context, userID string, spiceGradeID string, skip, take uint) ([]*Transaction, error)
}

type MarketService struct {
	repository Repository
	logger     util.Logger
}

func NewMarketService(repository Repository, logger util.Logger) Service {
	return &MarketService{
		repository: repository,
		logger:     logger,
	}
}

// Buy records a BUY transaction and creates a new buy_lot.
func (s *MarketService) Buy(ctx context.Context, userID string, spiceGradeID string, quantity float64, price float64, tradeDate time.Time) (*Transaction, error) {
	if userID == "" {
		return nil, errors.New("user_id is required")
	}
	if spiceGradeID == "" {
		return nil, errors.New("spice_grade_id is required")
	}
	if quantity <= 0 {
		return nil, errors.New("quantity must be greater than zero")
	}
	if price <= 0 {
		return nil, errors.New("price must be greater than zero")
	}
	if tradeDate.IsZero() {
		tradeDate = time.Now()
	}

	txCtx, tx, err := s.repository.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// 1. Record the immutable transaction.
	t := &Transaction{
		UserID:       userID,
		SpiceGradeID: spiceGradeID,
		Type:         "BUY",
		Quantity:     quantity,
		Price:        price,
		TradeDate:    tradeDate,
	}
	txnID, err := s.repository.InsertTransaction(txCtx, t)
	if err != nil {
		return nil, err
	}
	t.ID = txnID

	// 2. Create the inventory lot (original_qty = remaining_qty = full purchase qty).
	lot := &BuyLot{
		TransactionID: txnID,
		UserID:        userID,
		SpiceGradeID:  spiceGradeID,
		OriginalQty:   quantity,
		RemainingQty:  quantity,
		Price:         price,
		TradeDate:     tradeDate,
	}
	if _, err = s.repository.InsertBuyLot(txCtx, lot); err != nil {
		return nil, err
	}

	// 3. Upsert position — increase qty and cost.
	pos := &Position{
		UserID:       userID,
		SpiceGradeID: spiceGradeID,
		TotalQty:     quantity,
		TotalCost:    quantity * price,
		RealizedPnL:  0,
	}
	if err = s.repository.UpsertPosition(txCtx, pos); err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return t, nil
}

// Sell matches the requested quantity against open buy_lots in FIFO order.
// All lot deductions, sell_allocations, and position updates are atomic.
func (s *MarketService) Sell(ctx context.Context, userID string, spiceGradeID string, quantity float64, price float64, tradeDate time.Time) (*Transaction, error) {
	if userID == "" {
		return nil, errors.New("user_id is required")
	}
	if spiceGradeID == "" {
		return nil, errors.New("spice_grade_id is required")
	}
	if quantity <= 0 {
		return nil, errors.New("quantity must be greater than zero")
	}
	if price <= 0 {
		return nil, errors.New("price must be greater than zero")
	}
	if tradeDate.IsZero() {
		tradeDate = time.Now()
	}

	txCtx, tx, err := s.repository.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// 1. Lock open lots in FIFO order (FOR UPDATE prevents concurrent oversell).
	lots, err := s.repository.GetOpenBuyLots(txCtx, userID, spiceGradeID)
	if err != nil {
		return nil, err
	}

	// Service-layer inventory check.
	var totalAvailable float64
	for _, l := range lots {
		totalAvailable += l.RemainingQty
	}
	if totalAvailable < quantity {
		err = errors.New("insufficient inventory: sell quantity exceeds available buy lots")
		return nil, err
	}

	// 2. Record the immutable SELL transaction.
	t := &Transaction{
		UserID:       userID,
		SpiceGradeID: spiceGradeID,
		Type:         "SELL",
		Quantity:     quantity,
		Price:        price,
		TradeDate:    tradeDate,
	}
	txnID, err := s.repository.InsertTransaction(txCtx, t)
	if err != nil {
		return nil, err
	}
	t.ID = txnID

	// 3. Walk lots oldest→newest, consuming until the sell quantity is filled.
	remaining := quantity
	var totalRealizedPnL float64
	var totalCostConsumed float64

	for _, lot := range lots {
		if remaining <= 0 {
			break
		}

		consume := lot.RemainingQty
		if consume > remaining {
			consume = remaining
		}

		if err = s.repository.DeductBuyLotQty(txCtx, lot.ID, consume); err != nil {
			return nil, err
		}

		lotPnL := (price - lot.Price) * consume
		alloc := &SellAllocation{
			SellTransactionID: txnID,
			BuyLotID:          lot.ID,
			Quantity:          consume,
			BuyPrice:          lot.Price,
			SellPrice:         price,
			RealizedPnL:       lotPnL,
		}
		if err = s.repository.InsertSellAllocation(txCtx, alloc); err != nil {
			return nil, err
		}

		totalRealizedPnL += lotPnL
		totalCostConsumed += lot.Price * consume
		remaining -= consume
	}

	// 4. Update position: decrease qty + cost, accumulate realized P&L.
	pos := &Position{
		UserID:       userID,
		SpiceGradeID: spiceGradeID,
		TotalQty:     -(quantity),
		TotalCost:    -(totalCostConsumed),
		RealizedPnL:  totalRealizedPnL,
	}
	if err = s.repository.UpsertPosition(txCtx, pos); err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return t, nil
}

// GetPosition returns the aggregate position with live unrealized P&L from today's daily_price.
// If today's price is not yet published, UnrealizedPnL and TodayPrice are left as zero.
func (s *MarketService) GetGradePosition(ctx context.Context, userID string, spiceGradeID string) (*PositionView, error) {
	if userID == "" {
		return nil, errors.New("user_id is required")
	}
	if spiceGradeID == "" {
		return nil, errors.New("spice_grade_id is required")
	}

	pos, err := s.repository.GetGradePosition(ctx, userID, spiceGradeID)
	if err == sql.ErrNoRows {
		return nil, errors.New("no position found for this user and grade")
	}
	if err != nil {
		return nil, err
	}

	view := &PositionView{
		UserID:       pos.UserID,
		SpiceGradeID: pos.SpiceGradeID,
		TotalQty:     pos.TotalQty,
		TotalCost:    pos.TotalCost,
		RealizedPnL:  pos.RealizedPnL,
		UpdatedAt:    pos.UpdatedAt,
	}

	if pos.TotalQty > 0 {
		view.AvgCost = pos.TotalCost / pos.TotalQty
	}

	// Best-effort: fetch today's price for unrealized P&L.
	// If not published yet, we return the position without unrealized P&L.
	todayPrice, priceErr := s.repository.GetDailyPrice(ctx, spiceGradeID, time.Now())
	if priceErr == nil {
		view.TodayPrice = todayPrice
		view.UnrealizedPnL = (todayPrice - view.AvgCost) * pos.TotalQty
	}

	return view, nil
}

// ListTransactions returns paginated trade history
func (s *MarketService) ListGradeTransactions(ctx context.Context, userID string, spiceGradeID string, skip, take uint) ([]*Transaction, error) {
	if userID == "" {
		return nil, errors.New("user_id is required")
	}
	if spiceGradeID == "" {
		return nil, errors.New("spice_grade_id is required")
	}
	if take == 0 || take > 100 {
		take = 100
	}
	return s.repository.ListGradeTransactionsByUser(ctx, userID, spiceGradeID, skip, take)
}
