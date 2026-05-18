package api

import (
	"encoding/json"
	"errors"
	"investPort/internal/service"
	"investPort/internal/steam"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type API struct {
	ItemService  *service.ItemService
	PriceService *service.PriceHistoryService
	client       *steam.Client
	logger       *slog.Logger
}

func NewAPI(router chi.Router, itemService *service.ItemService, priceService *service.PriceHistoryService, client *steam.Client, logger *slog.Logger) *API {
	api := &API{
		ItemService:  itemService,
		PriceService: priceService,
		client:       client,
		logger:       logger,
	}
	router.Get("/items", api.getItems())
	router.Post("/items", api.addNewItemByURL())
	router.Get("/items/{id}", api.getItemById())
	router.Get("/items/{id}/history", api.getHistoryById())
	return api
}
func (api *API) getItems() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		logger := api.logger.With(
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)
		query, err := api.parsePaginationQuery(r)
		if err != nil {
			logger.Warn("invalid pagination query",
				slog.String("error", err.Error()))
			api.writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		items, err := api.ItemService.GetItemsPaginated(query.Offset, query.Limit)
		if err != nil {
			logger.Error("failed to get items from service",
				slog.Int("offset", query.Offset),
				slog.Int("limit", query.Limit),
				slog.String("error", err.Error()))
			api.writeError(w, http.StatusInternalServerError, "failed to get items")
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

		if err = api.writeJSON(w, http.StatusOK, itemsStruct); err != nil {
			logger.Error("failed to encode items response to JSON",
				slog.Int("items_count", len(items)),
				slog.String("error", err.Error()))
			return
		}
	}
}
func (api *API) getItemById() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		logger := api.logger.With(
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		idStr := chi.URLParam(r, "id")
		if idStr == "" {
			logger.Warn("empty id url param")
			api.writeError(w, http.StatusBadRequest, "empty id url param")
			return
		}

		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			logger.Warn("invalid id format",
				slog.String("id", idStr),
				slog.String("error", err.Error()))
			api.writeError(w, http.StatusBadRequest, "invalid id format")
			return
		}

		item, err := api.ItemService.GetByID(uint(id))
		if err != nil {
			if errors.Is(err, service.ErrItemNotFound) {
				logger.Info("item not found",
					slog.Uint64("item_id", id))
				api.writeError(w, http.StatusNotFound, "item not found")
				return
			}
			logger.Error("failed to get item by id",
				slog.Uint64("item_id", id),
				slog.String("error", err.Error()))
			api.writeError(w, http.StatusInternalServerError, "failed to get item by id")
			return
		}

		if err = api.writeJSON(w, http.StatusOK, itemResponse{
			ID:   item.ID,
			Name: item.Name,
			Url:  item.URL,
		}); err != nil {
			logger.Error("failed to encode item response",
				slog.Uint64("item_id", id),
				slog.String("error", err.Error()))
		}
	}
}

func (api *API) getHistoryById() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		logger := api.logger.With(
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)
		idStr := chi.URLParam(r, "id")
		if idStr == "" {
			logger.Warn("empty id url param")
			api.writeError(w, http.StatusBadRequest, "missing id")
			return
		}

		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			logger.Warn("invalid id format",
				slog.String("id", idStr),
				slog.String("error", err.Error()))
			api.writeError(w, http.StatusBadRequest, "invalid id param")
			return
		}

		query, err := api.parseHistoryQuery(r)

		if err != nil {
			logger.Warn("invalid history query",
				slog.String("error", err.Error()))
			api.writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		prices, err := api.PriceService.GetHistoryWithPagination(uint(id), query.Limit, query.Offset, query.Interval, query.Mode)
		if err != nil {
			logger.Error("failed to get price history by item ID",
				slog.Uint64("item_id", id),
				slog.String("interval", query.Interval),
				slog.String("mode", query.Mode),
				slog.Int("limit", query.Limit),
				slog.Int("offset", query.Offset),
				slog.String("error", err.Error()))
			api.writeError(w, http.StatusInternalServerError, "failed to get price item history")
			return
		}

		response := SteamPricesResponse{
			ItemId:   uint(id),
			Interval: query.Interval,
			Mode:     query.Mode,
			Limit:    query.Limit,
			Offset:   query.Offset,
			Prices:   make([]steamPriceResponse, 0, len(prices)),
		}

		for _, p := range prices {
			response.Prices = append(response.Prices, steamPriceResponse{
				Price: p.Price,
				Time:  p.Bucket.Format(time.RFC3339),
			})
		}

		if err = api.writeJSON(w, http.StatusOK, response); err != nil {
			logger.Error("failed to encode price history response",
				slog.Uint64("item_id", id),
				slog.String("error", err.Error()))
		}
	}
}

