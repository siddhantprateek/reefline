package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/siddhantprateek/reefline/internal/handlers"
	"github.com/siddhantprateek/reefline/internal/queue"
)

// Setup configures all application routes
func Setup(app *fiber.App, q queue.Queue) {
	api := app.Group("/api/v1")

	setupHealthRoutes(api)
	setupAnalyzeRoutes(api, q)
	setupJobRoutes(api, q)
	setupCompareRoutes(api)
	setupIntegrationRoutes(api)
	setupMetricsRoutes(api, q)
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
func setupAnalyzeRoutes(api fiber.Router, q queue.Queue) {
	analyzeHandler := handlers.NewAnalyzeHandler(q)

	// POST /api/v1/analyze — Submit Dockerfile and/or image ref for analysis
	api.Post("/analyze", analyzeHandler.Handle)
}

// setupJobRoutes configures job management and artifact download endpoints
func setupJobRoutes(api fiber.Router, q queue.Queue) {
	jobsHandler := handlers.NewJobsHandler(q)
	reportHandler := handlers.NewReportHandler()
	sseHandler := handlers.NewSSEHandler()

	jobs := api.Group("/jobs")

	// GET  /api/v1/jobs         — List user's jobs

	// GET  /api/v1/jobs         — List user's jobs
	jobs.Get("/", jobsHandler.List)

	// GET    /api/v1/jobs/:id   — Get job status + report
	// DELETE /api/v1/jobs/:id   — Delete job + artifacts
	jobs.Get("/:id", jobsHandler.Get)
	jobs.Delete("/:id", jobsHandler.Delete)

	// GET /api/v1/jobs/:id/stream     — SSE real-time progress
	jobs.Get("/:id/stream", sseHandler.Stream)

	// GET /api/v1/jobs/:id/report      — Download full report JSON
	// GET /api/v1/jobs/:id/grype.json  — Grype vulnerability scan result
	// GET /api/v1/jobs/:id/dive.json   — Dive layer efficiency analysis
	// GET /api/v1/jobs/:id/dockle.json — Dockle CIS benchmark scan result
	jobs.Get("/:id/report", reportHandler.DownloadReport)
	jobs.Get("/:id/grype.json", reportHandler.DownloadGrype)
	jobs.Get("/:id/dive.json", reportHandler.DownloadDive)
	jobs.Get("/:id/dockle.json", reportHandler.DownloadDockle)
}

// setupCompareRoutes configures the comparison endpoint
func setupCompareRoutes(api fiber.Router) {
	compareHandler := handlers.NewCompareHandler()

	// POST /api/v1/compare — Compare two completed analysis jobs
	api.Post("/compare", compareHandler.Handle)
}

// setupIntegrationRoutes configures integration management and provider-specific endpoints
func setupIntegrationRoutes(api fiber.Router) {
	integrationHandler := handlers.NewIntegrationHandler()

	integrations := api.Group("/integrations")

	// GET  /api/v1/integrations     — List all integrations with status
	integrations.Get("/", integrationHandler.List)

	// GET  /api/v1/integrations/:id  — Get specific integration details
	integrations.Get("/:id", integrationHandler.Get)

	// POST /api/v1/integrations/:id/connect    — Save credentials and validate
	// POST /api/v1/integrations/:id/disconnect  — Remove credentials
	// POST /api/v1/integrations/:id/test        — Re-validate existing credentials
	integrations.Post("/:id/connect", integrationHandler.Connect)
	integrations.Post("/:id/disconnect", integrationHandler.Disconnect)
	integrations.Post("/:id/test", integrationHandler.TestConnection)

	// === GitHub-specific endpoints ===
	gh := integrations.Group("/github")

	// GET  /api/v1/integrations/github/repos          — List GitHub repositories
	gh.Get("/repos", integrationHandler.ListGitHubRepos)

	// GET  /api/v1/integrations/github/repos/:owner/:repo/dockerfile — Fetch Dockerfile from repo
	gh.Get("/repos/:owner/:repo/dockerfile", integrationHandler.GetGitHubDockerfile)

	// GET  /api/v1/integrations/github/images          — List GHCR container images
	gh.Get("/images", integrationHandler.ListGitHubContainerImages)

	// POST /api/v1/integrations/github/repos/:owner/:repo/issues — Create optimization issue
	gh.Post("/repos/:owner/:repo/issues", integrationHandler.CreateGitHubIssue)

	// === Docker Hub-specific endpoints ===
	docker := integrations.Group("/docker")

	// GET /api/v1/integrations/docker/repos                         — List Docker Hub repos
	docker.Get("/repos", integrationHandler.ListDockerHubRepos)

	// GET /api/v1/integrations/docker/repos/:namespace/:repo/tags   — List tags for a repo
	docker.Get("/repos/:namespace/:repo/tags", integrationHandler.ListDockerHubTags)

	// === Harbor-specific endpoints ===
	harbor := integrations.Group("/harbor")

	// GET /api/v1/integrations/harbor/projects                                       — List Harbor projects
	harbor.Get("/projects", integrationHandler.ListHarborProjects)

	// GET /api/v1/integrations/harbor/projects/:project/repos/:repo/artifacts         — List artifacts
	harbor.Get("/projects/:project/repos/:repo/artifacts", integrationHandler.ListHarborArtifacts)
}

// setupMetricsRoutes configures analytics and metrics endpoints
func setupMetricsRoutes(api fiber.Router, q queue.Queue) {
	metricsHandler := handlers.NewMetricsHandler(q)

	metrics := api.Group("/metrics")

	// GET /api/v1/metrics/queue — Real-time queue statistics
	metrics.Get("/queue", metricsHandler.GetQueueStats)

	// GET /api/v1/metrics/jobs?time_range=24h|7d|30d — Job metrics and trends
	metrics.Get("/jobs", metricsHandler.GetJobMetrics)

	// GET /api/v1/metrics/tools — Tool performance metrics
	metrics.Get("/tools", metricsHandler.GetToolPerformance)
}
