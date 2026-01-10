package scripts

import (
	"context"
	"log"

	"github.com/Asif-Faizal/SpiceLedger/internal/config"
	"github.com/Asif-Faizal/SpiceLedger/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

func SeedAdmin(ctx context.Context, userRepo domain.UserRepository, cfg *config.Config) {
	// Check if admin exists
	_, err := userRepo.FindByEmail(ctx, cfg.AdminEmail)
	if err == nil {
		log.Println("Admin user already exists, skipping seed.")
		return
	}

	// Create admin
	hashed, err := bcrypt.GenerateFromPassword([]byte(cfg.AdminPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Failed to hash admin password: %v", err)
		return
	}

	admin := &domain.User{
		Name:         "Admin User",
		Email:        cfg.AdminEmail,
		Role:         "admin",
		PasswordHash: string(hashed),
	}

	if err := userRepo.Create(ctx, admin); err != nil {
		log.Printf("Failed to create admin user: %v", err)
	} else {
		log.Println("Admin user created successfully.")
	}
}
