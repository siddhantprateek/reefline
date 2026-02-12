package tools

import (
	"context"
	"fmt"

	"github.com/siddhantprateek/reefline/pkg/tools"
)

// DiveAnalyzerStatus returns the current status of the dive analyzer
func DiveAnalyzerStatus() ScannerStatus {
	if tools.DiveAnalyzer == nil {
		return ScannerStatus{
			Available:   false,
			Initialized: false,
		}
	}

	return ScannerStatus{
		Available:   true,
		Initialized: tools.DiveAnalyzer.IsInitialized(),
	}
}

// DiveAnalyzeImage analyzes a Docker image for layer efficiency and wasted space
func DiveAnalyzeImage(ctx context.Context, image string) (*tools.DiveAnalysis, error) {
	if tools.DiveAnalyzer == nil {
		return nil, fmt.Errorf("dive analyzer not initialized")
	}

	// Check cache first
	if analysis, ok := tools.DiveAnalyzer.GetAnalysis(image); ok {
		return analysis, nil
	}

	return tools.DiveAnalyzer.AnalyzeImage(ctx, image)
}

// DiveAnalyzeArchive analyzes a container image from a local tar archive
func DiveAnalyzeArchive(ctx context.Context, archivePath string) (*tools.DiveAnalysis, error) {
	if tools.DiveAnalyzer == nil {
		return nil, fmt.Errorf("dive analyzer not initialized")
	}

	return tools.DiveAnalyzer.AnalyzeImageFromArchive(ctx, archivePath)
}

// GetDiveAnalysisResults retrieves cached analysis results for a specific image
func GetDiveAnalysisResults(image string) (*tools.DiveAnalysis, error) {
	if tools.DiveAnalyzer == nil {
		return nil, fmt.Errorf("dive analyzer not initialized")
	}

	analysis, found := tools.DiveAnalyzer.GetAnalysis(image)
	if !found {
		return &tools.DiveAnalysis{
			Image:  image,
			Status: "not_found",
		}, nil
	}

	return analysis, nil
}
