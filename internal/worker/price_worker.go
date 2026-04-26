package worker

import (
	"context"
	"investPort/internal/service"
	"log"
	"time"
)

type PriceWorker struct {
	itemService  *service.ItemService
	PriceService *service.PriceHistoryService
}

func NewPriceWorker(itemServ *service.ItemService, historyServ *service.PriceHistoryService) *PriceWorker {
	return &PriceWorker{itemService: itemServ, PriceService: historyServ}
}
func (w *PriceWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)

	for {
		select {
		case <-ctx.Done():
			log.Printf("Worker: shutting down...")
		case <-ticker.C:
			items, err := w.itemService.GetItemsForProcessing()
			if err != nil {
				log.Printf("failed to get items: %v\n", err)
				time.Sleep(2 * time.Second)
				continue
			}
			for i, _ := range items {
				err = w.PriceService.UpdatePriceForItem(&items[i])
				if err != nil {
					log.Printf("failed to update price: %v\n", err)
					time.Sleep(5 * time.Second)
				}
				time.Sleep(4 * time.Second)
			}

		}
	}

}
