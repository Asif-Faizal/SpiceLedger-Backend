package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/Asif-Faizal/SpiceLedger/internal/domain"
)

type PriceService struct {
	priceRepo domain.PriceRepository
}

func NewPriceService(priceRepo domain.PriceRepository) *PriceService {
	return &PriceService{priceRepo: priceRepo}
}

func (s *PriceService) SetPrice(ctx context.Context, date string, productID uuid.UUID, gradeID uuid.UUID, price float64) error {
	return s.priceRepo.SetPrice(ctx, date, productID, gradeID, price)
}

func (s *PriceService) GetPrice(ctx context.Context, date string, productID uuid.UUID, gradeID uuid.UUID) (float64, error) {
	return s.priceRepo.GetPrice(ctx, date, productID, gradeID)
}

func (s *PriceService) GetPricesForDate(ctx context.Context, date string) ([]domain.DailyPrice, error) {
	return s.priceRepo.GetPricesForDate(ctx, date)
}
