package tools

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"
)

func TestSkopeoInspect(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg := ImageInspectorConfig{
		Enable:  true,
		Timeout: 30 * time.Second,
	}

	inspector := NewImageInspector(cfg, logger)
	inspector.Init()
	defer inspector.Stop()

	if !inspector.IsInitialized() {
		t.Fatal("Inspector failed to initialize")
	}

	ctx := context.Background()
	imageName := "alpine:latest"

	t.Logf("Inspecting image: %s", imageName)

	// Test without auth first
	result, err := inspector.InspectImage(ctx, imageName, nil)

	if err != nil {
		t.Fatalf("Skopeo inspect failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result is nil")
	}

	if result.Status != "completed" {
		t.Fatalf("Inspection status not completed: %s, error: %s", result.Status, result.Error)
	}

	t.Logf("Inspection completed.")
	t.Logf("Architecture: %s", result.Architecture)
	t.Logf("OS: %s", result.Os)
	t.Logf("Digest: %s", result.Digest)
	t.Logf("Layers: %d", len(result.Layers))

	// Verify Arch is amd64 as we forced it
	if result.Architecture != "amd64" {
		t.Errorf("Expected architecture amd64, got %s", result.Architecture)
	}
	if result.Os != "linux" {
		t.Errorf("Expected OS linux, got %s", result.Os)
	}
}
