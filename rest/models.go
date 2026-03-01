package rest

import pb "github.com/Asif-Faizal/SpiceLedger-Backend/control/pb"

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	DeviceID string `json:"device_id"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
	DeviceID     string `json:"device_id"`
}

type LogoutRequest struct {
	AccessToken string `json:"access_token"`
	DeviceID    string `json:"device_id"`
}

type Account struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	UserType string `json:"user_type"`
	Email    string `json:"email"`
}

type AuthenticatedResponse struct {
	Account      *Account `json:"account"`
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
}

type CreateOrUpdateAccountRequest struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	UserType string `json:"user_type"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ListAccountsResponse struct {
	Accounts []*Account `json:"accounts"`
}

func toAccount(a *pb.Account) *Account {
	if a == nil {
		return nil
	}
	return &Account{
		ID:       a.Id,
		Name:     a.Name,
		UserType: a.Usertype,
		Email:    a.Email,
	}
}
