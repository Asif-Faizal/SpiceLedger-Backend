package http

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/saravanan/spice_backend/internal/service"
)

type InventoryHandler struct {
	inventoryService *service.InventoryService
}

func NewInventoryHandler(inventoryService *service.InventoryService) *InventoryHandler {
	return &InventoryHandler{inventoryService: inventoryService}
}

func (h *InventoryHandler) AddLot(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	type Request struct {
		Date     string    `json:"date"`
		Quantity float64   `json:"quantity_kg"`
		UnitCost float64   `json:"unit_cost"`
		GradeID  uuid.UUID `json:"grade_id"`
	}
	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid date format (YYYY-MM-DD)"})
	}

	if err := h.inventoryService.AddPurchaseLot(c.Context(), userID, date, req.GradeID, req.Quantity, req.UnitCost); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Lot added successfully"})
}

func (h *InventoryHandler) AddSale(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	type Request struct {
		Date      string    `json:"date"`
		Quantity  float64   `json:"quantity_kg"`
		UnitPrice float64   `json:"unit_price"`
		GradeID   uuid.UUID `json:"grade_id"`
	}
	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid date format (YYYY-MM-DD)"})
	}

	if err := h.inventoryService.AddSale(c.Context(), userID, date, req.GradeID, req.Quantity, req.UnitPrice); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Sale recorded successfully"})
}

func (h *InventoryHandler) GetCurrentInventory(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	inventory, err := h.inventoryService.GetCurrentInventory(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(inventory)
}

func (h *InventoryHandler) GetInventoryOnDate(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	dateStr := c.Query("date")
	if dateStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Date parameter is required"})
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid date format (YYYY-MM-DD)"})
	}

	inventory, err := h.inventoryService.GetInventoryOnDate(c.Context(), userID, date)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(inventory)
}

// Helper to extract User ID from JWT in context (set by middleware)
func getUserID(c *fiber.Ctx) (uuid.UUID, error) {
	// Assuming middleware sets "user_id" in Locals
	idStr, ok := c.Locals("user_id").(string)
	if !ok || idStr == "" {
		// Try key from JWT claims if using standard fiber-jwt
		// This depends on how we configure the JWT middleware in routes.
		// For now assuming standard Claims extraction into Locals or simple string.

		// If using `gofiber/contrib/jwt` or similar, it might be under 'user' key.
		// We'll implement the middleware to set "user_id" string in Locals for simplicity.
		return uuid.Nil, fiber.ErrUnauthorized
	}
	return uuid.Parse(idStr)
}
