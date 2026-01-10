package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/Asif-Faizal/SpiceLedger/internal/config"
	"github.com/Asif-Faizal/SpiceLedger/internal/handler/http"
	"github.com/Asif-Faizal/SpiceLedger/internal/repository/postgres"
	"github.com/Asif-Faizal/SpiceLedger/internal/repository/redis"
	"github.com/Asif-Faizal/SpiceLedger/internal/service"
	"github.com/Asif-Faizal/SpiceLedger/pkg/db"
	pkgRedis "github.com/Asif-Faizal/SpiceLedger/pkg/redis"
)

func main() {
	// 1. Load Config
	cfg := config.LoadConfig()

	// 2. Init DB & Redis
	database := db.Init(cfg.DBDSN)
	redisClient := pkgRedis.Init(cfg.RedisAddr, cfg.RedisPassword)

	// 3. Init Repositories
	userRepo := postgres.NewUserRepository(database)
	inventoryRepo := postgres.NewInventoryRepository(database)
	priceRepo := redis.NewPriceRepository(redisClient)
    gradeRepo := postgres.NewGradeRepository(database)
    productRepo := postgres.NewProductRepository(database)

	// 4. Init Services
	authService := service.NewAuthService(userRepo, cfg)
    inventoryService := service.NewInventoryService(inventoryRepo, priceRepo, gradeRepo)
    priceService := service.NewPriceService(priceRepo)
    gradeService := service.NewGradeService(gradeRepo)
    productService := service.NewProductService(productRepo)
    dashboardService := service.NewDashboardService(userRepo, productRepo, gradeRepo, priceRepo)

	// 5. Setup Fiber
	app := fiber.New()

	// 7. Setup Routes
    http.SetupRoutes(app, cfg, authService, inventoryService, priceService, gradeService, productService, userRepo, dashboardService)

	// 8. Start Server
	log.Printf("Server starting on port %s", cfg.Port)
	log.Fatal(app.Listen(":" + cfg.Port))
}
