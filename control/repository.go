package control

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Asif-Faizal/SpiceLedger-Backend/util"
)

type Repository interface {
	Close()
	CheckEmailExists(ctx context.Context, email string) (bool, error)
	CreateOrUpdateAccount(ctx context.Context, account *Account) (*Account, error)
	GetAccountById(ctx context.Context, id string) (*Account, error)
	GetAccountByEmail(ctx context.Context, email string) (*Account, error)
	ListAccounts(ctx context.Context, skip uint, take uint) ([]*Account, error)

	// Session Management
	CreateOrUpdateSession(ctx context.Context, session *Session) error
	GetSession(ctx context.Context, id string) (*Session, error)
	GetSessionByRefreshToken(ctx context.Context, refreshToken string) (*Session, error)
	GetSessionByAccessToken(ctx context.Context, accessToken string) (*Session, error)
	RevokeSessionByAccessToken(ctx context.Context, accessToken string) error

	// Merchant Details
	CreateOrUpdateMerchantDetails(ctx context.Context, merchantDetails *MerchantDetails) (*MerchantDetails, error)
	GetMerchantDetails(ctx context.Context, accountID string) (*MerchantDetails, error)

	// Products
	CreateOrUpdateProduct(ctx context.Context, product *Product) (*Product, error)
	ListProducts(ctx context.Context, skip uint, take uint) ([]*Product, error)
	GetProductsWithGradesAndPrices(ctx context.Context, date time.Time) ([]*ProductWithGrades, error)

	// Grades
	CreateOrUpdateGrade(ctx context.Context, grade *Grade) (*Grade, error)
	ListGradesByProductId(ctx context.Context, productId string, skip uint, take uint) ([]*Grade, error)

	// Daily Price
	CreateOrUpdateDailyPrice(ctx context.Context, dailyPrice *DailyPrice) (*DailyPrice, error)
	GetTodaysByProductId(ctx context.Context, productId string, date time.Time) ([]*DailyPrice, error)
	ListDailyPricesByGradeId(ctx context.Context, gradeId string, date time.Time, duration int) ([]*DailyPrice, error)
	GetTodaysByGradeId(ctx context.Context, gradeId string, date time.Time) ([]*DailyPrice, error)
}

type MysqlRepository struct {
	db     *sql.DB
	logger util.Logger
}

