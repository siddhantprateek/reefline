package handlers

import (
	"github.com/gofiber/fiber/v2"
)

// HealthHandler contains handlers for health check endpoints
type HealthHandler struct{}

// NewHealthHandler creates a new HealthHandler instance
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Status returns the overall health status of the service
func (h *HealthHandler) Status(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":  "ok",
		"service": "reefline-server",
	})
}

// Ready returns the readiness status of the service
func (h *HealthHandler) Ready(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":  "ready",
		"service": "reefline-server",
	})
}

// Live returns the liveness status of the service
func (h *HealthHandler) Live(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":  "alive",
		"service": "reefline-server",
	})
}
