package control

import (
	"context"
	"errors"
	"time"

	"github.com/Asif-Faizal/SpiceLedger-Backend/util"
	"github.com/segmentio/ksuid"
)

type Service interface {
	CreateOrUpdateAccount(ctx context.Context, account *Account) (*Account, error)
	GetAccountByID(ctx context.Context, id string) (*Account, error)
	ListAccounts(ctx context.Context, skip uint, take uint) ([]*Account, error)
	CheckEmailExists(ctx context.Context, email string) (bool, error)
	Login(ctx context.Context, email string, password string, deviceID string) (*AuthenticatedResponse, error)
	Logout(ctx context.Context, accessToken string, deviceID string) error
	RefreshToken(ctx context.Context, refreshToken string, deviceID string) (*AuthenticatedResponse, error)
}

type AccountService struct {
	repository         Repository
	jwtSecret          string
	accessTokenExpiry  time.Duration
	refreshTokenExpiry time.Duration
}

func NewAccountService(
	repository Repository,
	jwtSecret string,
	accessTokenExpiry time.Duration,
	refreshTokenExpiry time.Duration,
) *AccountService {
	return &AccountService{
		repository:         repository,
		jwtSecret:          jwtSecret,
		accessTokenExpiry:  accessTokenExpiry,
		refreshTokenExpiry: refreshTokenExpiry,
	}
}

func (service *AccountService) CheckEmailExists(ctx context.Context, email string) (bool, error) {
	exists, err := service.repository.CheckEmailExists(ctx, email)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (service *AccountService) CreateOrUpdateAccount(ctx context.Context, account *Account) (*Account, error) {
	if account.UserType == "" {
		return nil, errors.New("user_type is required")
	}
	if account.Email == "" {
		return nil, errors.New("email is required")
	}

	id := account.ID

	// Check if email already exists for a different user
	existingAccount, err := service.repository.GetAccountByEmail(ctx, account.Email)
	if err == nil && existingAccount != nil {
		if id == "" || id != existingAccount.ID {
			return nil, errors.New("email already in use")
		}
	}

	if id == "" {
		id = ksuid.New().String()
	}
	hashed := ""
	if account.Password != "" {
		hash, err := util.HashPassword(account.Password)
		if err != nil {
			return nil, err
		}
		hashed = hash
	}
	newAccount := &Account{
		ID:       id,
		Name:     account.Name,
		UserType: account.UserType,
		Email:    account.Email,
		Password: hashed,
	}
	if _, err := service.repository.CreateOrUpdateAccount(ctx, newAccount); err != nil {
		return nil, err
	}
	return newAccount, nil
}

func (service *AccountService) GetAccountByID(ctx context.Context, id string) (*Account, error) {
	account, err := service.repository.GetAccountById(ctx, id)
	if err != nil {
		return nil, err
	}
	return account, nil
}

func (service *AccountService) ListAccounts(ctx context.Context, skip uint, take uint) ([]*Account, error) {
	if take > 100 || (skip == 0 && take == 0) {
		take = 100
	}
	accounts, err := service.repository.ListAccounts(ctx, skip, take)
	if err != nil {
		return nil, err
	}
	return accounts, nil
}

func (service *AccountService) Login(ctx context.Context, email string, password string, deviceID string) (*AuthenticatedResponse, error) {
	account, err := service.repository.GetAccountByEmail(ctx, email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	if !util.CheckPasswordHash(password, account.Password) {
		return nil, errors.New("invalid email or password")
	}

	accessToken, err := util.GenerateToken(account.ID, account.UserType, account.Email, service.jwtSecret, service.accessTokenExpiry)
	if err != nil {
		return nil, err
	}

	refreshToken, err := util.GenerateToken(account.ID, account.UserType, account.Email, service.jwtSecret, service.refreshTokenExpiry)
	if err != nil {
		return nil, err
	}

	session := &Session{
		ID:           ksuid.New().String(),
		AccountID:    account.ID,
		DeviceID:     deviceID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(service.refreshTokenExpiry),
		CreatedAt:    time.Now(),
		IsRevoked:    false,
	}

	if err := service.repository.CreateOrUpdateSession(ctx, session); err != nil {
		return nil, err
	}

	return &AuthenticatedResponse{
		Account:      account,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (service *AccountService) Logout(ctx context.Context, accessToken string, deviceID string) error {
	// 1. Validate Access Token
	_, err := util.ValidateToken(accessToken, service.jwtSecret)
	if err != nil {
		return errors.New("invalid or expired access token")
	}

	// 2. Fetch session and check device
	session, err := service.repository.GetSessionByAccessToken(ctx, accessToken)
	if err != nil {
		return errors.New("session not found")
	}

	if session.DeviceID != deviceID {
		return errors.New("device mismatch")
	}

	return service.repository.RevokeSessionByAccessToken(ctx, accessToken)
}

func (service *AccountService) RefreshToken(ctx context.Context, refreshToken string, deviceID string) (*AuthenticatedResponse, error) {
	// 1. Validate Refresh Token
	_, err := util.ValidateToken(refreshToken, service.jwtSecret)
	if err != nil {
		return nil, errors.New("invalid or expired refresh token")
	}

	// 2. Fetch session and check device
	session, err := service.repository.GetSessionByRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, errors.New("invalid or expired refresh token")
	}

	if session.DeviceID != deviceID {
		return nil, errors.New("device mismatch")
	}

	if session.ExpiresAt.Before(time.Now()) {
		session.IsRevoked = true
		_ = service.repository.CreateOrUpdateSession(ctx, session)
		return nil, errors.New("refresh token expired")
	}

	account, err := service.repository.GetAccountById(ctx, session.AccountID)
	if err != nil {
		return nil, err
	}

	newAccessToken, err := util.GenerateToken(account.ID, account.UserType, account.Email, service.jwtSecret, service.accessTokenExpiry)
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := util.GenerateToken(account.ID, account.UserType, account.Email, service.jwtSecret, service.refreshTokenExpiry)
	if err != nil {
		return nil, err
	}

	session.AccessToken = newAccessToken
	session.RefreshToken = newRefreshToken
	session.ExpiresAt = time.Now().Add(service.refreshTokenExpiry)

	if err := service.repository.CreateOrUpdateSession(ctx, session); err != nil {
		return nil, err
	}

	return &AuthenticatedResponse{
		Account:      account,
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}
