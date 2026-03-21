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
	CreateOrUpdateMerchantDetails(ctx context.Context, merchantDetails *MerchantDetails) (*MerchantDetails, error)
	GetMerchantDetails(ctx context.Context, accountID string) (*MerchantDetails, error)

	// Products
	CreateOrUpdateProduct(ctx context.Context, product *Product) (*Product, error)
	ListProducts(ctx context.Context, skip uint, take uint) ([]*Product, error)

	// Grades
	CreateOrUpdateGrade(ctx context.Context, grade *Grade) (*Grade, error)
	ListGradesByProductId(ctx context.Context, productId string, skip uint, take uint) ([]*Grade, error)

	// Daily Price
	CreateOrUpdateDailyPrice(ctx context.Context, dailyPrice *DailyPrice) (*DailyPrice, error)
	ListDailyPricesByGradeId(ctx context.Context, gradeId string, today time.Time, duration int) ([]*DailyPrice, error)
	GetTodaysByGradeId(ctx context.Context, gradeId string, date time.Time) ([]*DailyPrice, error)
	GetTodaysByProductId(ctx context.Context, productId string, date time.Time) ([]*DailyPrice, error)
	GetProductsWithGradesAndPrices(ctx context.Context, date time.Time) ([]*ProductWithGrades, error)
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
		if account.Password == "" {
			return nil, errors.New("password is required for new accounts")
		}
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

func (service *AccountService) CreateOrUpdateMerchantDetails(ctx context.Context, merchantDetails *MerchantDetails) (*MerchantDetails, error) {
	if merchantDetails.AccountID == "" {
		return nil, errors.New("account_id is required")
	}
	if merchantDetails.Phone == "" {
		return nil, errors.New("phone_number is required")
	}
	if merchantDetails.Address == "" {
		return nil, errors.New("address is required")
	}
	if merchantDetails.City == "" {
		return nil, errors.New("city is required")
	}
	if merchantDetails.State == "" {
		return nil, errors.New("state is required")
	}
	if merchantDetails.Pincode == "" {
		return nil, errors.New("pincode is required")
	}
	id := merchantDetails.ID

	// Check if merchant details already exists for this account
	existingMerchantDetails, err := service.repository.GetMerchantDetails(ctx, merchantDetails.AccountID)
	if err == nil && existingMerchantDetails != nil {
		// Use the existing ID to ensure we update the correct record
		id = existingMerchantDetails.ID
	}

	if id == "" {
		id = ksuid.New().String()
	}
	newMerchantDetails := &MerchantDetails{
		ID:        id,
		AccountID: merchantDetails.AccountID,
		Phone:     merchantDetails.Phone,
		Address:   merchantDetails.Address,
		City:      merchantDetails.City,
		State:     merchantDetails.State,
		Pincode:   merchantDetails.Pincode,
	}
	if _, err := service.repository.CreateOrUpdateMerchantDetails(ctx, newMerchantDetails); err != nil {
		return nil, err
	}
	return newMerchantDetails, nil
}

func (service *AccountService) GetMerchantDetails(ctx context.Context, accountID string) (*MerchantDetails, error) {
	merchantDetails, err := service.repository.GetMerchantDetails(ctx, accountID)
	if err != nil {
		return nil, err
	}
	return merchantDetails, nil
}

// Products
func (service *AccountService) CreateOrUpdateProduct(ctx context.Context, product *Product) (*Product, error) {
	id := product.ID
	if id == "" {
		id = ksuid.New().String()
	}
	newProduct := &Product{
		ID:          id,
		Name:        product.Name,
		Description: product.Description,
		Category:    product.Category,
		Status:      product.Status,
	}
	if _, err := service.repository.CreateOrUpdateProduct(ctx, newProduct); err != nil {
		return nil, err
	}
	return newProduct, nil
}

func (service *AccountService) ListProducts(ctx context.Context, skip uint, take uint) ([]*Product, error) {
	if take > 100 || (skip == 0 && take == 0) {
		take = 100
	}
	products, err := service.repository.ListProducts(ctx, skip, take)
	if err != nil {
		return nil, err
	}
	return products, nil
}

// Grades
func (service *AccountService) CreateOrUpdateGrade(ctx context.Context, grade *Grade) (*Grade, error) {
	id := grade.ID
	if id == "" {
		id = ksuid.New().String()
	}
	newGrade := &Grade{
		ID:          id,
		ProductID:   grade.ProductID,
		Name:        grade.Name,
		Description: grade.Description,
		Status:      grade.Status,
	}
	if _, err := service.repository.CreateOrUpdateGrade(ctx, newGrade); err != nil {
		return nil, err
	}
	return newGrade, nil
}

func (service *AccountService) ListGradesByProductId(ctx context.Context, productId string, skip uint, take uint) ([]*Grade, error) {
	if take > 100 || (skip == 0 && take == 0) {
		take = 100
	}
	grades, err := service.repository.ListGradesByProductId(ctx, productId, skip, take)
	if err != nil {
		return nil, err
	}
	return grades, nil
}

// Daily Price
func (service *AccountService) CreateOrUpdateDailyPrice(ctx context.Context, dailyPrice *DailyPrice) (*DailyPrice, error) {
	id := dailyPrice.ID
	if id == "" {
		id = ksuid.New().String()
	}
	newDailyPrice := &DailyPrice{
		ID:        id,
		ProductID: dailyPrice.ProductID,
		GradeID:   dailyPrice.GradeID,
		Price:     dailyPrice.Price,
		Date:      dailyPrice.Date,
		Time:      dailyPrice.Time,
	}
	if _, err := service.repository.CreateOrUpdateDailyPrice(ctx, newDailyPrice); err != nil {
		return nil, err
	}
	return newDailyPrice, nil
}

func (service *AccountService) ListDailyPricesByGradeId(ctx context.Context, gradeId string, today time.Time, duration int) ([]*DailyPrice, error) {
	if duration > 100 || (today.Equal(time.Time{}) && duration == 0) {
		duration = 100
	}
	dailyPrices, err := service.repository.ListDailyPricesByGradeId(ctx, gradeId, today, duration)
	if err != nil {
		return nil, err
	}
	return dailyPrices, nil
}

func (service *AccountService) GetTodaysByGradeId(ctx context.Context, gradeId string, date time.Time) ([]*DailyPrice, error) {
	if date.IsZero() {
		date = time.Now()
	}
	dailyPrice, err := service.repository.GetTodaysByGradeId(ctx, gradeId, date)
	if err != nil {
		return nil, err
	}
	return dailyPrice, nil
}

func (service *AccountService) GetTodaysByProductId(ctx context.Context, productId string, date time.Time) ([]*DailyPrice, error) {
	if date.IsZero() {
		date = time.Now()
	}
	dailyPrices, err := service.repository.GetTodaysByProductId(ctx, productId, date)
	if err != nil {
		return nil, err
	}
	return dailyPrices, nil
}
func (service *AccountService) GetProductsWithGradesAndPrices(ctx context.Context, date time.Time) ([]*ProductWithGrades, error) {
	if date.IsZero() {
		date = time.Now()
	}
	return service.repository.GetProductsWithGradesAndPrices(ctx, date)
}
