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
	"github.com/siddhantprateek/reefline/internal/routes"
	"github.com/siddhantprateek/reefline/pkg/database"
	"github.com/siddhantprateek/reefline/pkg/models"
	"github.com/siddhantprateek/reefline/pkg/storage"
	"github.com/siddhantprateek/reefline/pkg/telemetry"
)

func main() {
	// Initialize telemetry
	telemetryConfig := telemetry.GetConfigFromEnv()
	shutdown := telemetry.Initialize(telemetryConfig)
	defer shutdown()

	// Initialize database
	dbConfig := database.GetConfigFromEnv()
	db, err := database.Initialize(dbConfig)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Run migrations (add your models here)
	if err := database.AutoMigrate(db, &models.Integration{}); err != nil {
		log.Fatalf("Failed to run database migrations: %v", err)
	}

	// Initialize MinIO storage
	storageConfig := storage.GetConfigFromEnv()
	_, err = storage.Initialize(storageConfig)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	app := fiber.New(fiber.Config{
		AppName: "Reefline Server",
	})

	// Add telemetry middleware first
	app.Use(otelfiber.Middleware())
	app.Use(cors.New())
	app.Use(logger.New())
	app.Use(recover.New())

	routes.Setup(app)

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
