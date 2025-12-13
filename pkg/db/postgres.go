package db

import (
	"log"

	"github.com/saravanan/spice_backend/internal/domain"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Init(dsn string) *gorm.DB {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Database connection established")

	// Auto Migrate
	err = db.AutoMigrate(&domain.User{}, &domain.PurchaseLot{}, &domain.SaleTransaction{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	return db
}
