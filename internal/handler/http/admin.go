package http

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/Asif-Faizal/SpiceLedger/internal/domain"
	"github.com/Asif-Faizal/SpiceLedger/internal/service"
)

type AdminHandler struct {
	userRepo        domain.UserRepository
	dashboardSvc    *service.DashboardService
}

func NewAdminHandler(userRepo domain.UserRepository, dashboardSvc *service.DashboardService) *AdminHandler {
	return &AdminHandler{userRepo: userRepo, dashboardSvc: dashboardSvc}
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

func (h *AdminHandler) GetDashboard(c *fiber.Ctx) error {
	dateStr := c.Query("date")
	var date time.Time
	if dateStr == "" {
		date = time.Now()
	} else {
		d, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid date format (YYYY-MM-DD)"})
		}
		date = d
	}
	resp, err := h.dashboardSvc.GetDashboard(c.Context(), date)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(resp)
}
