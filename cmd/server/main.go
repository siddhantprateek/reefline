package main

import (
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/contrib/otelfiber"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
	"github.com/siddhantprateek/reefline/internal/queue"
	"github.com/siddhantprateek/reefline/internal/routes"
	"github.com/siddhantprateek/reefline/pkg/crypto"
	"github.com/siddhantprateek/reefline/pkg/database"
	"github.com/siddhantprateek/reefline/pkg/models"
	"github.com/siddhantprateek/reefline/pkg/storage"
	"github.com/siddhantprateek/reefline/pkg/telemetry"
	"github.com/siddhantprateek/reefline/pkg/tools"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found")
	}

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
	if err := database.AutoMigrate(db, &models.Integration{}, &models.Job{}); err != nil {
		log.Fatalf("Failed to run database migrations: %v", err)
	}

	// Initialize MinIO storage
	storageConfig := storage.GetConfigFromEnv()
	_, err = storage.Initialize(storageConfig)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	// Initialize encryption (AES-256-GCM)
	if err := crypto.Init(); err != nil {
		log.Fatalf("Failed to initialize encryption: %v", err)
	}
	log.Println("Encryption subsystem initialized (AES-256-GCM)")

	// Initialize image inspector (skopeo-like inspect via containers/image)
	enableInspector := os.Getenv("IMAGE_INSPECTOR_ENABLED")
	if enableInspector == "true" {
		log.Println("Initializing image inspector...")
		inspectorConfig := tools.ImageInspectorConfig{
			Enable:                true,
			InsecureSkipTLSVerify: os.Getenv("IMAGE_INSPECTOR_INSECURE_TLS") == "true",
		}
		tools.ImgInspector = tools.NewImageInspector(inspectorConfig, slog.Default())
		tools.ImgInspector.Init()
		log.Println("Image inspector initialized (containers/image)")
	} else {
		log.Println("Image inspector is disabled (set IMAGE_INSPECTOR_ENABLED=true to enable)")
	}

	// Initialize Job Queue
	var q queue.Queue
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost != "" {
		redisPort := os.Getenv("REDIS_PORT")
		if redisPort == "" {
			redisPort = "6379"
		}
		redisAddr := redisHost + ":" + redisPort
		redisPass := os.Getenv("REDIS_PASSWORD")
		q = queue.NewRedisQueue(redisAddr, redisPass)
		log.Printf("Using Redis job queue at %s", redisAddr)
	} else {
		// Fallback to In-Memory
		q = queue.NewInMemoryQueue(100)
		log.Println("Using In-Memory job queue")
	}

	// Start Queue (for enqueueing only, no workers needed here technically if using Redis,
	// but Asynq client doesn't need Start. However, our interface might expect it?
	// RedisQueue.Start() starts the server. We don't need the server here.
	// But let's check queue implementation.
	// RedisQueue.Start() calls q.server.Run().
	// We ONLY need the client to enqueue.
	// But existing code called q.Start().
	// IF we call q.Start(), this binary will ALSO process jobs if we registered handlers.
	// We are NOT registering handlers here.
	// So q.Start() will run a server with NO handlers. That's fine, just strict separation.
	// Actually, better to NOT start the server if we don't want to process jobs.
	// But our interface might strictly require Start().
	// Let's look at RedisQueue.Start(): "return q.server.Run(q.mux)"
	// If we don't call Start(), we can still usage q.Enqueue().
	// So we can SKIP Start() in the server?
	// The interface `queue.Queue` likely has Start/Stop.
	// If I skip it, I might break the interface contract if I assign to `queue.Queue`.
	// Let's call it for now to avoid breaking changes, but since no handlers are registered, it does nothing.
	// 	if err := q.Start(); err != nil {
	// 		log.Printf("Failed to start job queue: %v", err)
	// 	}
	defer q.Stop()

	app := fiber.New(fiber.Config{
		AppName: "Reefline Server",
	})

	// Add telemetry middleware first
	app.Use(otelfiber.Middleware())
	app.Use(cors.New())
	app.Use(logger.New())
	app.Use(recover.New())

	routes.Setup(app, q)

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

		// Stop image inspector if initialized
		if tools.ImgInspector != nil {
			log.Println("Stopping image inspector...")
			tools.ImgInspector.Stop()
		}

		app.Shutdown()
	}()

	log.Printf("Starting Reefline Server on port %s", port)
	log.Fatal(app.Listen(":" + port))
}
