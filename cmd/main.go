package main

import (
	"context"
	"errors"
	"investPort/db"
	"investPort/internal/api"
	"investPort/internal/bootstrap"
	"investPort/internal/repository"
	"investPort/internal/service"
	"investPort/internal/steam"
	"investPort/internal/web"
	"investPort/internal/worker"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog/v3"
	"github.com/joho/godotenv"
)

const workerCount = 2

func main() {

	// Загрузка переменных окружения
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Warning: .env file not found or failed to load")
	}

	// Настройка логгера с ECS-схемой и упрощённым выводом
	logFormat := httplog.SchemaECS.Concise(true)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		ReplaceAttr: logFormat.ReplaceAttr,
	}))

	// Создание маршрутизатора
	router := chi.NewRouter()

	// Добавление идентификатора запроса
	router.Use(middleware.RequestID)

	// добавление времени ожидания
	router.Use(middleware.Timeout(30 * time.Second))

	// Middleware для логирования HTTP-запросов
	router.Use(httplog.RequestLogger(logger, &httplog.Options{
		Level:         slog.LevelInfo,
		Schema:        httplog.SchemaECS,
		RecoverPanics: true,
		Skip: func(req *http.Request, respStatus int) bool {
			return respStatus == 404 || respStatus == 405
		},
		LogRequestBody:  isDebugHeaderSet,
		LogResponseBody: isDebugHeaderSet,
	}))

	// Middleware для установки атрибутов в контекст
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			httplog.SetAttrs(ctx)

			next.ServeHTTP(w, r)
		})
	})

	// Подключение к базе данных
	dsn := os.Getenv("DSN")
	if dsn == "" {
		logger.Error("DSN environment variable is required")
		return
	}
	dbConn, err := db.NewDB(dsn)
	if err != nil {
		logger.Error("Failed to initialize DB connection:",
			slog.String("error", err.Error()))
		return
	}

	// Репозитории
	itemRepo := repository.NewItemRepo(dbConn.DB)
	priceRepo := repository.NewPriceHistoryRepository(dbConn.DB)

	// Клиент Steam API
	client := steam.NewClient()

	// Сервисы
	itemService := service.NewItemService(itemRepo)
	priceService := service.NewPriceHistoryService(priceRepo, itemRepo, client)

	// Предварительная загрузка предметов (seed)
	go bootstrap.SeedItems(itemService, client)

	// Запуск фонового воркера
	priceWorker := worker.NewPriceWorker(itemService, priceService, workerCount)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Warn("Worker panicked", slog.Any("recover", r))
				return
			}
		}()
		priceWorker.Start(context.Background())
	}()

	// Инициализация API
	api.NewAPI(router, itemService, priceService, client, logger, web.StaticFileServer(), web.IndexHTML)

	// Настройка и запуск сервера
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	if port[0] != ':' {
		port = ":" + port
	}

	server := &http.Server{
		Addr:    port,
		Handler: router,
	}

	addr := "http://localhost" + port
	logger.Info("Server starting", slog.String("url", addr), slog.String("swagger", addr+"/swagger/index.html"))
	defer server.Close()
	if err = server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("Server failed to start", slog.String("error", err.Error()))
		return
	}

}

func isDebugHeaderSet(r *http.Request) bool {
	return r.Header.Get("Debug") == "reveal-body-logs"
}
