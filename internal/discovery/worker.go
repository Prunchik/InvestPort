package discovery

import (
	"context"
	"investPort/internal/service"
	"investPort/internal/steam"
	"log/slog"
	"net/url"
	"regexp"
	"time"
)

type Worker struct {
	source      Source
	itemService *service.ItemService
	client      *steam.Client
	logger      *slog.Logger
}

func NewWorker(source Source, itemService *service.ItemService, client *steam.Client, logger *slog.Logger) *Worker {
	return &Worker{
		source:      source,
		itemService: itemService,
		client:      client,
		logger:      logger,
	}
}

var wearRegexp = regexp.MustCompile(`\((Factory New|Minimal Wear|Field-Tested|Well-Worn|Battle-Scarred)\)`)

func (w *Worker) Start(ctx context.Context) {
	discoveredItems, err := w.source.fetch(ctx)
	if err != nil {
		w.logger.Error("failed to fetch discovered items", "error", err)
		return
	}
	w.logger.Info("fetched discovered items", "count", len(discoveredItems))

	existingItems, err := w.itemService.ListAll()
	if err != nil {
		w.logger.Error("failed to get items from db", "error", err)
		return
	}
	w.logger.Info("fetched existing items from db", "count", len(existingItems))

	existingMap := make(map[string]struct{}, len(discoveredItems))

	for _, item := range existingItems {
		existingMap[item.HashName] = struct{}{}
	}

	var newItems []DiscoveredItem
	for _, di := range discoveredItems {
		if _, exist := existingMap[di.MarketHashName]; !exist {
			newItems = append(newItems, di)
		}
	}
	w.logger.Info("found new items to add", "count", len(newItems))

	for _, newItem := range newItems {

		ItemUrl := "https://steamcommunity.com/market/listings/730/" + url.PathEscape(newItem.MarketHashName)

		_, err = w.client.ResolveItemName(ctx, ItemUrl)
		if err != nil {
			w.logger.Info("couldn't get the item name", "error", err, "url", ItemUrl)
			time.Sleep(time.Second)
			continue
		}

		item := &steam.ParsedItem{
			Name:         newItem.MarketHashName,
			URL:          ItemUrl,
			AppID:        730,
			HashName:     newItem.MarketHashName,
			WearCategory: ExtractWearCategory(newItem.MarketHashName),
			ImageURL:     newItem.ImageURL,
		}

		_, err = w.itemService.GetOrCreateItem(item)
		if err != nil {
			w.logger.Error("failed to create item", "hash_name", newItem.MarketHashName, "error", err)
			return
		}
		w.logger.Info("created new item", "hash_name", newItem.MarketHashName)
		time.Sleep(400 * time.Millisecond)
	}
}

func ExtractWearCategory(name string) *int {

	matches := wearRegexp.FindStringSubmatch(name)
	if len(matches) < 2 {
		return nil
	}

	if value, ok := steam.WearCategoryValue[matches[1]]; ok {
		return &value
	}
	return nil
}
