package graphql

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/Asif-Faizal/SpiceLedger-Backend/control/pb"
	marketpb "github.com/Asif-Faizal/SpiceLedger-Backend/market/pb"
)

// Products is the resolver for the products field.
func (r *queryResolver) Products(ctx context.Context, date *string, search *string) ([]*ProductWithGradesAndPrice, error) {
	dateStr := ""
	if date != nil {
		dateStr = *date
	} else {
		dateStr = time.Now().Format("2006-01-02")
	}

	searchStr := ""
	if search != nil {
		searchStr = *search
	}

	resp, err := r.server.controlClient.GetProductsWithGradesAndPrices(ctx, &pb.GetProductsWithGradesAndPricesRequest{
		Date:   dateStr,
		Search: searchStr,
	})
	if err != nil {
		return nil, err
	}

	products := make([]*ProductWithGradesAndPrice, len(resp.Products))
	for i, p := range resp.Products {
		grades := make([]*GradeWithPrice, len(p.Grades))
		for j, g := range p.Grades {
			grades[j] = &GradeWithPrice{
				ID:          g.Id,
				ProductID:   g.ProductId,
				Name:        g.Name,
				Description: g.Description,
				Status:      g.Status,
				Price:       g.Price,
			}
		}

		products[i] = &ProductWithGradesAndPrice{
			ID:          p.Id,
			Name:        p.Name,
			Category:    p.Category,
			Description: p.Description,
			Status:      p.Status,
			Grades:      grades,
		}
	}
	return products, nil
}

// GetGradePosition is the resolver for the getGradePosition field.
func (r *queryResolver) GetGradePosition(ctx context.Context, spiceGradeID string) (*PositionView, error) {
	resp, err := r.server.marketClient.GetGradePosition(ctx, &marketpb.GetGradePositionRequest{
		SpiceGradeId: spiceGradeID,
	})
	if err != nil {
		return nil, err
	}
	return &PositionView{
		UserID:        resp.Position.UserId,
		SpiceGradeID:  resp.Position.SpiceGradeId,
		TotalQty:      resp.Position.TotalQty,
		TotalCost:     resp.Position.TotalCost,
		AvgCost:       resp.Position.AvgCost,
		TodayPrice:    resp.Position.TodayPrice,
		RealizedPnL:   resp.Position.RealizedPnl,
		UnrealizedPnL: resp.Position.UnrealizedPnl,
		UpdatedAt:     resp.Position.UpdatedAt,
	}, nil
}

// GetPositions is the resolver for the getPositions field.
func (r *queryResolver) GetPositions(ctx context.Context) ([]*PositionView, error) {
	resp, err := r.server.marketClient.GetPositions(ctx, &marketpb.GetPositionsRequest{})
	if err != nil {
		return nil, err
	}
	positions := make([]*PositionView, len(resp.Positions))
	for i, p := range resp.Positions {
		positions[i] = &PositionView{
			UserID:        p.UserId,
			SpiceGradeID:  p.SpiceGradeId,
			TotalQty:      p.TotalQty,
			TotalCost:     p.TotalCost,
			AvgCost:       p.AvgCost,
			TodayPrice:    p.TodayPrice,
			RealizedPnL:   p.RealizedPnl,
			UnrealizedPnL: p.UnrealizedPnl,
			UpdatedAt:     p.UpdatedAt,
		}
	}
	return positions, nil
}

// ListGradeTransactions is the resolver for the listGradeTransactions field.
func (r *queryResolver) ListGradeTransactions(ctx context.Context, spiceGradeID string, skip *int, take *int) ([]*Transaction, error) {
	var skip32, take32 uint32
	if skip != nil {
		skip32 = uint32(*skip)
	}
	if take != nil {
		take32 = uint32(*take)
	}
	resp, err := r.server.marketClient.ListGradeTransactions(ctx, &marketpb.ListGradeTransactionsRequest{
		SpiceGradeId: spiceGradeID,
		Skip:         skip32,
		Take:         take32,
	})
	if err != nil {
		return nil, err
	}
	transactions := make([]*Transaction, len(resp.Transactions))
	for i, t := range resp.Transactions {
		transactions[i] = &Transaction{
			ID:           t.Id,
			UserID:       t.UserId,
			SpiceGradeID: t.SpiceGradeId,
			Type:         t.Type,
			Quantity:     t.Quantity,
			Price:        t.Price,
			TradeDate:    t.TradeDate,
			CreatedAt:    t.CreatedAt,
		}
	}
	return transactions, nil
}

