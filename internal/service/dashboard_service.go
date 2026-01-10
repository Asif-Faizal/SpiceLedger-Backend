package service

import (
	"context"
	"time"

	"github.com/Asif-Faizal/SpiceLedger/internal/domain"
	"github.com/google/uuid"
)

type DashboardService struct {
	userRepo    domain.UserRepository
	productRepo domain.ProductRepository
	gradeRepo   domain.GradeRepository
	priceRepo   domain.PriceRepository
}

func NewDashboardService(
	userRepo domain.UserRepository,
	productRepo domain.ProductRepository,
	gradeRepo domain.GradeRepository,
	priceRepo domain.PriceRepository,
) *DashboardService {
	return &DashboardService{
		userRepo:    userRepo,
		productRepo: productRepo,
		gradeRepo:   gradeRepo,
		priceRepo:   priceRepo,
	}
}

func startOfWeek(t time.Time) time.Time {
	w := int(t.Weekday())
	if w == 0 {
		w = 7
	}
	delta := w - 1
	monday := t.AddDate(0, 0, -delta)
	return time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, t.Location())
}

func startOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

func (s *DashboardService) GetDashboard(ctx context.Context, date time.Time) (*domain.DashboardResponse, error) {
	now := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())

	totalUsers, err := s.userRepo.Count(ctx)
	if err != nil {
		return nil, err
	}

	weekStart := startOfWeek(now)
	prevWeekStart := weekStart.AddDate(0, 0, -7)
	prevWeekEnd := weekStart
	weeklyNew, err := s.userRepo.CountCreatedBetween(ctx, weekStart, now.AddDate(0, 0, 1))
	if err != nil {
		return nil, err
	}
	prevWeeklyNew, err := s.userRepo.CountCreatedBetween(ctx, prevWeekStart, prevWeekEnd)
	if err != nil {
		return nil, err
	}

	monthStart := startOfMonth(now)
	prevMonthStart := monthStart.AddDate(0, -1, 0)
	prevMonthEnd := monthStart
	currentMonthTotal := totalUsers
	prevMonthTotal, err := s.userRepo.CountCreatedBetween(ctx, time.Time{}, prevMonthEnd)
	if err != nil {
		return nil, err
	}

	totalProducts, err := s.productRepo.Count(ctx)
	if err != nil {
		return nil, err
	}
	currentMonthProductsNew, err := s.productRepo.CountCreatedBetween(ctx, monthStart, now.AddDate(0, 0, 1))
	if err != nil {
		return nil, err
	}
	prevMonthProductsNew, err := s.productRepo.CountCreatedBetween(ctx, prevMonthStart, prevMonthEnd)
	if err != nil {
		return nil, err
	}

	totalGrades, err := s.gradeRepo.Count(ctx)
	if err != nil {
		return nil, err
	}

	priceDate := now.Format("2006-01-02")
	prevPriceDate := now.AddDate(0, 0, -1).Format("2006-01-02")
	currentPrices, err := s.priceRepo.GetPricesForDate(ctx, priceDate)
	if err != nil {
		return nil, err
	}

	grades, err := s.gradeRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	products, err := s.productRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	productByID := make(map[uuid.UUID]string)
	for _, p := range products {
		productByID[p.ID] = p.Name
	}
	gradeInfo := make(map[uuid.UUID]struct {
		name      string
		productID uuid.UUID
	})
	for _, g := range grades {
		gradeInfo[g.ID] = struct {
			name      string
			productID uuid.UUID
		}{name: g.Name, productID: g.ProductID}
	}

	var priceUpdates []domain.DashboardPriceUpdate
	for _, cp := range currentPrices {
		prevPrice, _ := s.priceRepo.GetPrice(ctx, prevPriceDate, cp.ProductID, cp.GradeID)
		delta := cp.PricePerKg - prevPrice
		var pct float64
		if prevPrice != 0 {
			pct = (delta / prevPrice) * 100
		}
		gi := gradeInfo[cp.GradeID]
		productName := productByID[gi.productID]
		priceUpdates = append(priceUpdates, domain.DashboardPriceUpdate{
			Date:          priceDate,
			ProductID:     cp.ProductID,
			Product:       productName,
			GradeID:       cp.GradeID,
			Grade:         gi.name,
			Price:         cp.PricePerKg,
			PreviousDate:  prevPriceDate,
			PreviousPrice: prevPrice,
			ChangeDelta:   delta,
			ChangePercent: pct,
		})
	}

	var weeklyChangePct float64
	if prevWeeklyNew != 0 {
		weeklyChangePct = (float64(weeklyNew-prevWeeklyNew) / float64(prevWeeklyNew)) * 100
	}
	var monthlyUserChangePct float64
	if prevMonthTotal != 0 {
		monthlyUserChangePct = (float64(currentMonthTotal-prevMonthTotal) / float64(prevMonthTotal)) * 100
	}

	var monthlyProductChangePct float64
	if prevMonthProductsNew != 0 {
		monthlyProductChangePct = (float64(currentMonthProductsNew-prevMonthProductsNew) / float64(prevMonthProductsNew)) * 100
	}

	resp := &domain.DashboardResponse{
		Date: now.Format("2006-01-02"),
		Users: domain.DashboardUsersSummary{
			Total:            totalUsers,
			WeeklyNew:        weeklyNew,
			WeeklyChangePct:  weeklyChangePct,
			MonthlyChangePct: monthlyUserChangePct,
		},
		Products: domain.DashboardProductsSummary{
			Total:            totalProducts,
			MonthlyChangePct: monthlyProductChangePct,
		},
		Grades: domain.DashboardGradesSummary{
			Total: totalGrades,
		},
		TotalItems:   totalProducts + totalGrades,
		PriceUpdates: priceUpdates,
	}
	return resp, nil
}
