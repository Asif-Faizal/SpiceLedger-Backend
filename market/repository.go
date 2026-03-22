package market

import (
	"context"
	"database/sql"
	"time"

	"github.com/Asif-Faizal/SpiceLedger-Backend/util"
)

type Repository interface {
	Close()

	// Transactions
	InsertTransaction(ctx context.Context, tx *Transaction) (string, error)
	GetTransactionByID(ctx context.Context, id string) (*Transaction, error)
	ListGradeTransactionsByUser(ctx context.Context, userID string, spiceGradeID string, skip, take uint) ([]*Transaction, error)
	ListTransactionsByUser(ctx context.Context, userID string, skip, take uint) ([]*Transaction, error)

	// Buy Lots (inventory)
	InsertBuyLot(ctx context.Context, lot *BuyLot) (string, error)
	// GetOpenBuyLots returns lots with remaining_qty > 0 in FIFO order (oldest trade_date first).
	// Uses FOR UPDATE — must be called inside a DB transaction.
	GetOpenBuyLots(ctx context.Context, userID string, spiceGradeID string) ([]*BuyLot, error)
	DeductBuyLotQty(ctx context.Context, lotID string, deductQty float64) error

	// Sell Allocations (FIFO audit trail)
	InsertSellAllocation(ctx context.Context, alloc *SellAllocation) error

	// Positions (aggregate state)
	UpsertPosition(ctx context.Context, pos *Position) error
	GetGradePosition(ctx context.Context, userID string, spiceGradeID string) (*Position, error)
	GetPositionsByUser(ctx context.Context, userID string) ([]*Position, error)

	// Daily Price (read from control service's shared table)
	// Returns ErrNoPriceAvailable when no price is published for that date yet.
	GetDailyPrice(ctx context.Context, gradeID string, date time.Time) (float64, error)

	// BeginTx starts a DB transaction and returns a context carrying it.
	// The service layer calls this to wrap multi-step FIFO operations atomically.
	BeginTx(ctx context.Context) (context.Context, *sql.Tx, error)
}

type MysqlRepository struct {
	db     *sql.DB
	logger util.Logger
}

func NewMysqlRepository(url string, logger util.Logger) (Repository, error) {
	db, err := sql.Open("mysql", url)
	if err != nil {
		logger.Database().Fatal().Err(err).Msg("Database connection failed")
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		logger.Database().Fatal().Err(err).Msg("Database connection failed")
		return nil, err
	}
	logger.Database().Info().Msg("Database connection established")
	return &MysqlRepository{db: db, logger: logger}, nil
}

func (r *MysqlRepository) Close() {
	r.db.Close()
}

// execer is satisfied by both *sql.DB and *sql.Tx.
type execer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type txKey struct{}

func (r *MysqlRepository) dbFromContext(ctx context.Context) execer {
	if tx, ok := ctx.Value(txKey{}).(*sql.Tx); ok && tx != nil {
		return tx
	}
	return r.db
}

func (r *MysqlRepository) BeginTx(ctx context.Context) (context.Context, *sql.Tx, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return ctx, nil, err
	}
	return context.WithValue(ctx, txKey{}, tx), tx, nil
}

// InsertTransaction inserts an immutable BUY or SELL record and returns its new ID.
func (r *MysqlRepository) InsertTransaction(ctx context.Context, t *Transaction) (string, error) {
	start := time.Now()
	query := `INSERT INTO transactions (id, user_id, spice_grade_id, type, quantity, price, trade_date)
	          VALUES (?, ?, ?, ?, ?, ?, ?)`

	_, err := r.dbFromContext(ctx).ExecContext(ctx, query,
		t.ID, t.UserID, t.SpiceGradeID, t.Type, t.Quantity, t.Price,
		t.TradeDate.Format("2006-01-02"),
	)

	r.logger.Database().Debug().
		Str("query", query).
		Str("duration", time.Since(start).String()).
		Bool("success", err == nil).
		Msg("InsertTransaction")

	if err != nil {
		return "", err
	}
	return t.ID, nil
}

