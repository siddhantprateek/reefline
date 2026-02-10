package handlers

import (
	"github.com/gofiber/fiber/v2"
)

// SSEHandler handles Server-Sent Events streaming for real-time job progress
type SSEHandler struct{}

// NewSSEHandler creates a new SSEHandler instance
func NewSSEHandler() *SSEHandler {
	return &SSEHandler{}
}

// Stream establishes an SSE connection for real-time job progress updates.
// The client receives progress events as the analysis pipeline runs through
// its stages: intake → AI prescan → tool execution → AI proposal → finalization.
//
// GET /api/v1/jobs/:id/stream
// Headers set:
//   - Content-Type: text/event-stream
//   - Cache-Control: no-cache
//   - Connection: keep-alive
//   - X-Accel-Buffering: no (disables nginx buffering)
//
// SSE Events:
//   - event: connected       — initial connection acknowledged
//   - event: progress        — stage update with percentage (0.0-1.0)
//   - event: complete        — job finished, includes report_url and score
//   - (keepalive comments every 15s to prevent proxy timeouts)
//
// Client usage:
//
//	const es = new EventSource('/api/v1/jobs/job_abc123/stream');
//	es.addEventListener('progress', (e) => { ... });
func (h *SSEHandler) Stream(c *fiber.Ctx) error {
	// TODO: Extract job ID from params
	// TODO: Verify job belongs to authenticated user
	// TODO: Set SSE headers (Content-Type: text/event-stream, etc.)
	// TODO: Subscribe to Redis pub/sub channel for this job (job:progress:<id>)
	// TODO: Stream progress events using c.Context().SetBodyStreamWriter
	// TODO: Send keepalive comments every 15s
	// TODO: Close stream on job completion or client disconnect
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"message": "SSE stream endpoint not yet implemented",
	})
}
