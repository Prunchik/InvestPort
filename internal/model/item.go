package model

type Item struct {
	ID       uint   `gorm:"primaryKey"`
	Name     string `json:"name"`
	URL      string `json:"url" gorm:"uniqueIndex"`
	AppID    int    `json:"app_id"`
	HashName string `json:"hash_name"`
}

func NewItem(name, hashName, url string, appID int) *Item {
	return &Item{
		Name:     name,
		URL:      url,
		AppID:    appID,
		HashName: hashName,
	}
}
