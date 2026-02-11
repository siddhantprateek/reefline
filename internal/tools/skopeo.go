package tools

import (
	"context"
	"fmt"

	"github.com/siddhantprateek/reefline/pkg/tools"
)

// ImageInspectorStatus returns the current status of the image inspector
func ImageInspectorStatus() ScannerStatus {
	if tools.ImgInspector == nil {
		return ScannerStatus{
			Available:   false,
			Initialized: false,
		}
	}

	return ScannerStatus{
		Available:   true,
		Initialized: tools.ImgInspector.IsInitialized(),
	}
}

// InspectImage inspects a remote container image and returns its metadata
func InspectImage(ctx context.Context, imageName string, auth *tools.ImageAuth) (*tools.InspectResult, error) {
	if tools.ImgInspector == nil {
		return nil, fmt.Errorf("image inspector not initialized")
	}

	// Check cache first
	if result, ok := tools.ImgInspector.GetInspection(imageName); ok {
		return result, nil
	}

	return tools.ImgInspector.InspectImage(ctx, imageName, auth)
}

// GetRawManifest retrieves only the raw manifest bytes for an image
func GetRawManifest(ctx context.Context, imageName string, auth *tools.ImageAuth) ([]byte, string, error) {
	if tools.ImgInspector == nil {
		return nil, "", fmt.Errorf("image inspector not initialized")
	}

	return tools.ImgInspector.GetRawManifest(ctx, imageName, auth)
}

// GetCachedInspection retrieves a cached inspection result for a specific image
func GetCachedInspection(imageName string) (*tools.InspectResult, error) {
	if tools.ImgInspector == nil {
		return nil, fmt.Errorf("image inspector not initialized")
	}

	result, found := tools.ImgInspector.GetInspection(imageName)
	if !found {
		return &tools.InspectResult{
			Image:  imageName,
			Status: "not_found",
		}, nil
	}

	return result, nil
}
