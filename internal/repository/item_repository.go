package repository

import (
	"fmt"
	"investPort/internal/model"
	"net/url"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

type ItemRepository struct{ DB *gorm.DB }

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
func (repo *ItemRepository) GetAll() ([]model.Item, error) {
	var items []model.Item

	if err := repo.DB.Find(&items).Error; err != nil {
		return nil, err
	}

	return items, nil
}
func (repo *ItemRepository) ParseMarketHashName(urlString string) (string, int, error) {
	parsed, err := url.Parse(urlString)
	if err != nil {
		return "", 0, fmt.Errorf("failed to parse URL: %w", err)
	}

	if parsed.Host != "steamcommunity.com" && !strings.HasSuffix(parsed.Host, ".steamcommunity.com") {
		return "", 0, fmt.Errorf("unsupported host: %s", parsed.Host)
	}

	parts := strings.Split(parsed.Path, "/")
	if len(parts) < 5 {
		return "", 0, fmt.Errorf("invalid URL path: too short (%s)", parsed.Path)
	}

	if len(parts) < 4 || parts[1] != "market" || parts[2] != "listings" {
		return "", 0, fmt.Errorf("invalid Steam market URL format: expected /market/listings/{app_id}/")
	}

	appIDStr := parts[3]
	appID, err := strconv.Atoi(appIDStr)
	if err != nil {
		return "", 0, fmt.Errorf("invalid appID '%s': %w", appIDStr, err)
	}
	if appID <= 0 {
		return "", 0, fmt.Errorf("invalid appID: must be positive")
	}
	l, _ := url.QueryUnescape(parts[4])
	fmt.Println(parts, "========", url.QueryEscape(parts[4]), l)

	hash := parts[4]
	marketHashName, err := url.QueryUnescape(hash)
	if err != nil {
		return "", 0, fmt.Errorf("failed to unescape market hash name: %w", err)
	}
	if marketHashName == "" {
		return "", 0, fmt.Errorf("empty market hash name")
	}

	return marketHashName, appID, nil
}
