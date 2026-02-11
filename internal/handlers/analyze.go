package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/siddhantprateek/reefline/internal/queue"
)

// AnalyzeHandler handles container image analysis requests
type AnalyzeHandler struct {
	Queue queue.Queue
}

// NewAnalyzeHandler creates a new AnalyzeHandler instance
func NewAnalyzeHandler(q queue.Queue) *AnalyzeHandler {
	return &AnalyzeHandler{Queue: q}
}

// AnalysisRequest represents the request body for analysis
type AnalysisRequest struct {
	Dockerfile          string            `json:"dockerfile"`
	ImageRef            string            `json:"image_ref"`
	AppContext          string            `json:"app_context"`
	RegistryCredentials map[string]string `json:"registry_credentials"`
}

// Handle processes a new analysis request.
// Accepts a Dockerfile (text), image reference, or both.
// Determines the input scenario (A: Dockerfile only, B: Image only, C: Both)
// and enqueues an Asynq job for background processing.
//
// POST /api/v1/analyze
// Request body:
//
//	{
//	  "dockerfile": "FROM ubuntu:22.04\n...",   // optional
//	  "image_ref": "nginx:1.25",                // optional
//	  "app_context": "Flask web API",           // optional user hint
//	  "registry_credentials": { ... }           // optional for private registries
//	}
//
// Response:
//
//	{
//	  "job_id": "job_abc123",
//	  "status": "QUEUED",
//	  "stream_url": "/api/v1/jobs/job_abc123/stream"
//	}
func (h *AnalyzeHandler) Handle(c *fiber.Ctx) error {
	var req AnalysisRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Validation
	if req.Dockerfile == "" && req.ImageRef == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "At least one of 'dockerfile' or 'image_ref' must be provided"})
	}

	// Determine Input Scenario (logging only for now)
	// Scenario A: Dockerfile only
	// Scenario B: Image only
	// Scenario C: Both

	// Create Job Payload
	payload := map[string]interface{}{
		"dockerfile":  req.Dockerfile,
		"image_ref":   req.ImageRef,
		"app_context": req.AppContext,
		// credentials handling to be added securely later
	}

	// Enqueue Job
	queueOpts := []queue.Option{}
	// Simulate delay for testing if needed, or immediate
	// For now, let's just enqueue immediately.
	// If we want to simulate the "delay" the user asked for testing, we can inject it here or in worker.
	// User requested: "delay timer of 5sec 10 second 20 second interval then finsh" for testing.
	// I will add a delay field to payload for the worker to respect, OR just a hardcoded delay in worker.
	// But this handle should just return QUEUED.

	jobID, err := h.Queue.Enqueue(c.Context(), "analyze_image", payload, queueOpts...)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to enqueue analysis job: " + err.Error()})
	}

	// Return Response
	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"job_id":     jobID,
		"status":     "QUEUED",
		"stream_url": "/api/v1/jobs/" + jobID + "/stream",
	})
}
