package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/siddhantprateek/reefline/internal/handlers"
)

// Setup configures all application routes
func Setup(app *fiber.App) {
	api := app.Group("/api/v1")

	setupHealthRoutes(api)
}

// setupHealthRoutes configures health check endpoints
func setupHealthRoutes(api fiber.Router) {
	healthHandler := handlers.NewHealthHandler()
	health := api.Group("/health")

	health.Get("/", healthHandler.Status)
	health.Get("/ready", healthHandler.Ready)
	health.Get("/live", healthHandler.Live)
}
