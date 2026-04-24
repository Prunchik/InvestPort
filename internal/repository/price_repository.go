package repository

import (
	"investPort/internal/model"

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
