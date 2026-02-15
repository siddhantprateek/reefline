package main

import (
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/siddhantprateek/reefline/internal/queue"
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

	// Initialize dive analyzer (image layer efficiency analysis)
	enableDive := os.Getenv("DIVE_ANALYZER_ENABLED")
	if enableDive == "true" {
		log.Println("Initializing dive analyzer...")
		diveConfig := tools.DiveConfig{
			Enable:       true,
			Source:       os.Getenv("DIVE_IMAGE_SOURCE"),
			IgnoreErrors: os.Getenv("DIVE_IGNORE_ERRORS") == "true",
		}
		// Default to docker if not specified
		if diveConfig.Source == "" {
			diveConfig.Source = "docker"
		}
		tools.DiveAnalyzer = tools.NewDiveAnalyzer(diveConfig, slog.Default())
		tools.DiveAnalyzer.Init()
		log.Println("Dive analyzer initialized (image efficiency analysis)")
	} else {
		log.Println("Dive analyzer is disabled (set DIVE_ANALYZER_ENABLED=true to enable)")
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

	// Log flow service configuration
	flowURL := os.Getenv("FLOW_SERVICE_URL")
	if flowURL == "" {
		flowURL = "http://localhost:8000"
	}
	flowProvider := os.Getenv("FLOW_PROVIDER")
	if flowProvider == "" {
		flowProvider = "openai"
	}
	log.Printf("Flow service URL: %s", flowURL)
	log.Printf("Flow provider:    %s", flowProvider)

	// Register Handler
	q.RegisterHandler("analyze_image", worker.ProcessAnalyzeJob)

	// Start Queue
	log.Println("Starting worker...")
	if err := q.Start(); err != nil {
		log.Fatalf("Failed to start job queue: %v", err)
	}

	// Wait for interrupt signal using channel
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Println("Gracefully shutting down worker...")

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

	// Stop dive analyzer if initialized
	if tools.DiveAnalyzer != nil {
		log.Println("Stopping dive analyzer...")
		tools.DiveAnalyzer.Stop()
	}

	q.Stop()
	log.Println("Worker stopped")
}