// AdminDashboard is the resolver for the adminDashboard field.
func (r *queryResolver) AdminDashboard(ctx context.Context) (*AdminDashboard, error) {
	// 1. Get system metrics (users, products)
	systemResp, err := r.server.controlClient.GetSystemMetrics(ctx, &pb.GetSystemMetricsRequest{})
	if err != nil {
		return nil, err
	}

	// 2. Get market metrics (transactions, volume, top products)
	marketResp, err := r.server.marketClient.GetMarketMetrics(ctx, &marketpb.GetMarketMetricsRequest{})
	if err != nil {
		return nil, err
	}

	// 3. Get recent transactions (last 5 across all users for admin)
	txns, err := r.server.marketClient.ListTransactions(ctx, &marketpb.ListTransactionsRequest{
		Take: 5,
	})
	if err != nil {
		return nil, err
	}

	recentTransactions := make([]*Transaction, len(txns.Transactions))
	for i, t := range txns.Transactions {
		recentTransactions[i] = &Transaction{
			ID:           t.Id,
			UserID:       t.UserId,
			SpiceGradeID: t.SpiceGradeId,
			Type:         t.Type,
			Quantity:     t.Quantity,
			Price:        t.Price,
			TradeDate:    t.TradeDate,
			CreatedAt:    t.CreatedAt,
		}
	}

	topProducts := make([]*TopProduct, len(marketResp.TopProducts))
	for i, p := range marketResp.TopProducts {
		topProducts[i] = &TopProduct{
			Name:   fmt.Sprintf("%s - %s", p.ProductName, p.GradeName),
			Volume: p.Volume,
		}
	}

	return &AdminDashboard{
		TotalUsers:         int(systemResp.TotalUsers),
		TotalProducts:      int(systemResp.TotalProducts),
		TotalTransactions:  int(marketResp.TotalTransactions),
		TotalVolume:        marketResp.TotalVolume,
		RecentTransactions: recentTransactions,
		TopProducts:        topProducts,
	}, nil
}

