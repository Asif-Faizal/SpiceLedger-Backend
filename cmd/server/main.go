package main

import (
	"context"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/saravanan/spice_backend/internal/config"
	"github.com/saravanan/spice_backend/internal/domain"
	"github.com/saravanan/spice_backend/internal/handler/http"
	"github.com/saravanan/spice_backend/internal/repository/postgres"
	"github.com/saravanan/spice_backend/internal/repository/redis"
	"github.com/saravanan/spice_backend/internal/scripts"
	"github.com/saravanan/spice_backend/internal/service"
	"github.com/saravanan/spice_backend/pkg/db"
	pkgRedis "github.com/saravanan/spice_backend/pkg/redis"
)

func main() {
	// 1. Load Config
	cfg := config.LoadConfig()

	// 2. Init DB & Redis
	database := db.Init(cfg.DBDSN)
	redisClient := pkgRedis.Init(cfg.RedisAddr, cfg.RedisPassword)

	// Ensure database migrations for Grades (auto-migrate if possible or assume handled)
	// We'll trust the plan which said database changes but didn't specify migration tool usage in code.
	// But `db.Init` usually does GORM connection.
	// To be safe, we can run AutoMigrate here if db package exposes GORM, but `db.Init` returns *gorm.DB.
	// Let's add AutoMigrate for Grade and Update User
	database.AutoMigrate(&domain.User{}, &domain.Grade{})

	// 3. Init Repositories
	userRepo := postgres.NewUserRepository(database)
	inventoryRepo := postgres.NewInventoryRepository(database)
	priceRepo := redis.NewPriceRepository(redisClient)
	gradeRepo := postgres.NewGradeRepository(database)

	// 4. Init Services
	authService := service.NewAuthService(userRepo, cfg)
	inventoryService := service.NewInventoryService(inventoryRepo, priceRepo)
	priceService := service.NewPriceService(priceRepo)
	gradeService := service.NewGradeService(gradeRepo)

	// 5. Seed Admin User
	// We create a background context for seeding since main context is not cancellable here yet
	scripts.SeedAdmin(context.Background(), userRepo, cfg)

	// 6. Setup Fiber
	app := fiber.New()

	// 7. Setup Routes
	http.SetupRoutes(app, cfg, authService, inventoryService, priceService, gradeService, userRepo)

	// 8. Start Server
	log.Printf("Server starting on port %s", cfg.Port)
	log.Fatal(app.Listen(":" + cfg.Port))
}
