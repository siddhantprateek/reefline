package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
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

	// Update Job status to RUNNING
	if err := database.DB.Model(&models.Job{}).Where("job_id = ?", data.JobID).Updates(map[string]interface{}{
		"status":   models.JobStatusRunning,
		"progress": 0,
	}).Error; err != nil {
		log.Printf("[Worker] Failed to update job status to RUNNING: %v", err)
	}

	bucket := storage.GetConfigFromEnv().DefaultBucket
	hasErrors := false

	// 1. Run Grype Scan
	if tools.ImgScanner != nil && tools.ImgScanner.IsEnabled() {
		log.Printf("[Worker] Running Grype scan for %s...", target)
		database.DB.Model(&models.Job{}).Where("job_id = ?", data.JobID).Update("progress", 10)

		scanResult, err := tools.ImgScanner.ScanImage(ctx, target)
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

		dockleResult, err := tools.DockleScn.ScanImage(ctx, target)
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

		diveResult, err := tools.DiveAnalyzer.AnalyzeImage(ctx, target)
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

	// 4. (Future) LLM Analysis triggers here?

	// Update final job status
	completedAt := time.Now()
	finalStatus := models.JobStatusCompleted
	if hasErrors {
		finalStatus = models.JobStatusFailed
	}

	if err := database.DB.Model(&models.Job{}).Where("job_id = ?", data.JobID).Updates(map[string]interface{}{
		"status":       finalStatus,
		"progress":     100,
		"completed_at": completedAt,
	}).Error; err != nil {
		log.Printf("[Worker] Failed to update job final status: %v", err)
	}

	log.Printf("[Worker] Finished analysis job %s for: %s with status: %s", data.JobID, target, finalStatus)
	return nil
}
