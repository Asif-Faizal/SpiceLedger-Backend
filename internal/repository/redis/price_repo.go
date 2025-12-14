package redis

import (
    "context"
    "encoding/json"
    "fmt"

    "github.com/google/uuid"
    "github.com/redis/go-redis/v9"
    "github.com/saravanan/spice_backend/internal/domain"
)

type priceRepository struct {
	rdb *redis.Client
}

func NewPriceRepository(rdb *redis.Client) domain.PriceRepository {
	return &priceRepository{rdb: rdb}
}

func (r *priceRepository) getKey(date string, productID uuid.UUID, gradeID uuid.UUID) string {
    return fmt.Sprintf("price:%s:%s:%s", date, productID.String(), gradeID.String())
}

func (r *priceRepository) SetPrice(ctx context.Context, date string, productID uuid.UUID, gradeID uuid.UUID, price float64) error {
    key := r.getKey(date, productID, gradeID)
    priceData := domain.DailyPrice{
        Date:       date,
        ProductID:  productID,
        GradeID:    gradeID,
        PricePerKg: price,
    }

	data, err := json.Marshal(priceData)
	if err != nil {
		return err
	}

	return r.rdb.Set(ctx, key, data, 0).Err()
}

func (r *priceRepository) GetPrice(ctx context.Context, date string, productID uuid.UUID, gradeID uuid.UUID) (float64, error) {
    key := r.getKey(date, productID, gradeID)
	val, err := r.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return 0, domain.ErrPriceNotFound
	}
	if err != nil {
		return 0, err
	}

	var priceData domain.DailyPrice
	if err := json.Unmarshal([]byte(val), &priceData); err != nil {
		return 0, err
	}

	return priceData.PricePerKg, nil
}

func (r *priceRepository) GetPricesForDate(ctx context.Context, date string) ([]domain.DailyPrice, error) {
    pattern := fmt.Sprintf("price:%s:*", date)

	var prices []domain.DailyPrice

	// Scan for keys matching the pattern
	iter := r.rdb.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		val, err := r.rdb.Get(ctx, key).Result()
		if err != nil {
			continue // Skip errs for now
		}

		var priceData domain.DailyPrice
		if err := json.Unmarshal([]byte(val), &priceData); err == nil {
			prices = append(prices, priceData)
		}
	}

	if err := iter.Err(); err != nil {
		return nil, err
	}

	return prices, nil
}
