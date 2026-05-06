package api

type AllItemsResponse struct {
	Items []itemResponse `json:"items"`
}
type itemResponse struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Url  string `json:"url"`
}

type PricesResponse struct {
	ItemId uint            `json:"item_id"`
	Prices []priceResponse `json:"prices"`
}
type priceResponse struct {
	Price float64 `json:"price"`
	Time  string  `json:"time"`
}
