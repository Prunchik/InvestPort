package api

import "time"

type allItemsResponse struct {
	Items []itemResponse `json:"items"`
}
type itemResponse struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Url  string `json:"url"`
}

type PricesResponse struct {
	Prices []priceResponse `json:"prices"`
}
type priceResponse struct {
	ItemId         uint      `json:"item_id"`
	Price          float64   `json:"price"`
	InspectionTime time.Time `json:"inspection_time"`
}
