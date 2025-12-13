package http

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/golang-jwt/jwt/v5"
	"github.com/saravanan/spice_backend/internal/config"
	"github.com/saravanan/spice_backend/internal/domain"
	"github.com/saravanan/spice_backend/internal/service"
)

func SetupRoutes(app *fiber.App, cfg *config.Config, authSvc *service.AuthService, invSvc *service.InventoryService, priceSvc *service.PriceService, gradeSvc *service.GradeService, userRepo domain.UserRepository) {
	app.Use(logger.New())
	app.Use(cors.New())

	authHandler := NewAuthHandler(authSvc)
	inventoryHandler := NewInventoryHandler(invSvc)
	priceHandler := NewPriceHandler(priceSvc)
	gradeHandler := NewGradeHandler(gradeSvc)
	adminHandler := NewAdminHandler(userRepo)

	api := app.Group("/api")

	// Auth
	auth := api.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)

	// Protected Routes
	api.Use(authMiddleware(cfg.JWTSecret))

	// Admin Routes
	admin := api.Group("/admin")
	admin.Use(adminMiddleware())
	admin.Get("/stats", adminHandler.GetUserStats)

	// Grades
	grades := api.Group("/grades")
	grades.Get("/", gradeHandler.ListGrades)
	grades.Post("/", adminMiddleware(), gradeHandler.CreateGrade) // Only admin can create grades

	// Inventory (Lots & Sales)
	api.Post("/lots", inventoryHandler.AddLot)
	api.Post("/sales", inventoryHandler.AddSale)

	// Reports
	inventory := api.Group("/inventory")
	inventory.Get("/current", inventoryHandler.GetCurrentInventory)
	inventory.Get("/on-date", inventoryHandler.GetInventoryOnDate) // ?date=YYYY-MM-DD

	// Prices
	prices := api.Group("/prices")
	prices.Post("/", adminMiddleware(), priceHandler.SetPrice) // Admin only
	prices.Get("/:date/:grade", priceHandler.GetPrice)
	prices.Get("/:date", priceHandler.GetPricesForDate)
}

func authMiddleware(secret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Missing authorization header"})
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid authorization header format"})
		}

		tokenStr := parts[1]
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.ErrUnauthorized
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token"})
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token claims"})
		}

		userID, ok := claims["user_id"].(string)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user_id in token"})
		}

		role, ok := claims["role"].(string)
		if !ok {
			// fallback/default if not present (e.g. old tokens, though we just changed it)
			role = "user"
		}

		c.Locals("user_id", userID)
		c.Locals("role", role)
		return c.Next()
	}
}

func adminMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role := c.Locals("role")
		if role != "admin" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Admins only"})
		}
		return c.Next()
	}
}
