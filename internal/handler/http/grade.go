package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/saravanan/spice_backend/internal/service"
)

type GradeHandler struct {
	gradeSvc *service.GradeService
}

func NewGradeHandler(gradeSvc *service.GradeService) *GradeHandler {
	return &GradeHandler{gradeSvc: gradeSvc}
}

func (h *GradeHandler) CreateGrade(c *fiber.Ctx) error {
	type Request struct {
		ProductID   string `json:"product_id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Grade name is required"})
	}

	// Validate product_id
	pid, err := uuid.Parse(req.ProductID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid product_id"})
	}

	if err := h.gradeSvc.CreateGrade(c.Context(), pid, req.Name, req.Description); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Grade created successfully"})
}

func (h *GradeHandler) ListGrades(c *fiber.Ctx) error {
	grades, err := h.gradeSvc.ListGrades(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(grades)
}
