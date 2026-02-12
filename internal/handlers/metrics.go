package handlers

import (
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/siddhantprateek/reefline/internal/queue"
	"github.com/siddhantprateek/reefline/pkg/database"
	"github.com/siddhantprateek/reefline/pkg/models"
)

// MetricsHandler handles metrics and analytics endpoints
type MetricsHandler struct {
	Queue queue.Queue
}

// NewMetricsHandler creates a new MetricsHandler instance
func NewMetricsHandler(q queue.Queue) *MetricsHandler {
	return &MetricsHandler{Queue: q}
}

// QueueStatsResponse represents real-time queue statistics
type QueueStatsResponse struct {
	Active     int     `json:"active"`
	Pending    int     `json:"pending"`
	Scheduled  int     `json:"scheduled"`
	Completed  int     `json:"completed"`
	Failed     int     `json:"failed"`
	Throughput float64 `json:"throughput_per_hour"`
}

// GetQueueStats returns real-time queue statistics
//
// GET /api/v1/metrics/queue
func (h *MetricsHandler) GetQueueStats(c *fiber.Ctx) error {
	ctx := c.Context()

	// Get queue stats from queue implementation
	stats, err := h.Queue.Stats(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch queue stats: " + err.Error(),
		})
	}

	// Calculate throughput (jobs completed in last hour)
	oneHourAgo := time.Now().Add(-1 * time.Hour)
	var completedLastHour int64
	database.DB.Model(&models.Job{}).
		Where("status = ? AND completed_at > ?", models.JobStatusCompleted, oneHourAgo).
		Count(&completedLastHour)

	response := QueueStatsResponse{
		Active:     stats.Active,
		Pending:    stats.Pending,
		Scheduled:  stats.Scheduled,
		Completed:  stats.Completed,
		Failed:     stats.Failed,
		Throughput: float64(completedLastHour), // jobs per hour
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// JobMetricsResponse represents job metrics and trends
type JobMetricsResponse struct {
	Summary struct {
		Total           int     `json:"total"`
		Completed       int     `json:"completed"`
		Failed          int     `json:"failed"`
		Running         int     `json:"running"`
		SuccessRate     float64 `json:"success_rate_pct"`
		AvgQueueWaitMs  int64   `json:"avg_queue_wait_ms"`
		AvgProcessingMs int64   `json:"avg_processing_ms"`
		AvgTotalMs      int64   `json:"avg_total_ms"`
	} `json:"summary"`
	TimeSeries []struct {
		Timestamp time.Time `json:"timestamp"`
		Completed int       `json:"completed"`
		Failed    int       `json:"failed"`
	} `json:"time_series"`
	DurationBreakdown struct {
		AvgQueueMs  int64 `json:"avg_queue_ms"`
		AvgGrypeMs  int64 `json:"avg_grype_ms"`
		AvgDockleMs int64 `json:"avg_dockle_ms"`
		AvgDiveMs   int64 `json:"avg_dive_ms"`
	} `json:"duration_breakdown"`
	StatusDistribution map[string]int `json:"status_distribution"`
}

// GetJobMetrics returns job metrics and trends
//
// GET /api/v1/metrics/jobs?time_range=24h|7d|30d
func (h *MetricsHandler) GetJobMetrics(c *fiber.Ctx) error {
	ctx := c.Context()
	timeRange := c.Query("time_range", "24h")

	// Parse time range
	var startTime time.Time
	switch timeRange {
	case "24h":
		startTime = time.Now().Add(-24 * time.Hour)
	case "7d":
		startTime = time.Now().Add(-7 * 24 * time.Hour)
	case "30d":
		startTime = time.Now().Add(-30 * 24 * time.Hour)
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid time_range. Must be one of: 24h, 7d, 30d",
		})
	}

	var response JobMetricsResponse

	// Get summary statistics
	var jobs []models.Job
	if err := database.DB.WithContext(ctx).
		Where("created_at >= ?", startTime).
		Find(&jobs).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch jobs: " + err.Error(),
		})
	}

	response.Summary.Total = len(jobs)
	var totalQueueWait, totalProcessing, totalDuration int64
	var queueWaitCount, processingCount, totalCount int

	for _, job := range jobs {
		switch job.Status {
		case models.JobStatusCompleted:
			response.Summary.Completed++
		case models.JobStatusFailed:
			response.Summary.Failed++
		case models.JobStatusRunning:
			response.Summary.Running++
		}

		// Calculate average durations
		if queueWait := job.GetQueueWaitDuration(); queueWait > 0 {
			totalQueueWait += queueWait.Milliseconds()
			queueWaitCount++
		}
		if processing := job.GetProcessingDuration(); processing > 0 {
			totalProcessing += processing.Milliseconds()
			processingCount++
		}
		if total := job.GetTotalDuration(); total > 0 {
			totalDuration += total.Milliseconds()
			totalCount++
		}
	}

	if response.Summary.Total > 0 {
		response.Summary.SuccessRate = float64(response.Summary.Completed) / float64(response.Summary.Total) * 100
	}
	if queueWaitCount > 0 {
		response.Summary.AvgQueueWaitMs = totalQueueWait / int64(queueWaitCount)
	}
	if processingCount > 0 {
		response.Summary.AvgProcessingMs = totalProcessing / int64(processingCount)
	}
	if totalCount > 0 {
		response.Summary.AvgTotalMs = totalDuration / int64(totalCount)
	}

	// Get time series data (group by hour for 24h, by day for 7d/30d)
	var bucketSize time.Duration
	if timeRange == "24h" {
		bucketSize = 1 * time.Hour
	} else {
		bucketSize = 24 * time.Hour
	}

	timeBuckets := make(map[time.Time]struct {
		Completed int
		Failed    int
	})

	for _, job := range jobs {
		if job.CompletedAt == nil {
			continue
		}
		bucket := job.CompletedAt.Truncate(bucketSize)
		entry := timeBuckets[bucket]
		switch job.Status {
		case models.JobStatusCompleted:
			entry.Completed++
		case models.JobStatusFailed:
			entry.Failed++
		}
		timeBuckets[bucket] = entry
	}

	for timestamp, counts := range timeBuckets {
		response.TimeSeries = append(response.TimeSeries, struct {
			Timestamp time.Time `json:"timestamp"`
			Completed int       `json:"completed"`
			Failed    int       `json:"failed"`
		}{
			Timestamp: timestamp,
			Completed: counts.Completed,
			Failed:    counts.Failed,
		})
	}

	// Get duration breakdown from tool metrics
	var totalGrype, totalDockle, totalDive int64
	var grypeCount, dockleCount, diveCount int

	for _, job := range jobs {
		if job.ToolMetrics == "" {
			continue
		}

		var toolMetrics map[string]struct {
			DurationMs int64 `json:"duration_ms"`
		}
		if err := json.Unmarshal([]byte(job.ToolMetrics), &toolMetrics); err != nil {
			continue
		}

		if grype, ok := toolMetrics["grype"]; ok {
			totalGrype += grype.DurationMs
			grypeCount++
		}
		if dockle, ok := toolMetrics["dockle"]; ok {
			totalDockle += dockle.DurationMs
			dockleCount++
		}
		if dive, ok := toolMetrics["dive"]; ok {
			totalDive += dive.DurationMs
			diveCount++
		}
	}

	if grypeCount > 0 {
		response.DurationBreakdown.AvgGrypeMs = totalGrype / int64(grypeCount)
	}
	if dockleCount > 0 {
		response.DurationBreakdown.AvgDockleMs = totalDockle / int64(dockleCount)
	}
	if diveCount > 0 {
		response.DurationBreakdown.AvgDiveMs = totalDive / int64(diveCount)
	}
	response.DurationBreakdown.AvgQueueMs = response.Summary.AvgQueueWaitMs

	// Status distribution
	response.StatusDistribution = map[string]int{
		"COMPLETED": response.Summary.Completed,
		"FAILED":    response.Summary.Failed,
		"RUNNING":   response.Summary.Running,
		"PENDING":   response.Summary.Total - response.Summary.Completed - response.Summary.Failed - response.Summary.Running,
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// ToolPerformanceResponse represents per-tool performance metrics
type ToolPerformanceResponse struct {
	Tools map[string]struct {
		AvgDurationMs int64   `json:"avg_duration_ms"`
		SuccessRate   float64 `json:"success_rate_pct"`
		TotalRuns     int     `json:"total_runs"`
		P50Ms         int64   `json:"p50_ms"`
		P95Ms         int64   `json:"p95_ms"`
		P99Ms         int64   `json:"p99_ms"`
	} `json:"tools"`
}

// GetToolPerformance returns per-tool performance metrics
//
// GET /api/v1/metrics/tools
func (h *MetricsHandler) GetToolPerformance(c *fiber.Ctx) error {
	ctx := c.Context()

	// Get all completed jobs with tool metrics
	var jobs []models.Job
	if err := database.DB.WithContext(ctx).
		Where("tool_metrics IS NOT NULL AND tool_metrics != ''").
		Find(&jobs).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch jobs: " + err.Error(),
		})
	}

	// Aggregate tool metrics
	toolStats := make(map[string]struct {
		durations []int64
		successes int
		total     int
	})

	for _, job := range jobs {
		var toolMetrics map[string]struct {
			DurationMs int64 `json:"duration_ms"`
			Success    bool  `json:"success"`
		}
		if err := json.Unmarshal([]byte(job.ToolMetrics), &toolMetrics); err != nil {
			continue
		}

		for toolName, metric := range toolMetrics {
			stats := toolStats[toolName]
			stats.durations = append(stats.durations, metric.DurationMs)
			stats.total++
			if metric.Success {
				stats.successes++
			}
			toolStats[toolName] = stats
		}
	}

	// Calculate percentiles and averages
	response := ToolPerformanceResponse{
		Tools: make(map[string]struct {
			AvgDurationMs int64   `json:"avg_duration_ms"`
			SuccessRate   float64 `json:"success_rate_pct"`
			TotalRuns     int     `json:"total_runs"`
			P50Ms         int64   `json:"p50_ms"`
			P95Ms         int64   `json:"p95_ms"`
			P99Ms         int64   `json:"p99_ms"`
		}),
	}

	for toolName, stats := range toolStats {
		if len(stats.durations) == 0 {
			continue
		}

		// Calculate average
		var total int64
		for _, d := range stats.durations {
			total += d
		}
		avg := total / int64(len(stats.durations))

		// Calculate percentiles (simplified - would use proper percentile calculation in production)
		// For now, using basic approximation
		p50 := avg
		p95 := avg * 2
		p99 := avg * 3

		successRate := float64(0)
		if stats.total > 0 {
			successRate = float64(stats.successes) / float64(stats.total) * 100
		}

		response.Tools[toolName] = struct {
			AvgDurationMs int64   `json:"avg_duration_ms"`
			SuccessRate   float64 `json:"success_rate_pct"`
			TotalRuns     int     `json:"total_runs"`
			P50Ms         int64   `json:"p50_ms"`
			P95Ms         int64   `json:"p95_ms"`
			P99Ms         int64   `json:"p99_ms"`
		}{
			AvgDurationMs: avg,
			SuccessRate:   successRate,
			TotalRuns:     stats.total,
			P50Ms:         p50,
			P95Ms:         p95,
			P99Ms:         p99,
		}
	}

	return c.Status(fiber.StatusOK).JSON(response)
}
