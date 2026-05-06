package api

import (
	"encoding/json"
	"errors"
	"investPort/internal/service"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type Api struct {
	ItemService  *service.ItemService
	PriceService *service.PriceHistoryService
}

func NewApi(router chi.Router, itemService *service.ItemService, priceService *service.PriceHistoryService) *Api {
	api := &Api{ItemService: itemService, PriceService: priceService}
	router.Get("/items", api.getAllItems())
	router.Get("/items/{id}", api.getItemById())
	router.Get("/items/{id}/history", api.getHistoryById())
	return api
}
func (api *Api) getAllItems() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		items, err := api.ItemService.GetAllItems()
		if err != nil {
			log.Printf("error fetching items: %v", err)
			http.Error(w, "failed to get items", http.StatusInternalServerError)
			return
		}
		itemsStruct := AllItemsResponse{
			Items: make([]itemResponse, len(items)),
		}
		for i, item := range items {
			itemsStruct.Items[i] = itemResponse{
				ID:   item.ID,
				Name: item.Name,
				Url:  item.URL,
			}
		}
		if err = api.encoder(w, itemsStruct); err != nil {
			log.Printf("error encoding items to json: %v", err)
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
			return
		}

	}
}
func (api *Api) getItemById() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		if idStr == "" {
			log.Println("empty id url param")
			http.Error(w, "missing id", http.StatusBadRequest)
			return
		}
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			log.Printf("invalid id format: %s\n", idStr)
			http.Error(w, "invalid id param", http.StatusBadRequest)
			return
		}
		item, err := api.ItemService.GetByID(uint(id))
		if err != nil {

			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Printf("item not found %d: %v\n", id, err)
				http.Error(w, "item not found", http.StatusNotFound)
				return
			}
			log.Printf("failed to get item by id %d: %v\n", id, err)
			http.Error(w, "failed to get item", http.StatusInternalServerError)
			return
		}
		if err = api.encoder(w, itemResponse{ID: item.ID, Name: item.Name, Url: item.URL}); err != nil {
			log.Printf("error encoding item to json: %v", err)
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}
func (api *Api) getHistoryById() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		idStr := chi.URLParam(r, "id")
		if idStr == "" {
			log.Println("empty id url param")
			http.Error(w, "missing id", http.StatusBadRequest)
			return
		}
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			log.Printf("invalid id format: %s\n", idStr)
			http.Error(w, "invalid id param", http.StatusBadRequest)
			return
		}

		interval := r.URL.Query().Get("interval")
		mode := r.URL.Query().Get("mode")
		limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
		if err != nil {
			log.Printf("failed to parse limit %d : %v", limit, err)
			http.Error(w, "failed to get price", http.StatusInternalServerError)
			return
		}

		prices, err := api.PriceService.GetHistoryByInterval(uint(id), limit, interval, mode)
		if err != nil {
			log.Printf("failed to get price by id %d: %v\n", id, err)
			http.Error(w, "failed to get price", http.StatusInternalServerError)
			return
		}
		if len(prices) > 0 {
			emptyResponse := PricesResponse{
				ItemId: uint(id),
				Prices: []priceResponse{},
			}
			err = api.encoder(w, emptyResponse)
			if err != nil {
				log.Printf("error encoding price to json: %v", err)
				http.Error(w, "failed to encode response", http.StatusInternalServerError)
				return
			}
		}
		pricesHistory := PricesResponse{
			ItemId: uint(id),
			Prices: make([]priceResponse, len(prices)),
		}
		for i, price := range prices {
			pricesHistory.Prices[i] = priceResponse{
				Price: price.Price,
				Time:  price.Bucket.Format(time.RFC3339),
			}
		}
		err = api.encoder(w, pricesHistory)
		if err != nil {
			log.Printf("error encoding price to json: %v", err)
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
			return
		}

	}
}

func (api *Api) encoder(w http.ResponseWriter, v any) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(v)
}
