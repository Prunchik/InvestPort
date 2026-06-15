package worker

import (
	"context"
	"investPort/internal/model"
	"investPort/internal/service"
	"log"
	"sync"
	"time"
)

type PriceWorker struct {
	itemService  *service.ItemService
	PriceService *service.PriceHistoryService
	workerCount  int
}

func NewPriceWorker(itemServ *service.ItemService, historyServ *service.PriceHistoryService, workerCount int) *PriceWorker {
	return &PriceWorker{
		itemService:  itemServ,
		PriceService: historyServ,
		workerCount:  workerCount,
	}
}
func (w *PriceWorker) Start(ctx context.Context) {
	itemCh := make(chan *model.Item, w.workerCount*2)

	var wg sync.WaitGroup

	for i := 0; i < w.workerCount; i++ {
		wg.Add(1)

		go func(workerID int) {
			defer wg.Done()
			w.workerLoop(ctx, itemCh, workerID)
		}(i)
	}

	defer func() {
		close(itemCh)
		wg.Wait()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		items, err := w.itemService.ListAll()
		if err != nil {
			log.Printf("failed to get items: %v", err)

			select {
			case <-time.After(10 * time.Second):
			case <-ctx.Done():
				return
			}

			continue
		}

		for i := range items {
			select {
			case itemCh <- &items[i]:
			case <-ctx.Done():
				return
			}
		}

		select {
		case <-time.After(5 * time.Minute):
		case <-ctx.Done():
			return
		}
	}
}
func (w *PriceWorker) workerLoop(ctx context.Context, itemCh <-chan *model.Item, workerID int) {
	for item := range itemCh {

		itemCtx, cancel := context.WithTimeout(ctx, 1*time.Minute)

		err := w.PriceService.UpdatePriceForItem(itemCtx, item)
		cancel()

		if err != nil {
			log.Printf("worker-%d: failed to update price for %s: %v", workerID, item.HashName, err)
		} else {
			log.Printf("worker-%d: updated price for %s", workerID, item.HashName)
		}
	}
}
