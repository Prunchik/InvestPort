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
	item, err := s.repo.GetByURL(parsedItem.URL)

	if err == nil {
		return item, ErrItemAlreadyExist
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		newItem := model.NewItem(
			parsedItem.Name,
			parsedItem.HashName,
			parsedItem.URL,
			parsedItem.AppID,
			parsedItem.WearCategory,
			parsedItem.ImageURL,
		)

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

func (s *ItemService) ListItems(filter *repository.ItemFilter) ([]model.Item, error) {
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	if filter.Limit <= 0 {
		filter.Limit = 20
	}

	items, err := s.repo.List(filter)

	if err != nil {
		return nil, err
	}

	return items, nil
}
func (s *ItemService) ListAll() ([]model.Item, error) {

	items, err := s.repo.ListAll()

	if err != nil {
		return nil, err
	}

	return items, nil
}
