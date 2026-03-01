package util

type ContextKey string

const (
	UserTypeSuperAdmin = "super_admin"
	UserTypeAdmin      = "admin"
	UserTypeMerchant   = "merchant"
	UserTypeCustomer   = "customer"

	AccountIDKey ContextKey = "account_id"
	UserTypeKey  ContextKey = "user_type"
	EmailKey     ContextKey = "email"
)