// GetTransactionByID fetches a single transaction by its primary key.
func (r *MysqlRepository) GetTransactionByID(ctx context.Context, id string) (*Transaction, error) {
	start := time.Now()
	query := `SELECT id, user_id, spice_grade_id, type, quantity, price, trade_date, created_at
	          FROM transactions WHERE id = ?`

	row := r.dbFromContext(ctx).QueryRowContext(ctx, query, id)
	t := &Transaction{}
	err := row.Scan(&t.ID, &t.UserID, &t.SpiceGradeID, &t.Type,
		&t.Quantity, &t.Price, &t.TradeDate, &t.CreatedAt)

	r.logger.Database().Debug().
		Str("query", query).
		Str("duration", time.Since(start).String()).
		Bool("success", err == nil).
		Msg("GetTransactionByID")

	if err != nil {
		return nil, err
	}
	return t, nil
}

// ListTransactionsByUser returns paginated transactions for a user + grade, newest first.
func (r *MysqlRepository) ListGradeTransactionsByUser(ctx context.Context, userID string, spiceGradeID string, skip, take uint) ([]*Transaction, error) {
	start := time.Now()
	query := `SELECT id, user_id, spice_grade_id, type, quantity, price, trade_date, created_at
	          FROM transactions
	          WHERE user_id = ? AND spice_grade_id = ?
	          ORDER BY trade_date DESC, id DESC
	          LIMIT ? OFFSET ?`

	rows, err := r.dbFromContext(ctx).QueryContext(ctx, query, userID, spiceGradeID, take, skip)

	r.logger.Database().Debug().
		Str("query", query).
		Str("duration", time.Since(start).String()).
		Bool("success", err == nil).
		Msg("ListTransactionsByUser")

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txns []*Transaction
	for rows.Next() {
		t := &Transaction{}
		if err := rows.Scan(&t.ID, &t.UserID, &t.SpiceGradeID, &t.Type,
			&t.Quantity, &t.Price, &t.TradeDate, &t.CreatedAt); err != nil {
			return nil, err
		}
		txns = append(txns, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return txns, nil
}

// ListTransactionsByUser returns paginated transactions for a user across all grades, newest first.
func (r *MysqlRepository) ListTransactionsByUser(ctx context.Context, userID string, skip, take uint) ([]*Transaction, error) {
	start := time.Now()
	query := `SELECT id, user_id, spice_grade_id, type, quantity, price, trade_date, created_at
	          FROM transactions
	          WHERE user_id = ?
	          ORDER BY trade_date DESC, id DESC
	          LIMIT ? OFFSET ?`

	rows, err := r.dbFromContext(ctx).QueryContext(ctx, query, userID, take, skip)

	r.logger.Database().Debug().
		Str("query", query).
		Str("duration", time.Since(start).String()).
		Bool("success", err == nil).
		Msg("ListTransactionsByUser")

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txns []*Transaction
	for rows.Next() {
		t := &Transaction{}
		if err := rows.Scan(&t.ID, &t.UserID, &t.SpiceGradeID, &t.Type,
			&t.Quantity, &t.Price, &t.TradeDate, &t.CreatedAt); err != nil {
			return nil, err
		}
		txns = append(txns, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return txns, nil
}

// InsertBuyLot creates a new inventory lot from a BUY transaction and returns its ID.
func (r *MysqlRepository) InsertBuyLot(ctx context.Context, lot *BuyLot) (string, error) {
	start := time.Now()
	query := `INSERT INTO buy_lots (id, transaction_id, user_id, spice_grade_id, original_qty, remaining_qty, price, trade_date)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.dbFromContext(ctx).ExecContext(ctx, query,
		lot.ID, lot.TransactionID, lot.UserID, lot.SpiceGradeID,
		lot.OriginalQty, lot.RemainingQty, lot.Price,
		lot.TradeDate.Format("2006-01-02"),
	)

	r.logger.Database().Debug().
		Str("query", query).
		Str("duration", time.Since(start).String()).
		Bool("success", err == nil).
		Msg("InsertBuyLot")

	if err != nil {
		return "", err
	}
	return lot.ID, nil
}

func (r *MysqlRepository) GetOpenBuyLots(ctx context.Context, userID string, spiceGradeID string) ([]*BuyLot, error) {
	start := time.Now()
	query := `SELECT id, transaction_id, user_id, spice_grade_id, original_qty, remaining_qty, price, trade_date, created_at
	          FROM buy_lots
	          WHERE user_id = ? AND spice_grade_id = ? AND remaining_qty > 0
	          ORDER BY trade_date ASC, id ASC
	          FOR UPDATE`

	rows, err := r.dbFromContext(ctx).QueryContext(ctx, query, userID, spiceGradeID)

	r.logger.Database().Debug().
		Str("query", query).
		Str("duration", time.Since(start).String()).
		Bool("success", err == nil).
		Msg("GetOpenBuyLots")

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lots []*BuyLot
	for rows.Next() {
		l := &BuyLot{}
		if err := rows.Scan(&l.ID, &l.TransactionID, &l.UserID, &l.SpiceGradeID,
			&l.OriginalQty, &l.RemainingQty, &l.Price, &l.TradeDate, &l.CreatedAt); err != nil {
			return nil, err
		}
		lots = append(lots, l)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return lots, nil
}

// DeductBuyLotQty subtracts deductQty from a lot's remaining_qty.
// The WHERE remaining_qty >= ? guard is the last-line defence against oversell.
func (r *MysqlRepository) DeductBuyLotQty(ctx context.Context, lotID string, deductQty float64) error {
	start := time.Now()
	query := `UPDATE buy_lots SET remaining_qty = remaining_qty - ? WHERE id = ? AND remaining_qty >= ?`

	res, err := r.dbFromContext(ctx).ExecContext(ctx, query, deductQty, lotID, deductQty)

	r.logger.Database().Debug().
		Str("query", query).
		Str("duration", time.Since(start).String()).
		Bool("success", err == nil).
		Msg("DeductBuyLotQty")

	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrInsufficientLotQty
	}
	return nil
}

// InsertSellAllocation records one FIFO pairing between a SELL transaction and a BuyLot.
func (r *MysqlRepository) InsertSellAllocation(ctx context.Context, alloc *SellAllocation) error {
	start := time.Now()
	query := `INSERT INTO sell_allocations (id, sell_transaction_id, buy_lot_id, quantity, buy_price, sell_price, realized_pnl)
	          VALUES (?, ?, ?, ?, ?, ?, ?)`

	_, err := r.dbFromContext(ctx).ExecContext(ctx, query,
		alloc.ID, alloc.SellTransactionID, alloc.BuyLotID,
		alloc.Quantity, alloc.BuyPrice, alloc.SellPrice, alloc.RealizedPnL,
	)

	r.logger.Database().Debug().
		Str("query", query).
		Str("duration", time.Since(start).String()).
		Bool("success", err == nil).
		Msg("InsertSellAllocation")

	return err
}

// UpsertPosition inserts or updates the aggregate position for a user + grade.
// total_qty, total_cost, and realized_pnl are updated relatively.
func (r *MysqlRepository) UpsertPosition(ctx context.Context, pos *Position) error {
	start := time.Now()
	var err error
	var query string

	if pos.TotalQty < 0 {
		// For SELL (reduction), use a pure UPDATE to avoid violating CHECK constraints on the INSERT attempt.
		query = `UPDATE positions 
		          SET total_qty = total_qty + ?, 
		              total_cost = total_cost + ?, 
		              realized_pnl = realized_pnl + ?
		          WHERE user_id = ? AND spice_grade_id = ?`
		_, err = r.dbFromContext(ctx).ExecContext(ctx, query,
			pos.TotalQty, pos.TotalCost, pos.RealizedPnL,
			pos.UserID, pos.SpiceGradeID,
		)
	} else {
		// For BUY (increase), use INSERT ... ON DUPLICATE KEY UPDATE.
		query = `INSERT INTO positions (user_id, spice_grade_id, total_qty, total_cost, realized_pnl)
		          VALUES (?, ?, ?, ?, ?)
		          ON DUPLICATE KEY UPDATE
		            total_qty    = total_qty + VALUES(total_qty),
		            total_cost   = total_cost + VALUES(total_cost),
		            realized_pnl = realized_pnl + VALUES(realized_pnl)`
		_, err = r.dbFromContext(ctx).ExecContext(ctx, query,
			pos.UserID, pos.SpiceGradeID, pos.TotalQty, pos.TotalCost, pos.RealizedPnL,
		)
	}

	r.logger.Database().Debug().
		Str("query", query).
		Str("duration", time.Since(start).String()).
		Bool("success", err == nil).
		Msg("UpsertPosition")

	return err
}

// GetPosition returns the current position for a user + grade.
// Returns sql.ErrNoRows if the user has never traded this grade.
func (r *MysqlRepository) GetGradePosition(ctx context.Context, userID string, spiceGradeID string) (*Position, error) {
	start := time.Now()
	query := `SELECT user_id, spice_grade_id, total_qty, total_cost, realized_pnl, updated_at
	          FROM positions WHERE user_id = ? AND spice_grade_id = ?`

	row := r.dbFromContext(ctx).QueryRowContext(ctx, query, userID, spiceGradeID)
	pos := &Position{}
	err := row.Scan(&pos.UserID, &pos.SpiceGradeID, &pos.TotalQty, &pos.TotalCost, &pos.RealizedPnL, &pos.UpdatedAt)

	r.logger.Database().Debug().
		Str("query", query).
		Str("duration", time.Since(start).String()).
		Bool("success", err == nil).
		Msg("GetPosition")

	if err != nil {
		return nil, err
	}
	return pos, nil
}

// GetPositionsByUser returns all positions for a user across all grades.
// Includes positions with zero quantity if they have realized P&L.
func (r *MysqlRepository) GetPositionsByUser(ctx context.Context, userID string) ([]*Position, error) {
	start := time.Now()
	query := `SELECT user_id, spice_grade_id, total_qty, total_cost, realized_pnl, updated_at
	          FROM positions WHERE user_id = ? AND (total_qty > 0 OR realized_pnl != 0)`

	rows, err := r.dbFromContext(ctx).QueryContext(ctx, query, userID)

	r.logger.Database().Debug().
		Str("query", query).
		Str("duration", time.Since(start).String()).
		Bool("success", err == nil).
		Msg("GetPositionsByUser")

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var positions []*Position
	for rows.Next() {
		pos := &Position{}
		if err := rows.Scan(&pos.UserID, &pos.SpiceGradeID, &pos.TotalQty, &pos.TotalCost, &pos.RealizedPnL, &pos.UpdatedAt); err != nil {
			return nil, err
		}
		positions = append(positions, pos)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return positions, nil
}

// GetDailyPrice returns the canonical market price for a grade on a given date.
// daily_price enforces UNIQUE(grade_id, date) so at most one row is returned.
// Returns ErrNoPriceAvailable if no price entry exists for that date yet.
func (r *MysqlRepository) GetDailyPrice(ctx context.Context, gradeID string, date time.Time) (float64, error) {
	start := time.Now()
	query := `SELECT price FROM daily_price WHERE grade_id = ? AND date = ? LIMIT 1`

	row := r.dbFromContext(ctx).QueryRowContext(ctx, query, gradeID, date.Format("2006-01-02"))
	var price float64
	err := row.Scan(&price)

	r.logger.Database().Debug().
		Str("query", query).
		Str("duration", time.Since(start).String()).
		Bool("success", err == nil).
		Msg("GetDailyPrice")

	if err == sql.ErrNoRows {
		return 0, ErrNoPriceAvailable
	}
	if err != nil {
		return 0, err
	}
	return price, nil
}

// --- Sentinel Errors ---

var ErrInsufficientLotQty = errInsufficientLotQty("insufficient buy lot quantity: possible concurrent oversell")

type errInsufficientLotQty string

func (e errInsufficientLotQty) Error() string { return string(e) }

var ErrNoPriceAvailable = errNoPriceAvailable("no daily price available for this grade on the given date")

type errNoPriceAvailable string

func (e errNoPriceAvailable) Error() string { return string(e) }
