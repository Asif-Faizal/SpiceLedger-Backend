package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/saravanan/spice_backend/internal/service"
)

type PriceHandler struct {
	priceService *service.PriceService
}

func NewPriceHandler(priceService *service.PriceService) *PriceHandler {
	return &PriceHandler{priceService: priceService}
}

func (h *PriceHandler) SetPrice(c *fiber.Ctx) error {
	type Request struct {
		Date       string  `json:"date"`
		Grade      string  `json:"grade"`
		PricePerKg float64 `json:"price_per_kg"`
	}
	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := h.priceService.SetPrice(c.Context(), req.Date, req.Grade, req.PricePerKg); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Price set successfully"})
}

func (h *PriceHandler) GetPrice(c *fiber.Ctx) error {
	date := c.Params("date")
	grade := c.Params("grade")

	price, err := h.priceService.GetPrice(c.Context(), date, grade)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Price not found"})
	}

	return c.JSON(fiber.Map{
		"date":         date,
		"grade":        grade,
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
