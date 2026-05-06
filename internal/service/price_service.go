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

const (
	GroupByHours = "hour"
	GroupByDay   = "day"
	GroupByWeek  = "week"
)

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
func (s *PriceHistoryService) GetHistoryByInterval(id uint, limit int, interval, mode string) ([]model.PriceByInterval, error) {
	if limit <= 0 {
		limit = 10
	}
	validInterval := "hour"
	switch interval {
	case GroupByHours, GroupByDay, GroupByWeek:
		validInterval = interval
	}
	validMode := "last"
	switch mode {
	case "avg", "last":
		validMode = mode
	}
	item, err := s.repo.GetPriceByPeriod(id, limit, validInterval, validMode)
	if err != nil {
		return nil, err
	}
	return item, nil
}
