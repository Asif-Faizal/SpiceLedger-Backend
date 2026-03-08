package rest

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

type CreateOrUpdateMerchantDetailsRequest struct {
	ID          string `json:"id"`
	AccountID   string `json:"account_id"`
	PhoneNumber string `json:"phone_number"`
	Address     string `json:"address"`
	City        string `json:"city"`
	State       string `json:"state"`
	Pincode     string `json:"pincode"`
}

type MerchantDetails struct {
	ID          string `json:"id"`
	AccountID   string `json:"account_id"`
	PhoneNumber string `json:"phone_number"`
	Address     string `json:"address"`
	City        string `json:"city"`
	State       string `json:"state"`
	Pincode     string `json:"pincode"`
}

type ListAccountsResponse struct {
	Accounts []*Account `json:"accounts"`
}

type Product struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

type CreateOrUpdateProductRequest struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

type ListProductsResponse struct {
	Products []*Product `json:"products"`
}

type Grade struct {
	ID          string `json:"id"`
	ProductID   string `json:"product_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

type CreateOrUpdateGradeRequest struct {
	ID          string `json:"id"`
	ProductID   string `json:"product_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

type ListGradesByProductIdResponse struct {
	Grades []*Grade `json:"grades"`
}

type DailyPrice struct {
	ID        string  `json:"id"`
	ProductID string  `json:"product_id"`
	GradeID   string  `json:"grade_id"`
	Price     float64 `json:"price"`
	Date      string  `json:"date"`
	Time      string  `json:"time"`
}

type CreateOrUpdateDailyPriceRequest struct {
	ID        string  `json:"id"`
	ProductID string  `json:"product_id"`
	GradeID   string  `json:"grade_id"`
	Price     float64 `json:"price"`
	Date      string  `json:"date"`
	Time      string  `json:"time"`
}

type ListDailyPricesResponse struct {
	DailyPrices []*DailyPrice `json:"daily_prices"`
}

type GetTodaysPriceResponse struct {
	DailyPrices []*DailyPrice `json:"daily_prices"`
}

type GetTodaysPriceByProductIdResponse struct {
	DailyPrices []*DailyPrice `json:"daily_prices"`
}
