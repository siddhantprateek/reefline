package handlers

import (
	"github.com/gofiber/fiber/v2"
)

// CompareHandler handles comparison of two analysis jobs
type CompareHandler struct{}

// NewCompareHandler creates a new CompareHandler instance
func NewCompareHandler() *CompareHandler {
	return &CompareHandler{}
}

// Handle compares two completed analysis jobs side by side.
// Useful for tracking optimization progress (e.g., before/after applying recommendations).
//
// POST /api/v1/compare
// Request body:
//
//	{
//	  "job_id_a": "job_abc123",
//	  "job_id_b": "job_def456"
//	}
//
// Response:
//
//	{
//	  "comparison": {
//	    "size_mb":       { "a": 916, "b": 85,  "delta": -831, "delta_pct": -90.7 },
//	    "total_cves":    { "a": 85,  "b": 12,  "delta": -73,  "delta_pct": -85.9 },
//	    "efficiency_pct":{ "a": 72,  "b": 98,  "delta": 26,   "delta_pct": 36.1  },
//	    "score":         { "a": 42,  "b": 91,  "delta": 49,   "delta_pct": 116.7 },
//	    "layers":        { "a": 12,  "b": 5,   "delta": -7,   "delta_pct": -58.3 },
//	    "runs_as_root":  { "a": true, "b": false }
//	  }
//	}
func (h *CompareHandler) Handle(c *fiber.Ctx) error {
	// TODO: Parse request body (job_id_a, job_id_b)
	// TODO: Validate both job IDs are provided
	// TODO: Verify both jobs belong to authenticated user
	// TODO: Verify both jobs are completed
	// TODO: Fetch reports for both jobs from PostgreSQL
	// TODO: Compute deltas for size, CVEs, efficiency, score, layers, root status
	// TODO: Return comparison response
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"message": "compare endpoint not yet implemented",
	})
}
