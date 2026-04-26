package api

import (
	"encoding/json"
	"errors"
	"investPort/internal/service"
	"log"
	"net/http"
	"strconv"

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
		itemsStruct := allItemsResponse{
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
			if idStr == "" {
				log.Println("empty id url param")
				http.Error(w, "missing id", http.StatusBadRequest)
				return
			}
		}
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			log.Printf("invalid id format: %s\n", idStr)
			http.Error(w, "invalid id param", http.StatusBadRequest)
			return
		}
		prices, err := api.PriceService.GetHistoryById(uint(id))
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Printf("price not found %d: %v\n", id, err)
				http.Error(w, "price not found", http.StatusNotFound)
				return
			}
			log.Printf("failed to get price by id %d: %v\n", id, err)
			http.Error(w, "failed to get price", http.StatusInternalServerError)
			return
		}
		pricesHistory := PricesResponse{
			Prices: make([]priceResponse, len(prices)),
		}
		for i, price := range prices {
			pricesHistory.Prices[i] = priceResponse{
				ItemId: price.ItemID, Price: price.Price, InspectionTime: price.InspectionTime,
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
