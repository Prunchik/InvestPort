package service

import (
	"errors"
	"investPort/internal/model"
	"investPort/internal/repository"
	"net/url"

	"gorm.io/gorm"
)

type ItemService struct {
	repo *repository.ItemRepository
}

func NewItemService(repo *repository.ItemRepository) *ItemService {
	return &ItemService{repo: repo}
}

func (s *ItemService) GetOrCreateByURL(URL string) (*model.Item, error) {
	item, err := s.repo.GetByURL(URL)
	if err == nil {
		return item, nil
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		name, appID, err := s.repo.ParseMarketHashName(URL)
		if err != nil {
			return nil, err
		}
		hash := url.QueryEscape(name)
		newItem := model.NewItem(name, hash, URL, appID)

		if err = s.repo.Create(newItem); err != nil {
			return nil, err
		}

		return newItem, nil
	}

	return nil, err
}

func (s *ItemService) GetByID(id uint) (*model.Item, error) {
	item, err := s.repo.GetById(id)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (s *ItemService) GetItemsForProcessing() ([]model.Item, error) {
	items, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}

	return items, nil
}
