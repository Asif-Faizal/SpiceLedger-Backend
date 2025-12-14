package domain

import "errors"

var (
    ErrUserNotFound       = errors.New("user not found")
    ErrEmailAlreadyExists = errors.New("email already exists")
    ErrInvalidCredentials = errors.New("invalid credentials")
    ErrPriceNotFound      = errors.New("price not found for date/grade")
    ErrInsufficientStock  = errors.New("insufficient stock for sale")
    ErrInvalidGradeForProduct = errors.New("grade does not belong to product")
)
