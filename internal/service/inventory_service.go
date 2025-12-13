package service

import (
	"context"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/saravanan/spice_backend/internal/domain"
)

type InventoryService struct {
	inventoryRepo domain.InventoryRepository
	priceRepo     domain.PriceRepository
}

func NewInventoryService(inventoryRepo domain.InventoryRepository, priceRepo domain.PriceRepository) *InventoryService {
	return &InventoryService{
		inventoryRepo: inventoryRepo,
		priceRepo:     priceRepo,
	}
}

func (s *InventoryService) AddPurchaseLot(ctx context.Context, userID uuid.UUID, date time.Time, gradeID uuid.UUID, quantity, unitCost float64) error {
	lot := &domain.PurchaseLot{
		UserID:   userID,
		Date:     date,
		GradeID:  gradeID,
		Quantity: quantity,
		UnitCost: unitCost,
	}
	return s.inventoryRepo.CreateLot(ctx, lot)
}

func (s *InventoryService) AddSale(ctx context.Context, userID uuid.UUID, date time.Time, gradeID uuid.UUID, quantity, unitPrice float64) error {
	sale := &domain.SaleTransaction{
		UserID:    userID,
		Date:      date,
		GradeID:   gradeID,
		Quantity:  quantity,
		UnitPrice: unitPrice,
	}
	return s.inventoryRepo.CreateSale(ctx, sale)
}

// GetInventoryOnDate calculates the inventory state for a specific date
func (s *InventoryService) GetInventoryOnDate(ctx context.Context, userID uuid.UUID, date time.Time) (*domain.OverallInventory, error) {
	// 1. Fetch all lots and sales up to this date
	lots, sales, err := s.inventoryRepo.GetHistory(ctx, userID, date)
	if err != nil {
		return nil, err
	}

	// 2. Group by Grade
	type Event struct {
		Date      time.Time
		Type      string // "BUY" or "SELL"
		Quantity  float64
		UnitPrice float64 // Cost for Buy, Price for Sell
	}

	eventsByGrade := make(map[string][]Event)

	for _, lot := range lots {
		gradeName := lot.Grade.Name
		if gradeName == "" {
			gradeName = "Unknown" // Should ideally be preloaded
		}
		eventsByGrade[gradeName] = append(eventsByGrade[gradeName], Event{
			Date:      lot.Date,
			Type:      "BUY",
			Quantity:  lot.Quantity,
			UnitPrice: lot.UnitCost,
		})
	}

	for _, sale := range sales {
		gradeName := sale.Grade.Name
		if gradeName == "" {
			gradeName = "Unknown"
		}
		eventsByGrade[gradeName] = append(eventsByGrade[gradeName], Event{
			Date:      sale.Date,
			Type:      "SELL",
			Quantity:  sale.Quantity,
			UnitPrice: sale.UnitPrice,
		})
	}

	// 3. Process each grade
	var snapshots []domain.InventorySnapshot
	var totalQty, totalValuation, totalCostBasis, totalPnL float64

	dateStr := date.Format("2006-01-02")
	marketPrices, _ := s.priceRepo.GetPricesForDate(ctx, dateStr)
	marketPriceMap := make(map[string]float64)
	for _, p := range marketPrices {
		marketPriceMap[p.Grade] = p.PricePerKg
	}

	for grade, events := range eventsByGrade {
		// Sort events by date
		sort.Slice(events, func(i, j int) bool {
			return events[i].Date.Before(events[j].Date)
		})

		var currentQty, currentTotalCost float64

		for _, event := range events {
			if event.Type == "BUY" {
				currentQty += event.Quantity
				currentTotalCost += event.Quantity * event.UnitPrice
			} else if event.Type == "SELL" {
				if currentQty > 0 {
					avgCost := currentTotalCost / currentQty
					currentQty -= event.Quantity
					currentTotalCost -= event.Quantity * avgCost
				} else {
					// Handle negative stock if it happens
					currentQty -= event.Quantity
				}
			}
		}

		// Avoid tiny floating point errors
		if currentQty < 0.0001 && currentQty > -0.0001 {
			currentQty = 0
			currentTotalCost = 0
		}

		avgCost := 0.0
		if currentQty > 0 {
			avgCost = currentTotalCost / currentQty
		}

		// Validation with market price
		marketPrice := marketPriceMap[grade]
		marketValue := currentQty * marketPrice
		unrealizedPnL := marketValue - currentTotalCost

		snapshot := domain.InventorySnapshot{
			Grade:          grade,
			TotalQuantity:  currentQty,
			AverageCost:    avgCost,
			TotalCostBasis: currentTotalCost,
			MarketPrice:    marketPrice,
			MarketValue:    marketValue,
			UnrealizedPnL:  unrealizedPnL,
		}
		snapshots = append(snapshots, snapshot)

		totalQty += currentQty
		totalValuation += marketValue
		totalCostBasis += currentTotalCost
		totalPnL += unrealizedPnL
	}

	return &domain.OverallInventory{
		Snapshots:     snapshots,
		TotalQuantity: totalQty,
		TotalValue:    totalValuation,
		TotalCost:     totalCostBasis,
		TotalPnL:      totalPnL,
	}, nil
}

func (s *InventoryService) GetCurrentInventory(ctx context.Context, userID uuid.UUID) (*domain.OverallInventory, error) {
	return s.GetInventoryOnDate(ctx, userID, time.Now())
}
