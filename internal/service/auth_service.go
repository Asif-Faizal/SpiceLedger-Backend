package service

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/saravanan/spice_backend/internal/config"
	"github.com/saravanan/spice_backend/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo domain.UserRepository
	config   *config.Config
}

func NewAuthService(userRepo domain.UserRepository, config *config.Config) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		config:   config,
	}
}

func (s *AuthService) Register(ctx context.Context, name, email, password string) error {
	existing, _ := s.userRepo.FindByEmail(ctx, email)
	if existing != nil {
		return domain.ErrEmailAlreadyExists
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := &domain.User{
		Name:         name,
		Email:        email,
		PasswordHash: string(hashed),
	}

	return s.userRepo.Create(ctx, user)
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, string, error) {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return "", "", domain.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", "", domain.ErrInvalidCredentials
	}

	// Generate Token Pair
	accessToken, refreshToken, err := s.GenerateTokenPair(user.ID.String(), user.Role)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *AuthService) GenerateTokenPair(userID, role string) (string, string, error) {
	// 1. Access Token (15 min)
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"type":    "access",
		"exp":     time.Now().Add(time.Minute * 15).Unix(),
	})
	accessTokenString, err := accessToken.SignedString([]byte(s.config.JWTSecret))
	if err != nil {
		return "", "", err
	}

	// 2. Refresh Token (7 days)
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"type":    "refresh",
		"exp":     time.Now().Add(time.Hour * 24 * 7).Unix(),
	})
	refreshTokenString, err := refreshToken.SignedString([]byte(s.config.JWTSecret))
	if err != nil {
		return "", "", err
	}

	return accessTokenString, refreshTokenString, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshTokenStr string) (string, string, error) {
	token, err := jwt.Parse(refreshTokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, domain.ErrInvalidCredentials // Reusing invalid credentials for signature error
		}
		return []byte(s.config.JWTSecret), nil
	})

	if err != nil || !token.Valid {
		return "", "", domain.ErrInvalidCredentials
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", "", domain.ErrInvalidCredentials
	}

	// Validate token type
	if claims["type"] != "refresh" {
		return "", "", domain.ErrInvalidCredentials
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return "", "", domain.ErrInvalidCredentials
	}

	role, ok := claims["role"].(string)
	if !ok {
		role = "user"
	}

	// Rotate tokens: Issue new pair
	return s.GenerateTokenPair(userID, role)
}
