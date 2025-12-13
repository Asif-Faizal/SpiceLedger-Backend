package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/saravanan/spice_backend/internal/domain"
)

type AdminHandler struct {
	userRepo domain.UserRepository
}

func NewAdminHandler(userRepo domain.UserRepository) *AdminHandler {
	return &AdminHandler{userRepo: userRepo}
}

func (h *AdminHandler) GetUserStats(c *fiber.Ctx) error {
	count, err := h.userRepo.Count(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"total_users": count,
	})
}
