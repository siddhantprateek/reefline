package handlers

import (
	"github.com/gofiber/fiber/v2"
)

// AnalyzeHandler handles container image analysis requests
type AnalyzeHandler struct{}

// NewAnalyzeHandler creates a new AnalyzeHandler instance
func NewAnalyzeHandler() *AnalyzeHandler {
	return &AnalyzeHandler{}
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
	// TODO: Parse request body (dockerfile, image_ref, app_context, registry_credentials)
	// TODO: Validate at least one of dockerfile or image_ref is provided
	// TODO: Determine input scenario (A, B, or C)
	// TODO: Create job record in PostgreSQL
	// TODO: Enqueue Asynq task for background processing
	// TODO: Return job_id, status, and stream_url
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"message": "analyze endpoint not yet implemented",
	})
}
