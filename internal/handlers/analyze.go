package handlers

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/siddhantprateek/reefline/internal/queue"
	"github.com/siddhantprateek/reefline/pkg/database"
	"github.com/siddhantprateek/reefline/pkg/models"
	"github.com/siddhantprateek/reefline/pkg/tools"
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
//	}
//
// Response:
//
//	{
//	  "job_id": "job_abc123",
//	  "status": "QUEUED",
//	  "stream_url": "/api/v1/jobs/job_abc123/stream"
//	}
//

func (h *AnalyzeHandler) Handle(c *fiber.Ctx) error {
	var req AnalysisRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Validation
	if req.Dockerfile == "" && req.ImageRef == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "At least one of 'dockerfile' or 'image_ref' must be provided"})
	}

	ctx := c.Context()
	jobID := uuid.New().String()

	var skopeoResult *tools.InspectResult
	var metadataJSON []byte

	// Step 1: Skopeo Inspect (if image provided)
	if req.ImageRef != "" {
		if tools.ImgInspector == nil || !tools.ImgInspector.IsEnabled() {
			// Warn or Error? User said "we shall keep skopeo in main api server".
			// If disabled, we might skip, but better to assume it's there.
		} else {
			// Prepare auth
			var auth *tools.ImageAuth
			// TODO: extract credentials from request or DB
			// For now using empty/nil if public

			res, err := tools.ImgInspector.InspectImage(ctx, req.ImageRef, auth)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Failed to inspect image: " + err.Error(),
				})
			}
			skopeoResult = res
		}
	}

	if skopeoResult != nil {
		metadataJSON, _ = json.Marshal(skopeoResult)
	}

	// Step 2: Store in DB
	job := models.Job{
		ID:         jobID,
		JobID:      jobID,
		UserID:     "default-user", // TODO: Auth
		ImageRef:   req.ImageRef,
		Dockerfile: req.Dockerfile,
		Status:     models.JobStatusQueued,
		Scenario:   "image_only", // simplified logic
		Metadata:   string(metadataJSON),
		Progress:   0,
	}
	if req.Dockerfile != "" && req.ImageRef != "" {
		job.Scenario = "both"
	} else if req.Dockerfile != "" {
		job.Scenario = "dockerfile_only"
	}

	if err := database.DB.Create(&job).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create job record: " + err.Error(),
		})
	}

	// Step 3: Enqueue Job
	payload := map[string]interface{}{
		"job_id":      jobID,
		"dockerfile":  req.Dockerfile,
		"image_ref":   req.ImageRef,
		"app_context": req.AppContext,
		"skopeo_meta": skopeoResult,
	}

	queueOpts := []queue.Option{}
	_, err := h.Queue.Enqueue(ctx, "analyze_image", payload, queueOpts...)
	if err != nil {
		// Update DB to failed?
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to enqueue analysis job: " + err.Error()})
	}

	// Return 202 Accepted with metadata
	resp := fiber.Map{
		"job_id":     jobID,
		"status":     "QUEUED",
		"stream_url": "/api/v1/jobs/" + jobID + "/stream",
	}
	if skopeoResult != nil {
		var size int64
		for _, l := range skopeoResult.Layers {
			size += l.Size
		}

		resp["image_info"] = fiber.Map{
			"size":    size,
			"arch":    skopeoResult.Architecture,
			"os":      skopeoResult.Os,
			"digest":  skopeoResult.Digest,
			"created": skopeoResult.Created,
		}
	}

	return c.Status(fiber.StatusAccepted).JSON(resp)
}
