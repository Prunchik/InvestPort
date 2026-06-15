package model

type Item struct {
	ID           uint   `gorm:"primaryKey" gorm:"uniqueIndex"`
	Name         string `json:"name"`
	URL          string `json:"url" gorm:"uniqueIndex"`
	AppID        int    `json:"app_id"`
	HashName     string `json:"hash_name" gorm:"uniqueIndex"`
	ImageURL     string `json:"image_url"`
	WearCategory *int   `json:"wear_category"`
}

func NewItem(name, hashName, url string, appID int, wearCategory *int, imageURL string) *Item {
	return &Item{
		Name:         name,
		URL:          url,
		AppID:        appID,
		HashName:     hashName,
		WearCategory: wearCategory,
		ImageURL:     imageURL,
	}
}
