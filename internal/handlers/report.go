package handlers

import (
	"fmt"
	"io"

	"github.com/gofiber/fiber/v2"
	"github.com/siddhantprateek/reefline/pkg/storage"
)

// ReportHandler handles downloading report artifacts from MinIO
type ReportHandler struct{}

// NewReportHandler creates a new ReportHandler instance
func NewReportHandler() *ReportHandler {
	return &ReportHandler{}
}

func (h *ReportHandler) streamArtifact(c *fiber.Ctx, objectName, filename, contentType string) error {
	bucket := storage.GetConfigFromEnv().DefaultBucket
	object, err := storage.DownloadFile(c.Context(), bucket, objectName)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": fmt.Sprintf("artifact not found: %s", objectName),
		})
	}
	defer object.Close()

	c.Set("Content-Type", contentType)
	if c.Query("download") == "true" {
		c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	}
	_, err = io.Copy(c.Response().BodyWriter(), object)
	return err
}

// DownloadReport returns the full analysis report JSON.
// GET /api/v1/jobs/:id/report
func (h *ReportHandler) DownloadReport(c *fiber.Ctx) error {
	jobID := c.Params("id")
	return h.streamArtifact(c, fmt.Sprintf("%s/report.json", jobID), "report.json", "application/json")
}

// DownloadGrype returns the Grype vulnerability scan result.
// GET /api/v1/jobs/:id/grype.json
func (h *ReportHandler) DownloadGrype(c *fiber.Ctx) error {
	jobID := c.Params("id")
	return h.streamArtifact(c, fmt.Sprintf("%s/artifacts/grype.json", jobID), "grype.json", "application/json")
}

// DownloadDive returns the Dive layer efficiency analysis result.
// GET /api/v1/jobs/:id/dive.json
func (h *ReportHandler) DownloadDive(c *fiber.Ctx) error {
	jobID := c.Params("id")
	return h.streamArtifact(c, fmt.Sprintf("%s/artifacts/dive.json", jobID), "dive.json", "application/json")
}

// DownloadDockle returns the Dockle CIS benchmark scan result.
// GET /api/v1/jobs/:id/dockle.json
func (h *ReportHandler) DownloadDockle(c *fiber.Ctx) error {
	jobID := c.Params("id")
	return h.streamArtifact(c, fmt.Sprintf("%s/artifacts/dockle.json", jobID), "dockle.json", "application/json")
}
