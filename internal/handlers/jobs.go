package handlers

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/siddhantprateek/reefline/internal/queue"
	"github.com/siddhantprateek/reefline/pkg/database"
	"github.com/siddhantprateek/reefline/pkg/models"
	"github.com/siddhantprateek/reefline/pkg/storage"
)

// JobsHandler handles job listing, status checking, and deletion
type JobsHandler struct {
	Queue queue.Queue
}

// NewJobsHandler creates a new JobsHandler instance
func NewJobsHandler(q queue.Queue) *JobsHandler {
	return &JobsHandler{Queue: q}
}

// JobListResponse represents a single job in the list response
type JobListResponse struct {
	ID           string  `json:"id"`
	JobID        string  `json:"job_id"`
	ImageRef     string  `json:"image_ref,omitempty"`
	Dockerfile   string  `json:"dockerfile,omitempty"`
	Status       string  `json:"status"`
	Scenario     string  `json:"scenario,omitempty"`
	ErrorMessage string  `json:"error_message,omitempty"`
	Progress     int     `json:"progress"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
	CompletedAt  *string `json:"completed_at,omitempty"`
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
//	[
//	  { "job_id": "...", "status": "...", "created_at": "..." },
//	  ...
//	]
func (h *JobsHandler) List(c *fiber.Ctx) error {
	ctx := c.Context()

	// TODO: Get authenticated user from context
	userID := "default-user"

	// Parse pagination params
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "100"))
	statusFilter := c.Query("status", "")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit

	// Build query
	query := database.DB.WithContext(ctx).Model(&models.Job{}).Where("user_id = ?", userID)

	// Apply status filter if provided
	if statusFilter != "" {
		query = query.Where("status = ?", statusFilter)
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to count jobs: " + err.Error(),
		})
	}

	// Get jobs
	var jobs []models.Job
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&jobs).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch jobs: " + err.Error(),
		})
	}

	// Convert to response format
	response := make([]JobListResponse, len(jobs))
	for i, job := range jobs {
		response[i] = JobListResponse{
			ID:           job.ID,
			JobID:        job.JobID,
			ImageRef:     job.ImageRef,
			Dockerfile:   job.Dockerfile,
			Status:       string(job.Status),
			Scenario:     job.Scenario,
			ErrorMessage: job.ErrorMessage,
			Progress:     job.Progress,
			CreatedAt:    job.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:    job.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
		if job.CompletedAt != nil {
			completedAt := job.CompletedAt.Format("2006-01-02T15:04:05Z07:00")
			response[i].CompletedAt = &completedAt
		}
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// JobReportResponse represents the full job report response
type JobReportResponse struct {
	JobID         string                 `json:"job_id"`
	Status        string                 `json:"status"`
	InputScenario string                 `json:"input_scenario"`
	Report        map[string]interface{} `json:"report,omitempty"`
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
	ctx := c.Context()
	jobID := c.Params("id")
	if jobID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Job ID is required"})
	}

	// TODO: Get authenticated user from context
	userID := "default-user"

	// Fetch job from database
	var job models.Job
	if err := database.DB.WithContext(ctx).Where("job_id = ? AND user_id = ?", jobID, userID).First(&job).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Job not found",
		})
	}

	response := JobReportResponse{
		JobID:         job.JobID,
		Status:        string(job.Status),
		InputScenario: job.Scenario,
	}

	// If job is completed, fetch the report from MinIO
	if job.Status == models.JobStatusCompleted {
		bucket := getEnv("MINIO_DEFAULT_BUCKET", "reefline")
		reportPath := fmt.Sprintf("%s/report.json", jobID)

		// Try to download the report
		reportObj, err := storage.DownloadFile(ctx, bucket, reportPath)
		if err != nil {
			// Report might not exist yet, return job status without report
			return c.Status(fiber.StatusOK).JSON(response)
		}
		defer reportObj.Close()

		// Parse the report JSON
		var reportData map[string]interface{}
		if err := json.NewDecoder(reportObj).Decode(&reportData); err != nil {
			// Failed to parse report, return without it
			return c.Status(fiber.StatusOK).JSON(response)
		}

		response.Report = reportData
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// Delete removes a job and all its associated artifacts.
//
// DELETE /api/v1/jobs/:id
// Response:
//
//	{ "message": "job deleted" }
func (h *JobsHandler) Delete(c *fiber.Ctx) error {
	ctx := c.Context()
	jobID := c.Params("id")
	if jobID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Job ID is required"})
	}

	// TODO: Get authenticated user from context
	userID := "default-user"

	// Fetch job from database to verify ownership
	var job models.Job
	if err := database.DB.WithContext(ctx).Where("job_id = ? AND user_id = ?", jobID, userID).First(&job).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Job not found",
		})
	}

	// Delete artifacts from MinIO
	bucket := getEnv("MINIO_DEFAULT_BUCKET", "reefline")
	artifacts := []string{
		fmt.Sprintf("%s/report.json", jobID),
		fmt.Sprintf("%s/dockerfile", jobID),
		fmt.Sprintf("%s/sbom.json", jobID),
		fmt.Sprintf("%s/graph.svg", jobID),
	}

	for _, artifact := range artifacts {
		// Ignore errors if artifact doesn't exist
		_ = storage.DeleteFile(ctx, bucket, artifact)
	}

	// Delete job record from database (soft delete)
	if err := database.DB.WithContext(ctx).Delete(&job).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete job: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Job deleted successfully",
	})
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
