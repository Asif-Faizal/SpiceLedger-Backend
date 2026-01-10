package http

import (
    "github.com/gofiber/fiber/v2"
    "github.com/google/uuid"
    "github.com/Asif-Faizal/SpiceLedger/internal/service"
)

type PriceHandler struct {
	priceService *service.PriceService
}

func NewPriceHandler(priceService *service.PriceService) *PriceHandler {
	return &PriceHandler{priceService: priceService}
}

func (h *PriceHandler) SetPrice(c *fiber.Ctx) error {
    type Request struct {
        Date       string    `json:"date"`
        ProductID  string    `json:"product_id"`
        GradeID    string    `json:"grade_id"`
        PricePerKg float64   `json:"price_per_kg"`
    }
	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

    pid, err := uuid.Parse(req.ProductID)
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid product_id"})
    }
    gid, err := uuid.Parse(req.GradeID)
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid grade_id"})
    }

    if err := h.priceService.SetPrice(c.Context(), req.Date, pid, gid, req.PricePerKg); err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
    }

	return c.JSON(fiber.Map{"message": "Price set successfully"})
}

func (h *PriceHandler) GetPrice(c *fiber.Ctx) error {
    date := c.Params("date")
    productID := c.Params("product_id")
    gradeID := c.Params("grade_id")

    pid, err := uuid.Parse(productID)
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid product_id"})
    }
    gid, err := uuid.Parse(gradeID)
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid grade_id"})
    }

    price, err := h.priceService.GetPrice(c.Context(), date, pid, gid)
    if err != nil {
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Price not found"})
    }

    return c.JSON(fiber.Map{
        "date":         date,
        "product_id":   productID,
        "grade_id":     gradeID,
        "price_per_kg": price,
    })
}

func (h *PriceHandler) GetPricesForDate(c *fiber.Ctx) error {
	date := c.Params("date")

	prices, err := h.priceService.GetPricesForDate(c.Context(), date)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"date":   date,
		"prices": prices,
	})
}
