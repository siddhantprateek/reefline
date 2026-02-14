package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/siddhantprateek/reefline/pkg/database"
	"github.com/siddhantprateek/reefline/pkg/models"
	"github.com/siddhantprateek/reefline/pkg/storage"
	"github.com/siddhantprateek/reefline/pkg/tools"
)

type AnalyzeJobPayload struct {
	JobID      string      `json:"job_id"`
	Dockerfile string      `json:"dockerfile"`
	ImageRef   string      `json:"image_ref"`
	AppContext string      `json:"app_context"`
	SkopeoMeta interface{} `json:"skopeo_meta,omitempty"` // Keep as interface{} to avoid circular dep if tools not wanted here, or use tools.InspectResult
}

// ProcessAnalyzeJob handles the image analysis workflow
func ProcessAnalyzeJob(ctx context.Context, payload []byte) error {
	var data AnalyzeJobPayload
	if err := json.Unmarshal(payload, &data); err != nil {
		log.Printf("[Worker] Error unmarshalling analysis payload: %v", err)
		return err
	}

	if data.JobID == "" {
		log.Printf("[Worker] Job ID missing in payload")
		return fmt.Errorf("job_id is required")
	}

	target := data.ImageRef
	// If only Dockerfile, we might need to build it first?
	// For now, let's assume image_ref is present or we skip tool execution if empty.
	if target == "" {
		log.Printf("[Worker] No image ref provided, skipping automated analysis for job %s", data.JobID)
		return nil
	}

	log.Printf("[Worker] Starting analysis job %s for image: %s", data.JobID, target)

	// Update Job status to RUNNING and set StartedAt timestamp
	startedAt := time.Now()
	if err := database.DB.Model(&models.Job{}).Where("job_id = ?", data.JobID).Updates(map[string]interface{}{
		"status":     models.JobStatusRunning,
		"progress":   0,
		"started_at": startedAt,
	}).Error; err != nil {
		log.Printf("[Worker] Failed to update job status to RUNNING: %v", err)
	}

	bucket := storage.GetConfigFromEnv().DefaultBucket
	hasErrors := false

	// Initialize tool metrics map
	type ToolMetric struct {
		StartedAt   string `json:"started_at"`
		CompletedAt string `json:"completed_at"`
		DurationMs  int64  `json:"duration_ms"`
		Success     bool   `json:"success"`
		Error       string `json:"error,omitempty"`
	}
	toolMetrics := make(map[string]ToolMetric)

	// 1. Run Grype Scan
	if tools.ImgScanner != nil && tools.ImgScanner.IsEnabled() {
		log.Printf("[Worker] Running Grype scan for %s...", target)
		database.DB.Model(&models.Job{}).Where("job_id = ?", data.JobID).Update("progress", 10)

		grypeStart := time.Now()
		scanResult, err := tools.ImgScanner.ScanImage(ctx, target)
		grypeEnd := time.Now()
		grypeDuration := grypeEnd.Sub(grypeStart)

		toolMetrics["grype"] = ToolMetric{
			StartedAt:   grypeStart.Format(time.RFC3339),
			CompletedAt: grypeEnd.Format(time.RFC3339),
			DurationMs:  grypeDuration.Milliseconds(),
			Success:     err == nil,
			Error: func() string {
				if err != nil {
					return err.Error()
				}
				return ""
			}(),
		}

		if err != nil {
			log.Printf("[Worker] Grype scan failed: %v", err)
			hasErrors = true
		} else {
			// Upload Grype result (processed in next step or by LLM)
			resultJSON, _ := json.Marshal(scanResult)
			reader := bytes.NewReader(resultJSON)
			objectName := fmt.Sprintf("%s/artifacts/grype.json", data.JobID)

			_, err := storage.Client.PutObject(ctx, bucket, objectName, reader, int64(len(resultJSON)), minio.PutObjectOptions{
				ContentType: "application/json",
			})
			if err != nil {
				log.Printf("[Worker] Failed to upload grype.json: %v", err)
				hasErrors = true
			} else {
				log.Printf("[Worker] Uploaded grype.json to %s/%s", bucket, objectName)
			}
		}
		database.DB.Model(&models.Job{}).Where("job_id = ?", data.JobID).Update("progress", 35)
	}

	// 2. Run Dockle Scan
	if tools.DockleScn != nil && tools.DockleScn.IsEnabled() {
		log.Printf("[Worker] Running Dockle scan for %s...", target)
		database.DB.Model(&models.Job{}).Where("job_id = ?", data.JobID).Update("progress", 40)

		dockleStart := time.Now()
		dockleResult, err := tools.DockleScn.ScanImage(ctx, target)
		dockleEnd := time.Now()
		dockleDuration := dockleEnd.Sub(dockleStart)

		toolMetrics["dockle"] = ToolMetric{
			StartedAt:   dockleStart.Format(time.RFC3339),
			CompletedAt: dockleEnd.Format(time.RFC3339),
			DurationMs:  dockleDuration.Milliseconds(),
			Success:     err == nil,
			Error: func() string {
				if err != nil {
					return err.Error()
				}
				return ""
			}(),
		}

		if err != nil {
			log.Printf("[Worker] Dockle scan failed: %v", err)
			hasErrors = true
		} else {
			// Upload Dockle result
			resultJSON, _ := json.Marshal(dockleResult)
			reader := bytes.NewReader(resultJSON)
			objectName := fmt.Sprintf("%s/artifacts/dockle.json", data.JobID)

			_, err := storage.Client.PutObject(ctx, bucket, objectName, reader, int64(len(resultJSON)), minio.PutObjectOptions{
				ContentType: "application/json",
			})
			if err != nil {
				log.Printf("[Worker] Failed to upload dockle.json: %v", err)
				hasErrors = true
			} else {
				log.Printf("[Worker] Uploaded dockle.json to %s/%s", bucket, objectName)
			}
		}
		database.DB.Model(&models.Job{}).Where("job_id = ?", data.JobID).Update("progress", 65)
	}

	// 3. Run Dive Analysis
	if tools.DiveAnalyzer != nil && tools.DiveAnalyzer.IsEnabled() {
		log.Printf("[Worker] Running Dive analysis for %s...", target)
		database.DB.Model(&models.Job{}).Where("job_id = ?", data.JobID).Update("progress", 70)

		diveStart := time.Now()
		diveResult, err := tools.DiveAnalyzer.AnalyzeImage(ctx, target)
		diveEnd := time.Now()
		diveDuration := diveEnd.Sub(diveStart)

		toolMetrics["dive"] = ToolMetric{
			StartedAt:   diveStart.Format(time.RFC3339),
			CompletedAt: diveEnd.Format(time.RFC3339),
			DurationMs:  diveDuration.Milliseconds(),
			Success:     err == nil,
			Error: func() string {
				if err != nil {
					return err.Error()
				}
				return ""
			}(),
		}

		if err != nil {
			log.Printf("[Worker] Dive analysis failed: %v", err)
			hasErrors = true
		} else {
			// Upload Dive result
			resultJSON, _ := json.Marshal(diveResult)
			reader := bytes.NewReader(resultJSON)
			objectName := fmt.Sprintf("%s/artifacts/dive.json", data.JobID)

			_, err := storage.Client.PutObject(ctx, bucket, objectName, reader, int64(len(resultJSON)), minio.PutObjectOptions{
				ContentType: "application/json",
			})
			if err != nil {
				log.Printf("[Worker] Failed to upload dive.json: %v", err)
				hasErrors = true
			} else {
				log.Printf("[Worker] Uploaded dive.json to %s/%s", bucket, objectName)
			}
		}
		database.DB.Model(&models.Job{}).Where("job_id = ?", data.JobID).Update("progress", 95)
	}

	// 4. Trigger flow service to generate AI report
	flowURL := os.Getenv("FLOW_SERVICE_URL")
	if flowURL == "" {
		flowURL = "http://localhost:8000"
	}
	if err := triggerFlowReport(ctx, flowURL, data.JobID, "openai"); err != nil {
		log.Printf("[Worker] Flow report generation failed for job %s: %v", data.JobID, err)
		// Non-fatal — scans are still stored
	}

	// Update final job status
	completedAt := time.Now()
	finalStatus := models.JobStatusCompleted
	if hasErrors {
		finalStatus = models.JobStatusFailed
	}

	// Serialize tool metrics to JSON
	toolMetricsJSON, err := json.Marshal(toolMetrics)
	if err != nil {
		log.Printf("[Worker] Failed to marshal tool metrics: %v", err)
		toolMetricsJSON = []byte("{}")
	}

	if err := database.DB.Model(&models.Job{}).Where("job_id = ?", data.JobID).Updates(map[string]interface{}{
		"status":       finalStatus,
		"progress":     100,
		"completed_at": completedAt,
		"tool_metrics": string(toolMetricsJSON),
	}).Error; err != nil {
		log.Printf("[Worker] Failed to update job final status: %v", err)
	}

	log.Printf("[Worker] Finished analysis job %s for: %s with status: %s", data.JobID, target, finalStatus)
	return nil
}

// triggerFlowReport calls the Python flow service to generate an AI report for the job.
func triggerFlowReport(ctx context.Context, baseURL, jobID, provider string) error {
	body, _ := json.Marshal(map[string]string{
		"job_id":   jobID,
		"provider": provider,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/report", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode >= 300 {
		return fmt.Errorf("flow service returned %d", resp.StatusCode)
	}
	log.Printf("[Worker] Flow service generated report for job %s", jobID)
	return nil
}
