package model

import "time"

type PriceHistory struct {
	ID             uint      `gorm:"primaryKey"`
	ItemID         uint      `gorm:"index;not null"`
	Price          float64   `json:"price"`
	InspectionTime time.Time `json:"inspectionTime"`
}

func NewPriceHistory(itemID uint, price float64) *PriceHistory {
	return &PriceHistory{
		ItemID:         itemID,
		Price:          price,
		InspectionTime: time.Now(),
	}
}
