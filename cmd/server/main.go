package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/contrib/otelfiber"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/siddhantprateek/reefline/pkg/telemetry"
)

func main() {
	// Initialize telemetry
	telemetryConfig := telemetry.GetConfigFromEnv()
	shutdown := telemetry.Initialize(telemetryConfig)
	defer shutdown()

	app := fiber.New(fiber.Config{
		AppName: "Reefline Server",
	})

	// Add telemetry middleware first
	app.Use(otelfiber.Middleware())
	app.Use(cors.New())
	app.Use(logger.New())
	app.Use(recover.New())

	setupRoutes(app)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create channel for graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("Gracefully shutting down server...")
		app.Shutdown()
	}()

	log.Printf("Starting Reefline Server on port %s", port)
	log.Fatal(app.Listen(":" + port))
}

func setupRoutes(app *fiber.App) {
	api := app.Group("/api/v1")

	setupHealthRoutes(api)
}

func setupHealthRoutes(api fiber.Router) {
	health := api.Group("/health")

	health.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"service": "reefline-server",
		})
	})

	health.Get("/ready", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ready",
			"service": "reefline-server",
		})
	})

	health.Get("/live", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "alive",
			"service": "reefline-server",
		})
	})
}
