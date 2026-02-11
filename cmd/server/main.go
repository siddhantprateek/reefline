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
	"github.com/siddhantprateek/reefline/internal/worker"
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
	if err := database.AutoMigrate(db, &models.Integration{}); err != nil {
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

	// Initialize vulnerability scanner
	enableScanner := os.Getenv("VULNERABILITY_SCANNER_ENABLED")
	if enableScanner == "true" {
		log.Println("Initializing vulnerability scanner...")
		scannerConfig := tools.ImageScans{
			Enable: true,
			Exclusions: tools.Exclusions{
				Namespaces: []string{},
				Labels:     map[string][]string{},
			},
		}
		tools.ImgScanner = tools.NewImageScanner(scannerConfig, slog.Default())

		// Initialize scanner in background
		go func() {
			tools.ImgScanner.Init("reefline", "1.0.0")
			log.Println("Vulnerability scanner initialized successfully")
		}()
	} else {
		log.Println("Vulnerability scanner is disabled (set VULNERABILITY_SCANNER_ENABLED=true to enable)")
	}

	// Initialize dockle (CIS Docker Benchmark) scanner
	enableDockle := os.Getenv("DOCKLE_SCANNER_ENABLED")
	if enableDockle == "true" {
		log.Println("Initializing dockle scanner...")
		dockleConfig := tools.DockleConfig{
			Enable: true,
		}
		tools.DockleScn = tools.NewDockleScanner(dockleConfig, slog.Default())
		tools.DockleScn.Init()
		log.Println("Dockle scanner initialized (CIS Docker Benchmark)")
	} else {
		log.Println("Dockle scanner is disabled (set DOCKLE_SCANNER_ENABLED=true to enable)")
	}

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

	// Register Handler
	q.RegisterHandler("analyze_image", worker.ProcessAnalyzeJob)

	// Start Queue
	if err := q.Start(); err != nil {
		log.Printf("Failed to start job queue: %v", err)
		// Fallback to in-memory? Or fatal?
		// User said: "one is in-memory if Redis is provided we chose Redis queue"
		// If Redis provided but fails, maybe fail?
		// But for development, I'll switch to Memory if Redis fails to connect?
		// asynq.NewServer doesn't connect immediately?
		// "Start" returns error?
		// My RedisQueue.Start does return error but it runs async.
		// Let's stick to Redis if configured.
	}
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

		// Stop vulnerability scanner if initialized
		if tools.ImgScanner != nil {
			log.Println("Stopping vulnerability scanner...")
			tools.ImgScanner.Stop()
		}

		// Stop dockle scanner if initialized
		if tools.DockleScn != nil {
			log.Println("Stopping dockle scanner...")
			tools.DockleScn.Stop()
		}

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
