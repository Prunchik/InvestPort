package repository

import (
	"investPort/internal/model"

	"gorm.io/gorm"
)

type ItemRepository struct{ DB *gorm.DB }
type ItemFilter struct {
	Query  string
	Offset int
	Limit  int
}

func NewItemRepo(db *gorm.DB) *ItemRepository { return &ItemRepository{db} }
func (repo *ItemRepository) Create(item *model.Item) error {
	result := repo.DB.Create(item)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
func (repo *ItemRepository) GetById(id uint) (*model.Item, error) {
	var item model.Item
	result := repo.DB.Where("id = ?", id).First(&item)
	if result.Error != nil {
		return nil, result.Error
	}
	return &item, nil
}
func (repo *ItemRepository) GetByName(name string) (*model.Item, error) {
	var item model.Item
	result := repo.DB.Where("name = ?", name).First(&item)
	if result.Error != nil {
		return nil, result.Error
	}
	return &item, nil
}
func (repo *ItemRepository) GetByURL(url string) (*model.Item, error) {
	var item model.Item
	result := repo.DB.Where("url = ?", url).First(&item)
	if result.Error != nil {
		return nil, result.Error
	}
	return &item, nil
}

func (repo *ItemRepository) ListAll() ([]model.Item, error) {
	var items []model.Item

	if err := repo.DB.Find(&items).Error; err != nil {
		return nil, err
	}

	return items, nil
}

func (repo *ItemRepository) List(filter *ItemFilter) ([]model.Item, error) {

	query := repo.DB.Model(&model.Item{})

	if filter.Query != "" {
		query = query.Where("name ILIKE ?", "%"+filter.Query+"%")
	}

	var items []model.Item

	err := query.
		Offset(filter.Offset).
		Limit(filter.Limit).
		Find(&items).
		Error

	return items, err
}
