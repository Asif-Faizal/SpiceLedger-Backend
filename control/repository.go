package control

import (
	"context"
	"database/sql"
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
	query := "INSERT INTO accounts (id, name, user_type, email, password) VALUES (?, NULLIF(?,''), ?, ?, ?) ON DUPLICATE KEY UPDATE name = NULLIF(?,''), user_type = ?, email = ?, password = ?"

	_, err := repository.db.ExecContext(ctx, query,
		account.ID, account.Name, account.UserType, account.Email, account.Password,
		account.Name, account.UserType, account.Email, account.Password,
	)

	repository.logger.Database().Debug().
		Str("query", query).
		Str("duration", time.Since(start).String()).
		Bool("success", err == nil).
		Msg("Execute Query")

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

func (repository *MysqlRepository) ListAccounts(ctx context.Context, skip uint, take uint) ([]*Account, error) {
	start := time.Now()
	query := "SELECT id, name, user_type, email FROM accounts ORDER by id DESC LIMIT ? OFFSET ?"

	rows, err := repository.db.QueryContext(ctx, query, take, skip)

	repository.logger.Database().Debug().
		Str("query", query).
		Str("duration", time.Since(start).String()).
		Bool("success", err == nil).
		Msg("Query Context")

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
