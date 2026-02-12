package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/minio/minio-go/v7"
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

	// Update Job status to RUNNING?
	// (Ideally we would update DB here, but let's focus on the tool execution as requested)

	bucket := storage.GetConfigFromEnv().DefaultBucket

	// 1. Run Grype Scan
	if tools.ImgScanner != nil && tools.ImgScanner.IsEnabled() {
		log.Printf("[Worker] Running Grype scan for %s...", target)
		scanResult, err := tools.ImgScanner.ScanImage(ctx, target)
		if err != nil {
			log.Printf("[Worker] Grype scan failed: %v", err)
			// Decide if fatal or continue? Let's continue.
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
			} else {
				log.Printf("[Worker] Uploaded grype.json to %s/%s", bucket, objectName)
			}
		}
	}

	// 2. Run Dockle Scan
	if tools.DockleScn != nil && tools.DockleScn.IsEnabled() {
		log.Printf("[Worker] Running Dockle scan for %s...", target)
		dockleResult, err := tools.DockleScn.ScanImage(ctx, target)
		if err != nil {
			log.Printf("[Worker] Dockle scan failed: %v", err)
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
			} else {
				log.Printf("[Worker] Uploaded dockle.json to %s/%s", bucket, objectName)
			}
		}
	}

	// 3. Run Dive Analysis
	if tools.DiveAnalyzer != nil && tools.DiveAnalyzer.IsEnabled() {
		log.Printf("[Worker] Running Dive analysis for %s...", target)
		diveResult, err := tools.DiveAnalyzer.AnalyzeImage(ctx, target)
		if err != nil {
			log.Printf("[Worker] Dive analysis failed: %v", err)
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
			} else {
				log.Printf("[Worker] Uploaded dive.json to %s/%s", bucket, objectName)
			}
		}
	}

	// 4. (Future) LLM Analysis triggers here?

	log.Printf("[Worker] Finished analysis job %s for: %s", data.JobID, target)
	return nil
}
