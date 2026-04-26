package main

import (
	"context"
	"investPort/db"
	"investPort/internal/api"
	"investPort/internal/bootstrap"
	"investPort/internal/repository"
	"investPort/internal/service"
	"investPort/internal/steam"
	"investPort/internal/worker"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

func main() {
	//router
	router := chi.NewRouter()
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

	//Items
	bootstrap.SeedItems(itemService)

	//worker
	priceWorker := worker.NewPriceWorker(itemService, priceService)
	//start
	go priceWorker.Start(context.Background())
	//api
	api.NewApi(router, itemService, priceService)
	//server

	server := http.Server{
		Addr:    ":8080",
		Handler: router,
	}
	defer server.Close()
	err = server.ListenAndServe()
	if err != nil {
		panic(err)
	}

}
