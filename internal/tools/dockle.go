package tools

import (
	"context"
	"fmt"

	"github.com/siddhantprateek/reefline/pkg/tools"
)

// DockleScannerStatus returns the current status of the dockle scanner
func DockleScannerStatus() ScannerStatus {
	if tools.DockleScn == nil {
		return ScannerStatus{
			Available:   false,
			Initialized: false,
		}
	}

	return ScannerStatus{
		Available:   true,
		Initialized: tools.DockleScn.IsInitialized(),
	}
}

// DockleScanImage scans a single container image for CIS Docker Benchmark compliance
func DockleScanImage(ctx context.Context, image string) (*tools.DockleScan, error) {
	if tools.DockleScn == nil {
		return nil, fmt.Errorf("dockle scanner not initialized")
	}

	// Check cache first
	if scan, ok := tools.DockleScn.GetScan(image); ok {
		return scan, nil
	}

	return tools.DockleScn.ScanImage(ctx, image)
}

// DockleScanImageFromFile scans a container image from a local tar archive
func DockleScanImageFromFile(ctx context.Context, filePath string) (*tools.DockleScan, error) {
	if tools.DockleScn == nil {
		return nil, fmt.Errorf("dockle scanner not initialized")
	}

	return tools.DockleScn.ScanImageFromFile(ctx, filePath)
}

// GetDockleScanResults retrieves cached scan results for a specific image
func GetDockleScanResults(image string) (*tools.DockleScan, error) {
	if tools.DockleScn == nil {
		return nil, fmt.Errorf("dockle scanner not initialized")
	}

	scan, found := tools.DockleScn.GetScan(image)
	if !found {
		return &tools.DockleScan{
			Image:  image,
			Status: "not_found",
		}, nil
	}

	return scan, nil
}