func (repository *MysqlRepository) GetTodaysByProductId(ctx context.Context, productId string, date time.Time) ([]*DailyPrice, error) {
	start := time.Now()
	query := "SELECT id, product_id, grade_id, price, date, time FROM daily_price WHERE product_id = ? AND date = ? ORDER BY time DESC"

	rows, err := repository.db.QueryContext(ctx, query, productId, date.Format("2006-01-02"))

	repository.logger.Database().Debug().
		Str("query", query).
		Str("duration", time.Since(start).String()).
		Bool("success", err == nil).
		Msg("Query Rows")

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dailyPrices := []*DailyPrice{}
	for rows.Next() {
		dailyPrice := &DailyPrice{}
		var timeStr string
		if err := rows.Scan(&dailyPrice.ID, &dailyPrice.ProductID, &dailyPrice.GradeID, &dailyPrice.Price, &dailyPrice.Date, &timeStr); err != nil {
			return nil, err
		}
		dailyPrice.Time, _ = time.Parse("15:04:05", timeStr)
		dailyPrices = append(dailyPrices, dailyPrice)
	}
	return dailyPrices, nil
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

func (repository *MysqlRepository) Close() {
	repository.db.Close()
}

func (repository *MysqlRepository) CheckEmailExists(ctx context.Context, email string) (bool, error) {
	start := time.Now()
	query := "SELECT EXISTS(SELECT 1 FROM accounts WHERE email = ?)"

	row := repository.db.QueryRowContext(ctx, query, email)
	var exists bool
	err := row.Scan(&exists)

	repository.logger.Database().Debug().
		Str("query", query).
		Str("duration", time.Since(start).String()).
		Bool("success", err == nil).
		Msg("Query Row")

	if err != nil {
		return false, err
	}
	return exists, nil
}

func (repository *MysqlRepository) CreateOrUpdateAccount(ctx context.Context, account *Account) (*Account, error) {
	start := time.Now()
	query := "INSERT INTO accounts (id, name, user_type, email, password) VALUES (?, NULLIF(?,''), ?, ?, ?) ON DUPLICATE KEY UPDATE name = NULLIF(?,''), user_type = ?, email = ?, password = IF(VALUES(password) = '', password, VALUES(password))"

	_, err := repository.db.ExecContext(ctx, query,
		account.ID, account.Name, account.UserType, account.Email, account.Password,
		account.Name, account.UserType, account.Email,
	)

	repository.logger.Database().Info().
		Str("query", query).
		Str("duration", time.Since(start).String()).
		Str("result", fmt.Sprintf("success: %v", err == nil)).
		Msg("DB")

	if err != nil {
		return nil, err
	}
	return account, nil
}

func (repository *MysqlRepository) GetAccountById(ctx context.Context, id string) (*Account, error) {
	start := time.Now()
	query := "SELECT id, name, user_type, email FROM accounts WHERE id = ?"

	row := repository.db.QueryRowContext(ctx, query, id)
	account := &Account{}
	var name sql.NullString
	err := row.Scan(&account.ID, &name, &account.UserType, &account.Email)

	repository.logger.Database().Debug().
		Str("query", query).
		Str("duration", time.Since(start).String()).
		Bool("success", err == nil).
		Msg("Query Row")

	if err != nil {
		return nil, err
	}
	account.Name = name.String
	return account, nil
}

func (repository *MysqlRepository) GetAccountByEmail(ctx context.Context, email string) (*Account, error) {
	start := time.Now()
	query := "SELECT id, name, user_type, email, password FROM accounts WHERE email = ?"

	row := repository.db.QueryRowContext(ctx, query, email)
	account := &Account{}
	var name sql.NullString
	err := row.Scan(&account.ID, &name, &account.UserType, &account.Email, &account.Password)

	repository.logger.Database().Info().
		Str("query", query+" ("+email+")").
		Str("duration", time.Since(start).String()).
		Str("result", fmt.Sprintf("success: %v", err == nil)).
		Msg("DB")

	if err != nil {
		return nil, err
	}
	account.Name = name.String
	return account, nil
}

func (repository *MysqlRepository) ListAccounts(ctx context.Context, skip uint, take uint) ([]*Account, error) {
	start := time.Now()
	query := "SELECT id, name, user_type, email FROM accounts ORDER by id DESC LIMIT ? OFFSET ?"

	rows, err := repository.db.QueryContext(ctx, query, take, skip)

	repository.logger.Database().Info().
		Str("query", query).
		Str("duration", time.Since(start).String()).
		Str("result", fmt.Sprintf("success: %v", err == nil)).
		Msg("DB")

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	accounts := []*Account{}
	for rows.Next() {
		account := &Account{}
		var name sql.NullString
		if err := rows.Scan(&account.ID, &name, &account.UserType, &account.Email); err != nil {
			return nil, err
		}
		account.Name = name.String
		accounts = append(accounts, account)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return accounts, nil
}

func (repository *MysqlRepository) CreateOrUpdateSession(ctx context.Context, session *Session) error {
	start := time.Now()
	query := `
		INSERT INTO sessions (id, account_id, device_id, access_token, refresh_token, expires_at, created_at, is_revoked)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE 
			access_token = VALUES(access_token),
			refresh_token = VALUES(refresh_token),
			expires_at = VALUES(expires_at),
			is_revoked = VALUES(is_revoked)
	`

	_, err := repository.db.ExecContext(ctx, query,
		session.ID,
		session.AccountID,
		session.DeviceID,
		session.AccessToken,
		session.RefreshToken,
		session.ExpiresAt,
		session.CreatedAt,
		session.IsRevoked,
	)

	repository.logger.Database().Debug().
		Str("query", query).
		Str("duration", time.Since(start).String()).
		Bool("success", err == nil).
		Msg("Execute Query")

	return err
}

func (repository *MysqlRepository) GetSession(ctx context.Context, id string) (*Session, error) {
	start := time.Now()
	query := "SELECT id, account_id, device_id, access_token, refresh_token, expires_at, created_at, is_revoked FROM sessions WHERE id = ?"

	row := repository.db.QueryRowContext(ctx, query, id)
	session := &Session{}
	err := row.Scan(&session.ID, &session.AccountID, &session.DeviceID, &session.AccessToken, &session.RefreshToken, &session.ExpiresAt, &session.CreatedAt, &session.IsRevoked)

	repository.logger.Database().Debug().
		Str("query", query).
		Str("duration", time.Since(start).String()).
		Bool("success", err == nil).
		Msg("Query Row")

	if err != nil {
		return nil, err
	}
	return session, nil
}

func (repository *MysqlRepository) GetSessionByRefreshToken(ctx context.Context, refreshToken string) (*Session, error) {
	start := time.Now()
	query := "SELECT id, account_id, device_id, access_token, refresh_token, expires_at, created_at, is_revoked FROM sessions WHERE refresh_token = ? AND is_revoked = 0"

	row := repository.db.QueryRowContext(ctx, query, refreshToken)
	session := &Session{}
	err := row.Scan(&session.ID, &session.AccountID, &session.DeviceID, &session.AccessToken, &session.RefreshToken, &session.ExpiresAt, &session.CreatedAt, &session.IsRevoked)

	repository.logger.Database().Debug().
		Str("query", query).
		Str("duration", time.Since(start).String()).
		Bool("success", err == nil).
		Msg("Query Row")

	if err != nil {
		return nil, err
	}
	return session, nil
}

func (repository *MysqlRepository) GetSessionByAccessToken(ctx context.Context, accessToken string) (*Session, error) {
	start := time.Now()
	query := "SELECT id, account_id, device_id, access_token, refresh_token, expires_at, created_at, is_revoked FROM sessions WHERE access_token = ?"

	row := repository.db.QueryRowContext(ctx, query, accessToken)
	session := &Session{}
	err := row.Scan(&session.ID, &session.AccountID, &session.DeviceID, &session.AccessToken, &session.RefreshToken, &session.ExpiresAt, &session.CreatedAt, &session.IsRevoked)

	repository.logger.Database().Debug().
		Str("query", query).
		Str("duration", time.Since(start).String()).
		Bool("success", err == nil).
		Msg("Query Row")

	if err != nil {
		return nil, err
	}
	return session, nil
}

func (repository *MysqlRepository) RevokeSessionByAccessToken(ctx context.Context, accessToken string) error {
	start := time.Now()
	query := "UPDATE sessions SET is_revoked = true WHERE access_token = ?"

	_, err := repository.db.ExecContext(ctx, query, accessToken)

	repository.logger.Database().Debug().
		Str("query", query).
		Str("duration", time.Since(start).String()).
		Bool("success", err == nil).
		Msg("Execute Query")

	return err
}

func (repository *MysqlRepository) CreateOrUpdateMerchantDetails(ctx context.Context, merchantDetails *MerchantDetails) (*MerchantDetails, error) {
	start := time.Now()
	query := "INSERT INTO merchant_details (id, account_id, phone_number, address, city, state, pincode) VALUES (?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE phone_number = ?, address = ?, city = ?, state = ?, pincode = ?"

	_, err := repository.db.ExecContext(ctx, query,
		merchantDetails.ID,
		merchantDetails.AccountID,
		merchantDetails.Phone,
		merchantDetails.Address,
		merchantDetails.City,
		merchantDetails.State,
		merchantDetails.Pincode,
		// ON DUPLICATE KEY UPDATE
		merchantDetails.Phone,
		merchantDetails.Address,
		merchantDetails.City,
		merchantDetails.State,
		merchantDetails.Pincode,
	)

	repository.logger.Database().Debug().
		Str("query", query).
		Str("duration", time.Since(start).String()).
		Bool("success", err == nil).
		Msg("Execute Query")

	if err != nil {
		return nil, err
	}
	return merchantDetails, nil
}

func (repository *MysqlRepository) GetMerchantDetails(ctx context.Context, accountID string) (*MerchantDetails, error) {
	start := time.Now()
	query := "SELECT id, account_id, phone_number, address, city, state, pincode FROM merchant_details WHERE account_id = ?"

	row := repository.db.QueryRowContext(ctx, query, accountID)
	merchantDetails := &MerchantDetails{}
	err := row.Scan(&merchantDetails.ID, &merchantDetails.AccountID, &merchantDetails.Phone, &merchantDetails.Address, &merchantDetails.City, &merchantDetails.State, &merchantDetails.Pincode)

	repository.logger.Database().Debug().
		Str("query", query).
		Str("duration", time.Since(start).String()).
		Bool("success", err == nil).
		Msg("Query Row")

	if err != nil {
		return nil, err
	}
	return merchantDetails, nil
}

func (repository *MysqlRepository) CreateOrUpdateProduct(ctx context.Context, product *Product) (*Product, error) {
	start := time.Now()
	query := "INSERT INTO products (id, name, category, description, status) VALUES (?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE name = ?, category = ?, description = ?, status = ?"

	_, err := repository.db.ExecContext(ctx, query,
		product.ID,
		product.Name,
		product.Category,
		product.Description,
		product.Status,
		product.Name,
		product.Category,
		product.Description,
		product.Status,
	)

	repository.logger.Database().Debug().
		Str("query", query).
		Str("duration", time.Since(start).String()).
		Bool("success", err == nil).
		Msg("Execute Query")

	if err != nil {
		return nil, err
	}
	return product, nil
}

func (repository *MysqlRepository) ListProducts(ctx context.Context, skip uint, take uint) ([]*Product, error) {
	start := time.Now()
	query := "SELECT id, name, category, description, status FROM products ORDER BY id DESC LIMIT ? OFFSET ?"

	rows, err := repository.db.QueryContext(ctx, query, take, skip)

	repository.logger.Database().Debug().
		Str("query", query).
		Str("duration", time.Since(start).String()).
		Bool("success", err == nil).
		Msg("Query Rows")

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := []*Product{}
	for rows.Next() {
		product := &Product{}
		if err := rows.Scan(&product.ID, &product.Name, &product.Category, &product.Description, &product.Status); err != nil {
			return nil, err
		}
		products = append(products, product)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return products, nil
}

func (repository *MysqlRepository) CreateOrUpdateGrade(ctx context.Context, grade *Grade) (*Grade, error) {
	start := time.Now()
	query := "INSERT INTO grade (id, product_id, name, description, status) VALUES (?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE product_id = ?, name = ?, description = ?, status = ?"

	_, err := repository.db.ExecContext(ctx, query,
		grade.ID,
		grade.ProductID,
		grade.Name,
		grade.Description,
		grade.Status,
		grade.ProductID,
		grade.Name,
		grade.Description,
		grade.Status,
	)

	repository.logger.Database().Debug().
		Str("query", query).
		Str("duration", time.Since(start).String()).
		Bool("success", err == nil).
		Msg("Execute Query")

	if err != nil {
		return nil, err
	}
	return grade, nil
}

func (repository *MysqlRepository) ListGradesByProductId(ctx context.Context, productId string, skip uint, take uint) ([]*Grade, error) {
	start := time.Now()
	query := "SELECT id, product_id, name, description, status FROM grade WHERE product_id = ? ORDER BY id DESC LIMIT ? OFFSET ?"

	rows, err := repository.db.QueryContext(ctx, query, productId, take, skip)

	repository.logger.Database().Debug().
		Str("query", query).
		Str("duration", time.Since(start).String()).
		Bool("success", err == nil).
		Msg("Query Rows")

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	grades := []*Grade{}
	for rows.Next() {
		grade := &Grade{}
		if err := rows.Scan(&grade.ID, &grade.ProductID, &grade.Name, &grade.Description, &grade.Status); err != nil {
			return nil, err
		}
		grades = append(grades, grade)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return grades, nil
}

func (repository *MysqlRepository) CreateOrUpdateDailyPrice(ctx context.Context, dailyPrice *DailyPrice) (*DailyPrice, error) {
	start := time.Now()
	query := "INSERT INTO daily_price (id, product_id, grade_id, price, date, time) VALUES (?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE price = ?, time = ?"

	_, err := repository.db.ExecContext(ctx, query,
		dailyPrice.ID,
		dailyPrice.ProductID,
		dailyPrice.GradeID,
		dailyPrice.Price,
		dailyPrice.Date.Format("2006-01-02"),
		dailyPrice.Time.Format("15:04:05"),
		dailyPrice.Price,
		dailyPrice.Time.Format("15:04:05"),
	)

	repository.logger.Database().Debug().
		Str("query", query).
		Str("duration", time.Since(start).String()).
		Bool("success", err == nil).
		Msg("Execute Query")

	if err != nil {
		return nil, err
	}
	return dailyPrice, nil
}

func (repository *MysqlRepository) ListDailyPricesByGradeId(ctx context.Context, gradeId string, date time.Time, duration int) ([]*DailyPrice, error) {
	start := time.Now()
	startDate := date.AddDate(0, 0, -duration)
	query := "SELECT id, product_id, grade_id, price, date, time FROM daily_price WHERE grade_id = ? AND date BETWEEN ? AND ? ORDER BY date DESC, time DESC"

	rows, err := repository.db.QueryContext(ctx, query, gradeId, startDate.Format("2006-01-02"), date.Format("2006-01-02"))

	repository.logger.Database().Debug().
		Str("query", query).
		Str("duration", time.Since(start).String()).
		Bool("success", err == nil).
		Msg("Query Rows")

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dailyPrices := []*DailyPrice{}
	for rows.Next() {
		dailyPrice := &DailyPrice{}
		var timeStr string
		if err := rows.Scan(&dailyPrice.ID, &dailyPrice.ProductID, &dailyPrice.GradeID, &dailyPrice.Price, &dailyPrice.Date, &timeStr); err != nil {
			return nil, err
		}
		dailyPrice.Time, _ = time.Parse("15:04:05", timeStr)
		dailyPrices = append(dailyPrices, dailyPrice)
	}
	return dailyPrices, nil
}

func (repository *MysqlRepository) GetTodaysByGradeId(ctx context.Context, gradeId string, date time.Time) ([]*DailyPrice, error) {
	start := time.Now()
	query := "SELECT id, product_id, grade_id, price, date, time FROM daily_price WHERE grade_id = ? AND date = ? ORDER BY time DESC"

	rows, err := repository.db.QueryContext(ctx, query, gradeId, date.Format("2006-01-02"))

	repository.logger.Database().Debug().
		Str("query", query).
		Str("duration", time.Since(start).String()).
		Bool("success", err == nil).
		Msg("Query Rows")

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dailyPrices := []*DailyPrice{}
	for rows.Next() {
		dailyPrice := &DailyPrice{}
		var timeStr string
		if err := rows.Scan(&dailyPrice.ID, &dailyPrice.ProductID, &dailyPrice.GradeID, &dailyPrice.Price, &dailyPrice.Date, &timeStr); err != nil {
			return nil, err
		}
		dailyPrice.Time, _ = time.Parse("15:04:05", timeStr)
		dailyPrices = append(dailyPrices, dailyPrice)
	}
	return dailyPrices, nil
}

func (repository *MysqlRepository) GetProductsWithGradesAndPrices(ctx context.Context, date time.Time) ([]*ProductWithGrades, error) {
	start := time.Now()
	query := `
		SELECT 
			p.id as product_id,
			p.name as product_name,
			p.category as product_category,
			p.description as product_description,
			p.status as product_status,
			g.id as grade_id,
			g.name as grade_name,
			g.description as grade_description,
			g.status as grade_status,
			dp.price
		FROM products p
		LEFT JOIN grade g ON g.product_id = p.id
		LEFT JOIN daily_price dp 
			ON dp.product_id = p.id 
			AND dp.grade_id = g.id
			AND dp.date = ?
		WHERE p.status = 'active'
		AND g.status = 'active'
		ORDER BY p.id, g.id
	`

	rows, err := repository.db.QueryContext(ctx, query, date.Format("2006-01-02"))

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	productMap := make(map[string]*ProductWithGrades)
	var products []*ProductWithGrades

	for rows.Next() {
		var pID, pName, pCategory, pDescription, pStatus string
		var gID, gName, gDescription, gStatus sql.NullString
		var dpPrice sql.NullFloat64

		err := rows.Scan(
			&pID, &pName, &pCategory, &pDescription, &pStatus,
			&gID, &gName, &gDescription, &gStatus,
			&dpPrice,
		)
		if err != nil {
			return nil, err
		}

		product, ok := productMap[pID]
		if !ok {
			product = &ProductWithGrades{
				ID:          pID,
				Name:        pName,
				Category:    pCategory,
				Description: pDescription,
				Status:      pStatus,
				Grades:      []*GradeWithPrice{},
			}
			productMap[pID] = product
			products = append(products, product)
		}

		if gID.Valid {
			product.Grades = append(product.Grades, &GradeWithPrice{
				ID:          gID.String,
				ProductID:   pID,
				Name:        gName.String,
				Description: gDescription.String,
				Status:      gStatus.String,
				Price:       dpPrice.Float64,
			})
		}
	}

	repository.logger.Database().Debug().
		Str("query", query).
		Str("duration", time.Since(start).String()).
		Msg("GetProductsWithGradesAndPrices completed")

	return products, nil
}
