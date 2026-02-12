package tools

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"
)

func TestDiveAnalysis(t *testing.T) {
	// Skip if Docker is not available
	// Assuming the user has docker running since they are running the app.

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg := DiveConfig{
		Enable:       true,
		Timeout:      5 * time.Minute,
		Source:       "docker",
		IgnoreErrors: true,
	}

	analyzer := NewDiveAnalyzer(cfg, logger)
	analyzer.Init()
	defer analyzer.Stop()

	if !analyzer.IsInitialized() {
		t.Fatal("Analyzer failed to initialize")
	}

	ctx := context.Background()
	// Use a small image for testing
	imageName := "alpine:latest"

	t.Logf("Analyzing image: %s", imageName)
	result, err := analyzer.AnalyzeImage(ctx, imageName)

	if err != nil {
		// If it's an error, it might be wrapped with the result object having Status="error"
		// or as a direct error. Our implementation returns (analysis, error) on failure.
		t.Fatalf("AnalyzeImage returned error: %v", err)
	}

	if result == nil {
		t.Fatal("Result is nil")
	}

	if result.Status == "error" {
		t.Fatalf("Analysis failed with status error: %s", result.Error)
	}

	t.Logf("Analysis completed successfully")
	t.Logf("Image: %s", result.Image)
	t.Logf("Efficiency: %.2f%%", result.Efficiency)
	t.Logf("Total Size: %d bytes (%.2f MB)", result.SizeBytes, float64(result.SizeBytes)/1024/1024)
	t.Logf("User Size: %d bytes (%.2f MB)", result.UserSizeBytes, float64(result.UserSizeBytes)/1024/1024)
	t.Logf("Wasted Bytes: %d bytes (%.2f MB)", result.WastedBytes, float64(result.WastedBytes)/1024/1024)
	t.Logf("Wasted User Percent: %.2f%%", result.WastedUserPercent)
	t.Logf("Layers: %d", len(result.Layers))
	t.Logf("Inefficiencies: %d", len(result.Inefficiencies))

	// Verify basic fields
	if result.Image != imageName {
		t.Errorf("Expected image name %s, got %s", imageName, result.Image)
	}

	if result.Efficiency < 0 || result.Efficiency > 100 {
		t.Errorf("Efficiency should be between 0-100, got %.2f", result.Efficiency)
	}

	if result.SizeBytes == 0 {
		t.Error("SizeBytes should not be 0")
	}

	if len(result.Layers) == 0 {
		t.Error("Expected at least one layer")
	}

	// Log layer details
	t.Logf("\n=== Layer Details ===")
	for _, layer := range result.Layers {
		layerID := layer.ID
		if len(layerID) > 12 {
			layerID = layerID[:12]
		}
		t.Logf("Layer %d: %s", layer.Index, layerID)
		t.Logf("  Command: %s", layer.Command)
		t.Logf("  Size: %d bytes (%.2f MB)", layer.SizeBytes, float64(layer.SizeBytes)/1024/1024)
		t.Logf("  Files: %d", layer.FileCount)
	}

	// Log inefficiencies if any
	if len(result.Inefficiencies) > 0 {
		t.Logf("\n=== Inefficiencies (top 5) ===")
		count := len(result.Inefficiencies)
		if count > 5 {
			count = 5
		}
		for i := 0; i < count; i++ {
			ineff := result.Inefficiencies[i]
			t.Logf("%d. %s", i+1, ineff.Path)
			t.Logf("   Size: %d bytes (%.2f KB)", ineff.SizeBytes, float64(ineff.SizeBytes)/1024)
			t.Logf("   Operations: %d", ineff.RemovedOperations)
		}
	}

	// Test cache retrieval
	t.Logf("\n=== Testing Cache ===")
	cachedResult, found := analyzer.GetAnalysis(imageName)
	if !found {
		t.Error("Expected cached result to be found")
	}
	if cachedResult.Image != result.Image {
		t.Error("Cached result doesn't match original result")
	}
	t.Logf("Cache retrieval successful")

	// Test second call uses cache
	t.Logf("\n=== Testing Cache Usage ===")
	result2, err := analyzer.AnalyzeImage(ctx, imageName)
	if err != nil {
		t.Fatalf("Second AnalyzeImage call failed: %v", err)
	}
	if result2.AnalysisTime != result.AnalysisTime {
		t.Error("Expected second call to use cache (same AnalysisTime)")
	}
	t.Logf("Cache is being used correctly")
}

func TestDiveAnalysisNonExistentImage(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg := DiveConfig{
		Enable:       true,
		Timeout:      30 * time.Second,
		Source:       "docker",
		IgnoreErrors: true,
	}

	analyzer := NewDiveAnalyzer(cfg, logger)
	analyzer.Init()
	defer analyzer.Stop()

	ctx := context.Background()
	imageName := "nonexistent-image-xyz:latest"

	t.Logf("Attempting to analyze non-existent image: %s", imageName)
	result, err := analyzer.AnalyzeImage(ctx, imageName)

	// We expect an error or a result with error status
	if err == nil && (result == nil || result.Status != "error") {
		t.Error("Expected error for non-existent image")
	}

	if result != nil && result.Status == "error" {
		t.Logf("Correctly handled non-existent image with error: %s", result.Error)
	} else if err != nil {
		t.Logf("Correctly returned error: %v", err)
	}
}
