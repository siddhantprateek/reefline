package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/siddhantprateek/reefline/internal/queue"
)

// JobsHandler handles job listing, status checking, and deletion
type JobsHandler struct {
	Queue queue.Queue
}

// NewJobsHandler creates a new JobsHandler instance
func NewJobsHandler(q queue.Queue) *JobsHandler {
	return &JobsHandler{Queue: q}
}

// List returns all jobs for the authenticated user.
//
// GET /api/v1/jobs
// Query params:
//   - page (int, default 1)
//   - limit (int, default 20)
//   - status (string, optional filter: QUEUED | RUNNING | COMPLETED | FAILED)
//
// Response:
//
//	{
//	  "jobs": [ { "job_id": "...", "status": "...", "created_at": "..." } ],
//	  "total": 42,
//	  "page": 1,
//	  "limit": 20
//	}
func (h *JobsHandler) List(c *fiber.Ctx) error {
	// TODO: Get authenticated user from context
	// TODO: Parse pagination and filter query params
	// TODO: Query PostgreSQL for user's jobs
	// TODO: Return paginated job list
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"message": "list jobs endpoint not yet implemented",
	})
}

// Get returns the status and report for a specific job.
//
// GET /api/v1/jobs/:id
// Response (when complete):
//
//	{
//	  "job_id": "job_abc123",
//	  "status": "COMPLETED",
//	  "input_scenario": "both",
//	  "report": { "measured": { ... }, "proposed": { ... }, "tool_data": { ... } }
//	}
func (h *JobsHandler) Get(c *fiber.Ctx) error {
	jobID := c.Params("id")
	if jobID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Job ID is required"})
	}

	// Get status from Queue (Redis/Memory)
	status, err := h.Queue.GetJobStatus(c.Context(), jobID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Job not found or error retrieving status: " + err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"job_id": jobID,
		"status": status,
		// TODO: Include report if COMPLETED (fetch from DB/Storage)
	})
}

// Delete removes a job and all its associated artifacts.
//
// DELETE /api/v1/jobs/:id
// Response:
//
//	{ "message": "job deleted" }
func (h *JobsHandler) Delete(c *fiber.Ctx) error {
	// TODO: Extract job ID from params
	// TODO: Verify job belongs to authenticated user
	// TODO: Delete artifacts from MinIO/S3
	// TODO: Delete job record from PostgreSQL
	// TODO: Return success response
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"message": "delete job endpoint not yet implemented",
	})
}
