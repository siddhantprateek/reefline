package handlers

import (
	"github.com/gofiber/fiber/v2"
)

// ReportHandler handles downloading report artifacts from MinIO/S3
type ReportHandler struct{}

// NewReportHandler creates a new ReportHandler instance
func NewReportHandler() *ReportHandler {
	return &ReportHandler{}
}

// DownloadReport returns the full analysis report JSON for a completed job.
//
// GET /api/v1/jobs/:id/report
// Response: Full report JSON with measured data, proposed optimizations, and tool data.
func (h *ReportHandler) DownloadReport(c *fiber.Ctx) error {
	// TODO: Extract job ID from params
	// TODO: Verify job belongs to authenticated user and is completed
	// TODO: Fetch report.json from MinIO/S3
	// TODO: Return report JSON
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"message": "download report endpoint not yet implemented",
	})
}

// DownloadDockerfile returns the AI-proposed optimized Dockerfile.
//
// GET /api/v1/jobs/:id/dockerfile
// Response: Plain text Dockerfile content with Content-Type: text/plain
func (h *ReportHandler) DownloadDockerfile(c *fiber.Ctx) error {
	// TODO: Extract job ID from params
	// TODO: Verify job belongs to authenticated user and is completed
	// TODO: Fetch optimized.Dockerfile from MinIO/S3
	// TODO: Return Dockerfile content as text/plain
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"message": "download dockerfile endpoint not yet implemented",
	})
}

// DownloadSBOM returns the Software Bill of Materials (SBOM) in SPDX format.
// Only available if the job included an image reference (scenario B or C).
//
// GET /api/v1/jobs/:id/sbom
// Response: SBOM JSON (SPDX format)
func (h *ReportHandler) DownloadSBOM(c *fiber.Ctx) error {
	// TODO: Extract job ID from params
	// TODO: Verify job belongs to authenticated user and is completed
	// TODO: Fetch sbom.spdx.json from MinIO/S3
	// TODO: Return 404 if job was Dockerfile-only (no SBOM generated)
	// TODO: Return SBOM JSON
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"message": "download sbom endpoint not yet implemented",
	})
}

// DownloadGraph returns the build graph SVG visualization.
// Shows multi-stage build dependencies and COPY --from references.
// Only available if the job included a Dockerfile (scenario A or C).
//
// GET /api/v1/jobs/:id/graph
// Response: SVG image with Content-Type: image/svg+xml
func (h *ReportHandler) DownloadGraph(c *fiber.Ctx) error {
	// TODO: Extract job ID from params
	// TODO: Verify job belongs to authenticated user and is completed
	// TODO: Fetch graph.svg from MinIO/S3
	// TODO: Return 404 if job was Image-only (no graph generated)
	// TODO: Return SVG with correct Content-Type
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"message": "download graph endpoint not yet implemented",
	})
}
