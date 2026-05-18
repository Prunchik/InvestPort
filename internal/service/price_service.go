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
	priceRepo *repository.PriceHistoryRepository
	itemRepo  *repository.ItemRepository
	client    *steam.Client
}

func NewPriceHistoryService(r *repository.PriceHistoryRepository, i *repository.ItemRepository, c *steam.Client) *PriceHistoryService {
	return &PriceHistoryService{
		priceRepo: r,
		itemRepo:  i,
		client:    c,
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
	latest, err := s.priceRepo.GetLatestPrice(itemID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		newRecord := model.NewPriceHistory(itemID, newPrice)
		if err = s.priceRepo.Create(newRecord); err != nil {
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
	if err = s.priceRepo.Create(price); err != nil {
		return nil, err
	}
	return price, nil
}
func (s *PriceHistoryService) GetHistoryWithPagination(id uint, limit, offset int, interval, mode string) ([]model.PriceByInterval, error) {
	item, err := s.priceRepo.GetPriceByPeriod(id, limit, offset, interval, mode)
	if err != nil {
		return nil, err
	}
	return item, nil
}
