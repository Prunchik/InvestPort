package service

import (
	"errors"
	"math"

	"investPort/internal/model"
	"investPort/internal/repository"
	"investPort/internal/steam"

	"gorm.io/gorm"
)

type PriceHistoryService struct {
	repo   *repository.PriceHistoryRepository
	client *steam.Client
}

func NewPriceHistoryService(r *repository.PriceHistoryRepository, c *steam.Client) *PriceHistoryService {
	return &PriceHistoryService{
		repo:   r,
		client: c,
	}
}

func (s *PriceHistoryService) UpdatePriceForItem(item *model.Item) error {
	price, err := s.client.GetPrice(item.HashName, item.AppID)
	if err != nil {
		return err
	}

	_, err = s.ProcessPrice(item.ID, price)
	return err
}

func (s *PriceHistoryService) ProcessPrice(itemID uint, newPrice float64) (*model.PriceHistory, error) {
	const epsilon = 1e-6
	latest, err := s.repo.GetLatestPrice(itemID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		newRecord := model.NewPriceHistory(itemID, newPrice)
		if err = s.repo.Create(newRecord); err != nil {
			return nil, err
		}
		return newRecord, nil
	}
	if err != nil {
		return nil, err
	}

	if math.Abs(newPrice-latest.Price) < epsilon {
		return latest, nil
	}
	price := model.NewPriceHistory(itemID, newPrice)
	if err = s.repo.Create(price); err != nil {
		return nil, err
	}
	return price, nil
}
func (s *PriceHistoryService) GetHistoryById(id uint) ([]model.PriceHistory, error) {
	item, err := s.repo.GetHistory(id)
	if err != nil {
		return nil, err
	}
	return item, nil
}
