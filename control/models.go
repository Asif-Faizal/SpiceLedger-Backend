package control

import "time"

type Account struct {
	ID       string `json:"id" validate:"required,uuid4"`
	Name     string `json:"name" validate:"omitempty,min=3,max=50"`
	UserType string `json:"user_type" validate:"required,oneof=admin merchant"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"-" validate:"required,min=8,max=50"`
}

type Session struct {
	ID           string    `json:"id"`
	AccountID    string    `json:"account_id"`
	DeviceID     string    `json:"device_id"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
	IsRevoked    bool      `json:"is_revoked"`
}

type AuthenticatedResponse struct {
	Account      *Account `json:"account"`
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
}

type MerchantDetails struct {
	ID        string `json:"id" validate:"required,uuid4"`
	AccountID string `json:"account_id" validate:"required,uuid4"`
	Phone     string `json:"phone_number" validate:"required,min=10,max=15"`
	Address   string `json:"address" validate:"required,min=3,max=255"`
	City      string `json:"city" validate:"required,min=3,max=50"`
	State     string `json:"state" validate:"required,min=3,max=50"`
	Pincode   string `json:"pincode" validate:"required,min=6,max=6"`
}
