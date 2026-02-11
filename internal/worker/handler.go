package worker

import (
	"context"
	"encoding/json"
	"log"
	"time"
)

type AnalyzeJobPayload struct {
	JobID      string      `json:"job_id"`
	Dockerfile string      `json:"dockerfile"`
	ImageRef   string      `json:"image_ref"`
	AppContext string      `json:"app_context"`
	SkopeoMeta interface{} `json:"skopeo_meta,omitempty"` // Keep as interface{} to avoid circular dep if tools not wanted here, or use tools.InspectResult
}

func ProcessAnalyzeJob(ctx context.Context, payload []byte) error {
	var data AnalyzeJobPayload
	if err := json.Unmarshal(payload, &data); err != nil {
		log.Printf("[Worker] Error unmarshalling analysis payload: %v", err)
		return err
	}

	target := data.ImageRef
	if target == "" {
		target = "Dockerfile"
	}

	log.Printf("[Worker] Starting analysis job for: %s", target)

	// Simulate Analysis Steps (as requested for testing: just delay)
	// 5s intake
	log.Println("[Worker] Stage 1: Intake & Validation...")
	select {
	case <-time.After(5 * time.Second):
	case <-ctx.Done():
		return ctx.Err()
	}

	// 10s tool execution
	log.Println("[Worker] Stage 2: Tool Execution (Simulated)...")
	select {
	case <-time.After(10 * time.Second):
	case <-ctx.Done():
		return ctx.Err()
	}

	// 5s Finalize
	log.Println("[Worker] Stage 3: Report Assembly...")
	select {
	case <-time.After(5 * time.Second):
	case <-ctx.Done():
		return ctx.Err()
	}

	log.Printf("[Worker] Finished analysis job for: %s", target)
	return nil
}
