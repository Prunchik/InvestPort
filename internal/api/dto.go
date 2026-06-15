package api

type AllItemsResponse struct {
	Items []itemResponse `json:"items"`
}
type itemResponse struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}
type itemResponseWithError struct {
	Error string `json:"error"`
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Url   string `json:"url"`
}

type SteamPricesResponse struct {
	ItemId   uint                 `json:"item_id"`
	Interval string               `json:"interval"`
	Mode     string               `json:"mode"`
	Limit    int                  `json:"limit"`
	Offset   int                  `json:"offset"`
	Prices   []steamPriceResponse `json:"prices"`
}
type steamPriceResponse struct {
	Price float64 `json:"price"`
	Time  string  `json:"time"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type UrlRequest struct {
	Url string `json:"url"`
}

type PaginationQuery struct {
	Offset int    `json:"offset"`
	Limit  int    `json:"limit"`
	Query  string `json:"search"`
}
type HistoryQuery struct {
	PaginationQuery
	Interval string `json:"interval"`
	Mode     string `json:"mode"`
}
