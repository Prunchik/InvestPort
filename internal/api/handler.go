package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"investPort/internal/cache"
	"investPort/internal/repository"
	"investPort/internal/service"
	"investPort/internal/steam"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const cacheTTLItem = 5 * time.Minute

type API struct {
	ItemService  *service.ItemService
	PriceService *service.PriceHistoryService
	client       *steam.Client
	logger       *slog.Logger
	cache        cache.Cache
}

func NewAPI(router chi.Router,
	itemService *service.ItemService,
	priceService *service.PriceHistoryService,
	client *steam.Client,
	logger *slog.Logger,
	cache cache.Cache,
	staticFS http.Handler, getIndexHTML func() ([]byte, error)) *API {
	api := &API{
		ItemService:  itemService,
		PriceService: priceService,
		client:       client,
		logger:       logger,
		cache:        cache,
	}
	router.Get("/api/items", api.getItems())
	router.Post("/api/items", api.addNewItemByURL())
	router.Get("/api/items/{id}", api.getItemById())
	router.Get("/api/items/{id}/history", api.getHistoryById())

	RegisterSwaggerRoutes(router)

	router.Handle("/static/*", http.StripPrefix("/static/", staticFS))
	router.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		indexHTML, err := getIndexHTML()
		if err != nil {
			http.Error(w, "frontend not built", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(indexHTML)
	})

	return api
}

// @Summary		List tracked items
// @Description	Get paginated list of all tracked Steam Market items
// @Param		offset	query	int	false	"Pagination offset"	default(0)
// @Param		limit	query	int	false	"Items per page (1-100)"	default(20)
// @Success		200	{object}	AllItemsResponse
// @Failure		400	{object}	ErrorResponse
// @Router		/api/items [get]
func (api *API) getItems() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		logger := api.logger.With(
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)
		query, err := parsePaginationQuery(r)
		if err != nil {
			logger.Warn("failed parse pagination query",
				slog.String("error", err.Error()))
			api.writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		items, err := api.ItemService.ListItems(&repository.ItemFilter{
			Query:  query.Query,
			Offset: query.Offset,
			Limit:  query.Limit,
		})
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
				URL:  item.URL,
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

// @Summary		Get item by ID
// @Description	Get a single tracked item by its database ID
// @Param		id	path	int	true	"Item ID"
// @Success		200	{object}	itemResponse
// @Failure		400	{object}	ErrorResponse
// @Failure		404	{object}	ErrorResponse
// @Router		/api/items/{id} [get]
func (api *API) getItemById() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()

		logger := api.logger.With(
			slog.String("request_id", middleware.GetReqID(ctx)),
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

		cacheKey := fmt.Sprintf("item:%d", id)

		if cached, err := api.cache.GetItem(ctx, cacheKey); err == nil {

			w.Header().Set("Content-Type", "application/json")
			w.Write(cached)
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

		resp, err := json.Marshal(itemResponse{ID: item.ID, Name: item.Name, URL: item.URL})
		if err != nil {
			slog.Error("error", err.Error())
			api.writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed marshal item %v", err))
			return
		}

		err = api.cache.SetItem(ctx, cacheKey, resp, cacheTTLItem)
		if err != nil {
			logger.Warn("failed to set item to cache",
				slog.String("key", cacheKey),
				slog.String("error", err.Error()))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(resp)
	}
}

// @Summary		Get price history
// @Description	Get bucketed price history for an item
// @Param		id		path	int		true	"Item ID"
// @Param		interval	query	string	false	"Bucket size: hour, day, week"	default(hour)
// @Param		mode		query	string	false	"Aggregation: last, avg"		default(last)
// @Param		offset		query	int		false	"Pagination offset"			default(0)
// @Param		limit		query	int		false	"Items per page (1-100)"		default(20)
// @Success		200	{object}	SteamPricesResponse
// @Failure		400	{object}	ErrorResponse
// @Router		/api/items/{id}/history [get]
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

		query, err := parseHistoryQuery(r)

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

// @Summary		Add item by Steam URL
// @Description	Add a new Steam Market item to track by its URL
// @Accept		json
// @Produce		json
// @Param		body	body	UrlRequest	true	"Steam Market URL"
// @Success		201	{object}	itemResponse
// @Failure		400	{object}	ErrorResponse
// @Failure		409	{object}	itemResponseWithError
// @Router		/api/items [post]
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

		if err := decoder.Decode(&request); err != nil {
			logger.Error("could not read JSON value",
				slog.String("error", err.Error()))
			api.writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}

		if decoder.More() {
			api.writeError(w, http.StatusBadRequest, "only one JSON object allowed")
			return
		}

		_, err := decoder.Token()
		if err != io.EOF {
			api.writeError(w, http.StatusBadRequest, "trailing data after JSON")
		}

		if request.Url == "" {
			logger.Warn("url field is empty")
			api.writeError(w, http.StatusBadRequest, "url field is empty")
			return
		}
		parsedUrl, err := url.Parse(request.Url)
		if err != nil {
			logger.Warn("invalid url format")
			api.writeError(w, http.StatusBadRequest, "invalid url format")
			return
		}
		if parsedUrl.Scheme != "http" && parsedUrl.Scheme != "https" {
			logger.Warn("url must use http or https")
			api.writeError(w, http.StatusBadRequest, "url must use http or https")
			return
		}

		parsedItem, err := api.client.ResolveSteamMarketItem(request.Url)
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
			URL:  item.URL,
		}); err != nil {
			logger.Error("could not encode response",
				slog.String("error", err.Error()))
		}
	}
}

func parseHistoryQuery(r *http.Request) (*HistoryQuery, error) {
	query := &HistoryQuery{}

	interval := r.URL.Query().Get("interval")
	mode := r.URL.Query().Get("mode")

	const (
		GroupByHour = "hour"
		GroupByDay  = "day"
		GroupByWeek = "week"
		ModeAVG     = "avg"
		ModeLast    = "last"
	)

	validInterval := GroupByHour
	switch interval {
	case GroupByHour, GroupByDay, GroupByWeek:
		validInterval = interval
	}
	query.Interval = validInterval

	validMode := ModeLast
	switch mode {
	case ModeAVG, ModeLast:
		validMode = mode
	}
	query.Mode = validMode

	paginationQuery, err := parsePaginationQuery(r)
	if err != nil {
		return nil, err
	}
	query.Offset = paginationQuery.Offset
	query.Limit = paginationQuery.Limit

	return query, nil
}

func parsePaginationQuery(r *http.Request) (*PaginationQuery, error) {
	var offset, limit int
	var err error

	offsetStr := r.URL.Query().Get("offset")
	limitStr := r.URL.Query().Get("limit")
	query := r.URL.Query().Get("q")

	if offsetStr == "" {
		offset = 0
	} else {
		offset, err = strconv.Atoi(offsetStr)
		if err != nil {
			return nil, errors.New("invalid offset")
		}
		if offset < 0 {
			return nil, errors.New("offset must be >= 0")
		}
	}

	if limitStr == "" {
		limit = 20
	} else {
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			return nil, errors.New("invalid limit")
		}
		if limit <= 0 || limit > 500 {
			return nil, errors.New("limit must be between 1 and 100")
		}
	}

	return &PaginationQuery{Offset: offset, Limit: limit, Query: query}, nil
}

func (api *API) writeJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(v)
}

func (api *API) writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(ErrorResponse{
		Error: message,
	}); err != nil {
		api.logger.Error(
			"failed to encode error response", slog.String("error", err.Error()),
		)
	}
}
