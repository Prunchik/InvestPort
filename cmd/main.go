package main

import (
	"context"
	"investPort/db"
	"investPort/internal/repository"
	"investPort/internal/service"
	"investPort/internal/steam"
	"investPort/internal/worker"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	var items []string
	items = append(items,
		"https://steamcommunity.com/market/listings/730/Sealed%20Dead%20Hand%20Terminal",
		"https://steamcommunity.com/market/listings/730/Dreams%20%26%20Nightmares%20Case",
		"https://steamcommunity.com/market/listings/730/Fever%20Case",
	)
	// database
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file: ", err)
	}
	DB, err := db.NewDB(os.Getenv("DSN"))
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}
	// repository
	itemRepo := repository.NewItemRepo(DB.DB)
	priceRepo := repository.NewPriceHistoryRepository(DB.DB)

	//Steam client
	client := steam.NewClient()

	//services
	itemService := service.NewItemService(itemRepo)
	priceService := service.NewPriceHistoryService(priceRepo, client)

	for i, _ := range items {
		_, err = itemService.GetOrCreateByURL(items[i])
		if err != nil {
			panic(err)
		}
	}
	//worker
	priceWorker := worker.NewPriceWorker(itemService, priceService)

	//start
	priceWorker.Start(context.Background())
}