func (api *API) addNewItemByURL() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request UrlRequest
		defer func() {
			_ = r.Body.Close()
		}()

		logger := api.logger.With(
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()

		err := decoder.Decode(&request)

		if err != nil {
			logger.Error("could not read JSON value",
				slog.String("error", err.Error()))
			api.writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		if request.Url == "" {
			logger.Warn("url field is empty")
			api.writeError(w, http.StatusBadRequest, "url field is empty")
			return
		}

		parsedItem, err := api.client.ParseItemURL(request.Url)
		if err != nil {
			logger.Warn("failed to parse item URL",
				slog.String("url", request.Url),
				slog.String("error", err.Error()))
			api.writeError(w, http.StatusBadRequest, "failed to parse url")
			return
		}

		item, err := api.ItemService.GetOrCreateItem(parsedItem)
		if errors.Is(err, service.ErrItemAlreadyExist) {
			logger.Info("item already exists",
				slog.Uint64("item_id", uint64(item.ID)))
			_ = api.writeJSON(w, http.StatusConflict, itemResponseWithError{
				Error: "item already exists",
				ID:    item.ID,
				Name:  item.Name,
				Url:   item.URL,
			})
			return
		}
		if err != nil {
			logger.Error("failed to get or create item",
				slog.String("url", request.Url),
				slog.String("error", err.Error()))
			api.writeError(w, http.StatusInternalServerError, "failed to get or create item")
			return
		}

		logger.Info("item created successfully",
			slog.Uint64("item_id", uint64(item.ID)),
			slog.String("name", item.Name))

		if err = api.writeJSON(w, http.StatusCreated, itemResponse{
			ID:   item.ID,
			Name: item.Name,
			Url:  item.URL,
		}); err != nil {
			logger.Error("could not encode response",
				slog.String("error", err.Error()))
		}
	}
}

func (api *API) parseHistoryQuery(r *http.Request) (*HistoryQuery, error) {

	logger := api.logger.With(
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	query := &HistoryQuery{}

	interval := r.URL.Query().Get("interval")
	mode := r.URL.Query().Get("mode")

	const (
		GroupByHours = "hour"
		GroupByDay   = "day"
		GroupByWeek  = "week"
		ModeAVG      = "avg"
		ModeLast     = "last"
	)

	validInterval := GroupByHours
	switch interval {
	case GroupByHours, GroupByDay, GroupByWeek:
		validInterval = interval
		logger.Debug("using custom interval",
			slog.String("interval", validInterval))
	default:
		logger.Debug("invalid interval, using default",
			slog.String("provided", interval),
			slog.String("default", validInterval))
	}
	query.Interval = validInterval

	validMode := ModeLast
	switch mode {
	case ModeAVG, ModeLast:
		validMode = mode
		logger.Debug("using custom mode",
			slog.String("mode", validMode))
	default:
		logger.Debug("invalid mode, using default",
			slog.String("provided", mode),
			slog.String("default", validMode))
	}
	query.Mode = validMode

	paginationQuery, err := api.parsePaginationQuery(r)
	if err != nil {
		return nil, err
	}
	query.Offset = paginationQuery.Offset
	query.Limit = paginationQuery.Limit

	return query, nil
}

func (api *API) parsePaginationQuery(r *http.Request) (*PaginationQuery, error) {
	logger := api.logger.With(
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	const (
		defaultLimit  = 20
		defaultOffset = 0
	)

	offsetStr := r.URL.Query().Get("offset")
	limitStr := r.URL.Query().Get("limit")

	query := &PaginationQuery{}

	var offset, limit int
	var err error

	if offsetStr == "" {
		offset = defaultOffset
		logger.Debug("using default offset")
		slog.Int("offset", defaultOffset)
	} else {
		offset, err = strconv.Atoi(offsetStr)
		if err != nil {
			logger.Warn("failed to parse offset",
				slog.String("offset", offsetStr),
				slog.String("error", err.Error()))
			return nil, errors.New("invalid offset")
		}
		if offset < 0 {
			limit = defaultOffset
			logger.Debug("using default offset",
				slog.Int("offset", defaultOffset))
		}
	}
	query.Offset = offset

	if limitStr == "" {
		limit = defaultLimit
		logger.Debug("using default limit",
			slog.Int("limit", defaultLimit))
	} else {
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			logger.Warn("failed to parse limit",
				slog.String("limit", limitStr),
				slog.String("error", err.Error()))
			return nil, errors.New("invalid limit")
		}
		if limit <= 0 || limit > 100 {
			limit = defaultLimit
			logger.Debug("using default limit",
				slog.Int("limit", defaultLimit))
		}
	}
	query.Limit = limit
	return query, nil
}
func (api *API) writeJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}
func (api *API) writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}
