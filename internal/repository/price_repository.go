package repository

import (
	"fmt"
	"investPort/internal/model"
	"slices"

	"gorm.io/gorm"
)

type PriceHistoryRepository struct {
	DB *gorm.DB
}

func NewPriceHistoryRepository(db *gorm.DB) *PriceHistoryRepository {
	return &PriceHistoryRepository{DB: db}
}

func (r *PriceHistoryRepository) Create(price *model.PriceHistory) error {
	return r.DB.Create(price).Error
}

func (r *PriceHistoryRepository) GetLatestPrice(itemID uint) (*model.PriceHistory, error) {
	var price model.PriceHistory

	if err := r.DB.
		Where("item_id = ?", itemID).
		Order("inspection_time DESC").
		First(&price).Error; err != nil {
		return nil, err
	}

	return &price, nil
}

func (r *PriceHistoryRepository) GetHistory(itemID uint) ([]model.PriceHistory, error) {
	var prices []model.PriceHistory

	if err := r.DB.
		Where("item_id = ?", itemID).
		Order("inspection_time ASC").
		Find(&prices).Error; err != nil {
		return nil, err
	}

	return prices, nil
}

func (r *PriceHistoryRepository) GetPriceByPeriod(itemID uint, limit int, interval, mode string) ([]model.PriceByInterval, error) {
	var prices []model.PriceByInterval
	
	var query string

	switch mode {
	case "last":
		query = `
		SELECT bucket, price
		FROM (
			SELECT DISTINCT ON (bucket)
				bucket,
				price
			FROM (
				SELECT
					date_trunc($1, inspection_time) AS bucket,
					price,
					inspection_time
				FROM price_histories
				WHERE item_id = $2
			) t1
			ORDER BY bucket, inspection_time DESC
		) t2
		ORDER BY bucket DESC
		LIMIT $3
		`

	case "avg":
		query = `
		SELECT bucket, price
		FROM (
			SELECT
				date_trunc($1, inspection_time) AS bucket,
				AVG(price) AS price
			FROM price_histories
			WHERE item_id = $2
			GROUP BY bucket
			ORDER BY bucket DESC
			LIMIT $3
		) t
		`

	default:
		return nil, fmt.Errorf("invalid mode: %s", mode)
	}

	err := r.DB.
		Raw(query, interval, itemID, limit).
		Scan(&prices).Error

	if err != nil {
		return nil, err
	}

	slices.Reverse(prices)

	return prices, nil
}