// MerchantDashboard is the resolver for the merchantDashboard field.
func (r *queryResolver) MerchantDashboard(ctx context.Context, days *int) (*MerchantDashboard, error) {
	windowDays := uint32(7)
	if days != nil && *days > 0 {
		windowDays = uint32(*days)
		if windowDays > 90 {
			windowDays = 90
		}
	}

	// 1. Get enriched holdings (JWT-scoped)
	holdingsResp, err := r.server.marketClient.GetHoldings(ctx, &marketpb.GetHoldingsRequest{})
	if err != nil {
		return nil, err
	}

	// 2. Get trade stats for the period
	statsResp, err := r.server.marketClient.GetTradeStats(ctx, &marketpb.GetTradeStatsRequest{Days: windowDays})
	if err != nil {
		return nil, err
	}

	// 3. Get realized P&L history
	pnlResp, err := r.server.marketClient.GetRealizedPnLHistory(ctx, &marketpb.GetRealizedPnLHistoryRequest{Days: windowDays})
	if err != nil {
		return nil, err
	}

	// 4. Get trade activity
	activityResp, err := r.server.marketClient.GetTradeActivity(ctx, &marketpb.GetTradeActivityRequest{Days: windowDays})
	if err != nil {
		return nil, err
	}

	// 5. Get price snapshots for held grades
	snapshotsResp, err := r.server.marketClient.GetPriceSnapshots(ctx, &marketpb.GetPriceSnapshotsRequest{})
	if err != nil {
		return nil, err
	}

	// 6. Get recent transactions (last 5 for merchant)
	txnsResp, err := r.server.marketClient.ListTransactions(ctx, &marketpb.ListTransactionsRequest{
		Take: 5,
	})
	if err != nil {
		return nil, err
	}

	holdings := make([]*MerchantHolding, 0, len(holdingsResp.Holdings))
	var totalPortfolioValue float64
	for _, row := range holdingsResp.Holdings {
		h := &MerchantHolding{
			SpiceGradeID: row.SpiceGradeId,
			ProductName:  row.ProductName,
			GradeName:    row.GradeName,
			Quantity:     row.Quantity,
			CostBasis:    row.TotalCost,
			RealizedPnL:  row.RealizedPnl,
			TodayPrice:   row.TodayPrice,
		}
		if row.Quantity > 0 {
			h.AvgCost = row.TotalCost / row.Quantity
		}
		if row.TodayPrice > 0 {
			h.MarketValue = row.Quantity * row.TodayPrice
			h.UnrealizedPnL = (row.TodayPrice - h.AvgCost) * row.Quantity
		} else {
			h.MarketValue = row.TotalCost
		}
		if row.TotalCost > 0 {
			h.UnrealizedPnLPercent = (h.UnrealizedPnL / row.TotalCost) * 100
		}
		totalPortfolioValue += h.MarketValue
		holdings = append(holdings, h)
	}
	if totalPortfolioValue > 0 {
		for _, h := range holdings {
			h.WeightPercent = (h.MarketValue / totalPortfolioValue) * 100
		}
	}

	portfolioMix := make([]*PortfolioSlice, len(holdings))
	for i, h := range holdings {
		portfolioMix[i] = &PortfolioSlice{
			Label:    fmt.Sprintf("%s - %s", h.ProductName, h.GradeName),
			Value:    h.MarketValue,
			Quantity: h.Quantity,
		}
	}

	summary := &MerchantSummary{
		OpenPositions:      len(holdings),
		TradesInPeriod:     int(statsResp.TradesInPeriod),
		BuyVolumeInPeriod:  statsResp.BuyVolumeInPeriod,
		SellVolumeInPeriod: statsResp.SellVolumeInPeriod,
	}
	for _, h := range holdings {
		summary.TotalCost += h.CostBasis
		summary.TotalRealizedPnL += h.RealizedPnL
		summary.TotalUnrealizedPnL += h.UnrealizedPnL
		summary.PortfolioValue += h.MarketValue
		summary.TotalQuantityKg += h.Quantity
	}
	summary.NetPnL = summary.TotalRealizedPnL + summary.TotalUnrealizedPnL

	pnlByDate := make(map[string]float64, len(pnlResp.Rows))
	for _, row := range pnlResp.Rows {
		pnlByDate[row.Date] += row.Amount
	}
	pnlTrend := make([]*PnLPoint, 0, windowDays+1)
	var cumulativePnL float64
	for i := int(windowDays); i >= 0; i-- {
		key := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		daily := pnlByDate[key]
		cumulativePnL += daily
		pnlTrend = append(pnlTrend, &PnLPoint{
			Date:                  key,
			DailyRealizedPnL:      daily,
			CumulativeRealizedPnL: cumulativePnL,
		})
	}

	activityByDate := make(map[string]*ActivityDay, len(activityResp.Rows))
	for _, row := range activityResp.Rows {
		day, ok := activityByDate[row.Date]
		if !ok {
			day = &ActivityDay{Date: row.Date}
			activityByDate[row.Date] = day
		}
		switch row.Type {
		case "BUY":
			day.BuyQuantity += row.Quantity
			day.BuyCount += int(row.Count)
		case "SELL":
			day.SellQuantity += row.Quantity
			day.SellCount += int(row.Count)
		}
	}
	activityTrend := make([]*ActivityDay, 0, windowDays+1)
	for i := int(windowDays); i >= 0; i-- {
		key := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		if day, ok := activityByDate[key]; ok {
			activityTrend = append(activityTrend, day)
		} else {
			activityTrend = append(activityTrend, &ActivityDay{Date: key})
		}
	}

	recentTransactions := make([]*Transaction, len(txnsResp.Transactions))
	for i, t := range txnsResp.Transactions {
		recentTransactions[i] = &Transaction{
			ID:           t.Id,
			UserID:       t.UserId,
			SpiceGradeID: t.SpiceGradeId,
			Type:         t.Type,
			Quantity:     t.Quantity,
			Price:        t.Price,
			TradeDate:    t.TradeDate,
			CreatedAt:    t.CreatedAt,
		}
	}

	var insights []*MerchantInsight
	if len(holdings) == 0 {
		insights = []*MerchantInsight{{
			Kind:     "IDLE",
			Title:    "Start your portfolio",
			Body:     "You have no open positions yet. Buy your first spice grade to see holdings and P&L here.",
			Severity: "info",
		}}
	} else {
		if statsResp.TradesInPeriod == 0 {
			insights = append(insights, &MerchantInsight{
				Kind:     "IDLE",
				Title:    "Quiet period",
				Body:     "No trades in the selected window. Your portfolio snapshot still reflects current holdings.",
				Severity: "info",
			})
		}
		var best, worst *MerchantHolding
		for _, h := range holdings {
			if h.TodayPrice <= 0 {
				continue
			}
			if best == nil || h.UnrealizedPnLPercent > best.UnrealizedPnLPercent {
				best = h
			}
			if worst == nil || h.UnrealizedPnLPercent < worst.UnrealizedPnLPercent {
				worst = h
			}
		}
		if best != nil && best.UnrealizedPnLPercent > 0 {
			id := best.SpiceGradeID
			insights = append(insights, &MerchantInsight{
				Kind:         "WINNER",
				Title:        "Top performer",
				Body:         fmt.Sprintf("%s - %s is up %.1f%% unrealized.", best.ProductName, best.GradeName, best.UnrealizedPnLPercent),
				SpiceGradeID: &id,
				Severity:     "success",
			})
		}
		if worst != nil && worst.UnrealizedPnLPercent < 0 && (best == nil || worst.SpiceGradeID != best.SpiceGradeID) {
			id := worst.SpiceGradeID
			insights = append(insights, &MerchantInsight{
				Kind:         "LOSER",
				Title:        "Under pressure",
				Body:         fmt.Sprintf("%s - %s is down %.1f%% vs cost basis.", worst.ProductName, worst.GradeName, math.Abs(worst.UnrealizedPnLPercent)),
				SpiceGradeID: &id,
				Severity:     "warning",
			})
		}
		for _, h := range holdings {
			if h.WeightPercent >= 70 {
				id := h.SpiceGradeID
				insights = append(insights, &MerchantInsight{
					Kind:         "CONCENTRATION",
					Title:        "Concentrated portfolio",
					Body:         fmt.Sprintf("%.0f%% of portfolio value is in %s - %s.", h.WeightPercent, h.ProductName, h.GradeName),
					SpiceGradeID: &id,
					Severity:     "warning",
				})
				break
			}
		}
		var periodRealized float64
		for _, row := range pnlResp.Rows {
			periodRealized += row.Amount
		}
		if periodRealized > 0 {
			insights = append(insights, &MerchantInsight{
				Kind:     "MILESTONE",
				Title:    "Profitable period",
				Body:     fmt.Sprintf("You locked in %.2f realized P&L in the selected window.", periodRealized),
				Severity: "success",
			})
		}
	}

	movers := make([]*PriceMover, len(snapshotsResp.Snapshots))
	for i, snap := range snapshotsResp.Snapshots {
		m := &PriceMover{
			SpiceGradeID:  snap.SpiceGradeId,
			ProductName:   snap.ProductName,
			GradeName:     snap.GradeName,
			TodayPrice:    snap.TodayPrice,
			PreviousPrice: snap.PreviousPrice,
			Direction:     "FLAT",
		}
		if snap.PreviousPrice > 0 && snap.TodayPrice > 0 {
			m.ChangePercent = ((snap.TodayPrice - snap.PreviousPrice) / snap.PreviousPrice) * 100
			if m.ChangePercent > 0.01 {
				m.Direction = "UP"
			} else if m.ChangePercent < -0.01 {
				m.Direction = "DOWN"
			}
		}
		movers[i] = m
	}

	return &MerchantDashboard{
		Summary:            summary,
		Holdings:           holdings,
		PortfolioMix:       portfolioMix,
		PnlTrend:           pnlTrend,
		ActivityTrend:      activityTrend,
		RecentTransactions: recentTransactions,
		Insights:           insights,
		Movers:             movers,
	}, nil
}

// ListTransactions is the resolver for the listTransactions field.
func (r *queryResolver) ListTransactions(ctx context.Context, skip *int, take *int) ([]*Transaction, error) {
	var skip32, take32 uint32
	if skip != nil {
		skip32 = uint32(*skip)
	}
	if take != nil {
		take32 = uint32(*take)
	}
	resp, err := r.server.marketClient.ListTransactions(ctx, &marketpb.ListTransactionsRequest{
		Skip: skip32,
		Take: take32,
	})
	if err != nil {
		return nil, err
	}
	transactions := make([]*Transaction, len(resp.Transactions))
	for i, t := range resp.Transactions {
		transactions[i] = &Transaction{
			ID:           t.Id,
			UserID:       t.UserId,
			SpiceGradeID: t.SpiceGradeId,
			Type:         t.Type,
			Quantity:     t.Quantity,
			Price:        t.Price,
			TradeDate:    t.TradeDate,
			CreatedAt:    t.CreatedAt,
		}
	}
	return transactions, nil
}
