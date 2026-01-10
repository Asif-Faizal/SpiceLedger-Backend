package service

import (
	"context"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/Asif-Faizal/SpiceLedger/internal/domain"
)

type InventoryService struct {
	inventoryRepo domain.InventoryRepository
	priceRepo     domain.PriceRepository
	gradeRepo     domain.GradeRepository
}

func NewInventoryService(inventoryRepo domain.InventoryRepository, priceRepo domain.PriceRepository, gradeRepo domain.GradeRepository) *InventoryService {
	return &InventoryService{
		inventoryRepo: inventoryRepo,
		priceRepo:     priceRepo,
		gradeRepo:     gradeRepo,
	}
}

func (s *InventoryService) AddPurchaseLot(ctx context.Context, userID uuid.UUID, date time.Time, productID uuid.UUID, gradeID uuid.UUID, quantity, unitCost float64) error {
	grade, err := s.gradeRepo.FindByID(ctx, gradeID)
	if err != nil {
		return err
	}
	if grade.ProductID != productID {
		return domain.ErrInvalidGradeForProduct
	}
	lot := &domain.PurchaseLot{
		UserID:    userID,
		Date:      date,
		ProductID: productID,
		GradeID:   gradeID,
		Quantity:  quantity,
		UnitCost:  unitCost,
	}
	return s.inventoryRepo.CreateLot(ctx, lot)
}

func (s *InventoryService) AddSale(ctx context.Context, userID uuid.UUID, date time.Time, productID uuid.UUID, gradeID uuid.UUID, quantity, unitPrice float64) error {
	grade, err := s.gradeRepo.FindByID(ctx, gradeID)
	if err != nil {
		return err
	}
	if grade.ProductID != productID {
		return domain.ErrInvalidGradeForProduct
	}
	sale := &domain.SaleTransaction{
		UserID:    userID,
		Date:      date,
		ProductID: productID,
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

	// 2. Group by Product and Grade
	type Event struct {
		Date      time.Time
		Type      string // "BUY" or "SELL"
		Quantity  float64
		UnitPrice float64 // Cost for Buy, Price for Sell
	}

	type Key struct {
		ProductID uuid.UUID
		Product   string
		GradeID   uuid.UUID
		Grade     string
	}

	eventsByKey := make(map[Key][]Event)

	for _, lot := range lots {
		k := Key{ProductID: lot.ProductID, Product: lot.Product.Name, GradeID: lot.GradeID, Grade: lot.Grade.Name}
		eventsByKey[k] = append(eventsByKey[k], Event{Date: lot.Date, Type: "BUY", Quantity: lot.Quantity, UnitPrice: lot.UnitCost})
	}

	for _, sale := range sales {
		k := Key{ProductID: sale.ProductID, Product: sale.Product.Name, GradeID: sale.GradeID, Grade: sale.Grade.Name}
		eventsByKey[k] = append(eventsByKey[k], Event{Date: sale.Date, Type: "SELL", Quantity: sale.Quantity, UnitPrice: sale.UnitPrice})
	}

	// 3. Process each grade
	var snapshots []domain.InventorySnapshot
	var productsMap = make(map[uuid.UUID]*domain.ProductInventory)
	var totalQty, totalValuation, totalCostBasis, totalPnL float64

	dateStr := date.Format("2006-01-02")
	marketPrices, _ := s.priceRepo.GetPricesForDate(ctx, dateStr)
	priceByGradeID := make(map[uuid.UUID]float64)
	for _, p := range marketPrices {
		priceByGradeID[p.GradeID] = p.PricePerKg
	}

	for key, events := range eventsByKey {
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
		marketPrice := priceByGradeID[key.GradeID]
		marketValue := currentQty * marketPrice
		unrealizedPnL := marketValue - currentTotalCost

		snapshot := domain.InventorySnapshot{
			ProductID:      key.ProductID,
			Product:        key.Product,
			Grade:          key.Grade,
			TotalQuantity:  currentQty,
			AverageCost:    avgCost,
			TotalCostBasis: currentTotalCost,
			MarketPrice:    marketPrice,
			MarketValue:    marketValue,
			UnrealizedPnL:  unrealizedPnL,
		}
		snapshots = append(snapshots, snapshot)

		pi, ok := productsMap[key.ProductID]
		if !ok {
			pi = &domain.ProductInventory{ProductID: key.ProductID, Product: key.Product}
			productsMap[key.ProductID] = pi
		}
		pi.Grades = append(pi.Grades, snapshot)
		pi.TotalQuantity += currentQty
		pi.TotalValue += marketValue
		pi.TotalCost += currentTotalCost
		pi.TotalPnL += unrealizedPnL

		totalQty += currentQty
		totalValuation += marketValue
		totalCostBasis += currentTotalCost
		totalPnL += unrealizedPnL
	}

	products := make([]domain.ProductInventory, 0, len(productsMap))
	for _, pi := range productsMap {
		if pi.TotalCost != 0 {
			pi.TotalPnLPct = (pi.TotalPnL / pi.TotalCost) * 100
		}
		products = append(products, *pi)
	}
	totalPnLPct := 0.0
	if totalCostBasis != 0 {
		totalPnLPct = (totalPnL / totalCostBasis) * 100
	}
	return &domain.OverallInventory{
		Snapshots:     snapshots,
		Products:      products,
		TotalQuantity: totalQty,
		TotalValue:    totalValuation,
		TotalCost:     totalCostBasis,
		TotalPnL:      totalPnL,
		TotalPnLPct:   totalPnLPct,
	}, nil
}

func (s *InventoryService) GetCurrentInventory(ctx context.Context, userID uuid.UUID) (*domain.OverallInventory, error) {
	return s.GetInventoryOnDate(ctx, userID, time.Now())
}
