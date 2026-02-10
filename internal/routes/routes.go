package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/siddhantprateek/reefline/internal/handlers"
)

// Setup configures all application routes
func Setup(app *fiber.App) {
	api := app.Group("/api/v1")

	setupHealthRoutes(api)
	setupAnalyzeRoutes(api)
	setupJobRoutes(api)
	setupCompareRoutes(api)
}

// setupHealthRoutes configures health check endpoints
func setupHealthRoutes(api fiber.Router) {
	healthHandler := handlers.NewHealthHandler()
	health := api.Group("/health")

	health.Get("/", healthHandler.Status)
	health.Get("/ready", healthHandler.Ready)
	health.Get("/live", healthHandler.Live)
}

// setupAnalyzeRoutes configures the analysis submission endpoint
func setupAnalyzeRoutes(api fiber.Router) {
	analyzeHandler := handlers.NewAnalyzeHandler()

	// POST /api/v1/analyze — Submit Dockerfile and/or image ref for analysis
	api.Post("/analyze", analyzeHandler.Handle)
}

// setupJobRoutes configures job management and artifact download endpoints
func setupJobRoutes(api fiber.Router) {
	jobsHandler := handlers.NewJobsHandler()
	reportHandler := handlers.NewReportHandler()
	sseHandler := handlers.NewSSEHandler()

	jobs := api.Group("/jobs")

	// GET  /api/v1/jobs         — List user's jobs
	jobs.Get("/", jobsHandler.List)

	// GET    /api/v1/jobs/:id   — Get job status + report
	// DELETE /api/v1/jobs/:id   — Delete job + artifacts
	jobs.Get("/:id", jobsHandler.Get)
	jobs.Delete("/:id", jobsHandler.Delete)

	// GET /api/v1/jobs/:id/stream     — SSE real-time progress
	jobs.Get("/:id/stream", sseHandler.Stream)

	// GET /api/v1/jobs/:id/report     — Download full report JSON
	// GET /api/v1/jobs/:id/dockerfile — Download proposed optimized Dockerfile
	// GET /api/v1/jobs/:id/sbom       — Download SBOM (SPDX format)
	// GET /api/v1/jobs/:id/graph      — Download build graph SVG
	jobs.Get("/:id/report", reportHandler.DownloadReport)
	jobs.Get("/:id/dockerfile", reportHandler.DownloadDockerfile)
	jobs.Get("/:id/sbom", reportHandler.DownloadSBOM)
	jobs.Get("/:id/graph", reportHandler.DownloadGraph)
}

// setupCompareRoutes configures the comparison endpoint
func setupCompareRoutes(api fiber.Router) {
	compareHandler := handlers.NewCompareHandler()

	// POST /api/v1/compare — Compare two completed analysis jobs
	api.Post("/compare", compareHandler.Handle)
}
