package service

import (
	"context"

	"github.com/saravanan/spice_backend/internal/domain"
)

type PriceService struct {
	priceRepo domain.PriceRepository
}

func NewPriceService(priceRepo domain.PriceRepository) *PriceService {
	return &PriceService{priceRepo: priceRepo}
}

func (s *PriceService) SetPrice(ctx context.Context, date string, grade string, price float64) error {
	return s.priceRepo.SetPrice(ctx, date, grade, price)
}

func (s *PriceService) GetPrice(ctx context.Context, date string, grade string) (float64, error) {
	return s.priceRepo.GetPrice(ctx, date, grade)
}

func (s *PriceService) GetPricesForDate(ctx context.Context, date string) ([]domain.DailyPrice, error) {
	return s.priceRepo.GetPricesForDate(ctx, date)
}
