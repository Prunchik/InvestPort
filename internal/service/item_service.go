package service

import (
	"errors"
	"investPort/internal/model"
	"investPort/internal/repository"
	"investPort/internal/steam"

	"gorm.io/gorm"
)

type ItemService struct {
	repo *repository.ItemRepository
}

func NewItemService(repo *repository.ItemRepository) *ItemService {
	return &ItemService{
		repo: repo,
	}
}

func (s *ItemService) GetOrCreateItem(parsedItem *steam.ParsedItem) (*model.Item, error) {
	item, err := s.repo.GetByURL(parsedItem.Url)

	if err == nil {
		return item, ErrItemAlreadyExist
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		newItem := model.NewItem(parsedItem.Name, parsedItem.HashName, parsedItem.Url, parsedItem.AppID)

		if err = s.repo.Create(newItem); err != nil {
			return nil, err
		}

		return newItem, nil
	}

	return nil, err
}

func (s *ItemService) GetByID(id uint) (*model.Item, error) {
	item, err := s.repo.GetById(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrItemNotFound
	}

	if err != nil {
		return nil, err
	}

	return item, nil
}

func (s *ItemService) GetItemsPaginated(offset, limit int) ([]model.Item, error) {
	if offset < 0 {
		offset = 0
	}

	if limit <= 0 {
		limit = 10
	}

	items, err := s.repo.GetItemsPaginated(offset, limit)
	if err != nil {
		return nil, err
	}

	return items, nil
}
func (s *ItemService) GetItemsForProcessing() ([]model.Item, error) {

	items, err := s.repo.GetItemsForProcessing()

	if err != nil {
		return nil, err
	}

	return items, nil
}
