package util

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTClaims struct {
	AccountID string `json:"account_id"`
	UserType  string `json:"user_type"`
	Email     string `json:"email"`
	jwt.RegisteredClaims
}

func GenerateToken(accountID, userType, email, secret string, ttl time.Duration) (string, error) {
	claims := JWTClaims{
		AccountID: accountID,
		UserType:  userType,
		Email:     email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ValidateToken(tokenString, secret string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}
