package api

import (
	_ "investPort/docs"

	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// @title			InvestPort API
// @version			1.0
// @description		Steam Market Price Tracker
// @host			localhost:8080
// @BasePath		/
func RegisterSwaggerRoutes(router chi.Router) {
	router.Get("/swagger/*", httpSwagger.Handler())
}
