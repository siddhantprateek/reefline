package tools

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"
)

func TestDockleScan(t *testing.T) {
	// Skip if Docker is not available?
	// Ideally we want to run this to reproduce the panic.
	// Assuming the user has docker running since they are running the app.

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg := DockleConfig{
		Enable:  true,
		Timeout: 5 * time.Minute,
	}

	scanner := NewDockleScanner(cfg, logger)
	scanner.Init()

	ctx := context.Background()
	// Use a small image for testing
	imageName := "alpine:latest"

	t.Logf("Scanning image: %s", imageName)
	result, err := scanner.ScanImage(ctx, imageName)

	if err != nil {
		// If it's the panic error we wrapped, it might come returned as a result with Status="error"
		// or as a direct error if it failed before recovery (but we handle recovery).
		// In our current implementation, doScan returns (scan, nil) even on panic, with scan.Error set.
		t.Fatalf("ScanImage returned error: %v", err)
	}

	if result.Status == "error" {
		t.Fatalf("Scan failed with status error: %s", result.Error)
	}

	t.Logf("Scan completed. Fatal: %d, Warn: %d, Pass: %d",
		result.Summary.Fatal, result.Summary.Warn, result.Summary.Pass)
}
