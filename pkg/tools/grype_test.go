package tools

import (
	"context"
	"log/slog"
	"os"
	"testing"
)

func TestGrypeScan(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg := ImageScans{
		Enable: true,
	}

	scanner := NewImageScanner(cfg, logger)
	scanner.Init("reefline-test", "v0.0.1")
	defer scanner.Stop()

	// Wait a bit for DB loading if async (Init blocks? it does block for DB load)
	// Init calls LoadVulnerabilityDB which might take time if not cached.

	if !scanner.isInitialized() {
		t.Fatal("Scanner failed to initialize")
	}

	ctx := context.Background()
	imageName := "alpine:latest"

	t.Logf("Scanning image: %s", imageName)
	result, err := scanner.ScanImage(ctx, imageName)

	if err != nil {
		t.Fatalf("Grype scan failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result is nil")
	}

	t.Logf("Scan completed. Vulnerabilities found: %d", result.Tally.Total)
	t.Logf("Critical: %d, High: %d, Medium: %d, Low: %d",
		result.Tally.Critical, result.Tally.High, result.Tally.Medium, result.Tally.Low)
}
