package tools

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/wagoodman/dive/dive"
	"github.com/wagoodman/dive/dive/image"
)

const (
	diveAnalysisTimeout = 5 * time.Minute
)

// Global dive analyzer instance
var DiveAnalyzer *diveAnalyzer

// DiveConfig holds configuration for the dive analyzer
type DiveConfig struct {
	Enable       bool          `json:"enable"`
	Timeout      time.Duration `json:"timeout"`
	Source       string        `json:"source"` // "docker", "podman", "docker-archive"
	DockerHost   string        `json:"dockerHost,omitempty"`
	IgnoreErrors bool          `json:"ignoreErrors"`
}

// diveAnalyzer wraps dive's image analysis functionality
type diveAnalyzer struct {
	mx          sync.RWMutex
	initialized bool
	config      DiveConfig
	scans       map[string]*DiveAnalysis
	log         *slog.Logger
}

// DiveAnalysis holds results from a single image analysis
type DiveAnalysis struct {
	Image             string             `json:"image"`
	Layers            []DiveLayer        `json:"layers"`
	Efficiency        float64            `json:"efficiency"`        // 0-100%
	SizeBytes         uint64             `json:"sizeBytes"`         // Total image size
	UserSizeBytes     uint64             `json:"userSizeBytes"`     // User-added layers size
	WastedBytes       uint64             `json:"wastedBytes"`       // Total wasted space
	WastedUserPercent float64            `json:"wastedUserPercent"` // % of wasted in user layers
	Inefficiencies    []DiveInefficiency `json:"inefficiencies"`
	AnalysisTime      time.Time          `json:"analysisTime"`
	Status            string             `json:"status"` // "completed", "error"
	Error             string             `json:"error,omitempty"`
}

// DiveLayer represents a single image layer
type DiveLayer struct {
	Index     int    `json:"index"`
	ID        string `json:"id"`
	DigestID  string `json:"digestId"`
	Command   string `json:"command"`
	SizeBytes uint64 `json:"sizeBytes"`
	FileCount int    `json:"fileCount"`
}

// DiveInefficiency represents files contributing to wasted space
type DiveInefficiency struct {
	Path              string `json:"path"`
	SizeBytes         uint64 `json:"sizeBytes"`
	RemovedOperations int    `json:"removedOperations"` // How many times this file was added/removed
}

// NewDiveAnalyzer creates a new dive analyzer
func NewDiveAnalyzer(cfg DiveConfig, l *slog.Logger) *diveAnalyzer {
	if cfg.Timeout == 0 {
		cfg.Timeout = diveAnalysisTimeout
	}
	if cfg.Source == "" {
		cfg.Source = "docker"
	}
	return &diveAnalyzer{
		config: cfg,
		scans:  make(map[string]*DiveAnalysis),
		log:    l.With("subsys", "dive"),
	}
}

// Init marks the analyzer as initialized
func (a *diveAnalyzer) Init() {
	a.mx.Lock()
	defer a.mx.Unlock()

	// Set Docker host if configured
	if a.config.DockerHost != "" {
		// Docker client will read from DOCKER_HOST env var
		// We could set it here if needed, but typically it's already set
		a.log.Info("Using configured Docker host", "host", a.config.DockerHost)
	}

	a.initialized = true
	a.log.Info("Dive analyzer initialized", "source", a.config.Source)
}

// IsEnabled returns whether the analyzer is enabled
func (a *diveAnalyzer) IsEnabled() bool {
	return a.config.Enable
}

// IsInitialized returns whether the analyzer has been initialized
func (a *diveAnalyzer) IsInitialized() bool {
	a.mx.RLock()
	defer a.mx.RUnlock()
	return a.initialized
}

// GetAnalysis retrieves a cached analysis result
func (a *diveAnalyzer) GetAnalysis(img string) (*DiveAnalysis, bool) {
	a.mx.RLock()
	defer a.mx.RUnlock()
	analysis, ok := a.scans[img]
	return analysis, ok
}

func (a *diveAnalyzer) setAnalysis(img string, analysis *DiveAnalysis) {
	a.mx.Lock()
	defer a.mx.Unlock()
	a.scans[img] = analysis
}

// AnalyzeImage analyzes a Docker image from the Docker daemon
func (a *diveAnalyzer) AnalyzeImage(ctx context.Context, imageName string) (*DiveAnalysis, error) {
	if !a.IsInitialized() {
		return nil, fmt.Errorf("dive analyzer not initialized")
	}
	if imageName == "" {
		return nil, fmt.Errorf("image name is required")
	}

	// Check cache first
	if analysis, ok := a.GetAnalysis(imageName); ok {
		a.log.Info("Returning cached dive analysis", "image", imageName)
		return analysis, nil
	}

	// Parse image source
	source := dive.ParseImageSource(a.config.Source)

	return a.doAnalyze(ctx, imageName, "", source)
}

// AnalyzeImageFromArchive analyzes an image from a tar archive
func (a *diveAnalyzer) AnalyzeImageFromArchive(ctx context.Context, archivePath string) (*DiveAnalysis, error) {
	if !a.IsInitialized() {
		return nil, fmt.Errorf("dive analyzer not initialized")
	}
	if archivePath == "" {
		return nil, fmt.Errorf("archive path is required")
	}

	return a.doAnalyze(ctx, "", archivePath, dive.SourceDockerArchive)
}

func (a *diveAnalyzer) doAnalyze(ctx context.Context, imageName, archivePath string, source dive.ImageSource) (analysis *DiveAnalysis, err error) {
	scanID := imageName
	if scanID == "" {
		scanID = archivePath
	}

	start := time.Now()
	a.log.Info("Starting dive analysis", "image", scanID, "source", source)

	// Handle panics gracefully
	defer func() {
		if r := recover(); r != nil {
			recErr := fmt.Errorf("panic in dive analysis: %v", r)
			a.log.Error("Recovered from panic in dive analysis",
				"image", scanID,
				"error", recErr,
			)
			analysis = &DiveAnalysis{
				Image:        scanID,
				AnalysisTime: time.Now(),
				Status:       "error",
				Error:        recErr.Error(),
			}
			a.setAnalysis(scanID, analysis)
			err = nil // Return the error analysis object instead of error
		}
	}()

	// Add timeout to context
	ctx, cancel := context.WithTimeout(ctx, a.config.Timeout)
	defer cancel()

	// Get image resolver based on source
	resolver, err := dive.GetImageResolver(source)
	if err != nil {
		analysis = &DiveAnalysis{
			Image:        scanID,
			AnalysisTime: time.Now(),
			Status:       "error",
			Error:        fmt.Sprintf("failed to get image resolver: %v", err),
		}
		a.setAnalysis(scanID, analysis)
		return analysis, fmt.Errorf("failed to get image resolver for %s: %w", scanID, err)
	}

	// Fetch the image
	var img *image.Image
	if archivePath != "" {
		// For archive, fetch using the archive path
		img, err = resolver.Fetch(archivePath)
	} else {
		// For Docker daemon, fetch using image name
		img, err = resolver.Fetch(imageName)
	}

	if err != nil {
		analysis = &DiveAnalysis{
			Image:        scanID,
			AnalysisTime: time.Now(),
			Status:       "error",
			Error:        fmt.Sprintf("failed to fetch image: %v", err),
		}
		a.setAnalysis(scanID, analysis)
		return analysis, fmt.Errorf("failed to fetch image %s: %w", scanID, err)
	}

	// Analyze the image
	imageAnalysis, err := img.Analyze()
	if err != nil {
		analysis = &DiveAnalysis{
			Image:        scanID,
			AnalysisTime: time.Now(),
			Status:       "error",
			Error:        fmt.Sprintf("failed to analyze image: %v", err),
		}
		a.setAnalysis(scanID, analysis)
		return analysis, fmt.Errorf("failed to analyze image %s: %w", scanID, err)
	}

	// Convert to our result format
	analysis = convertImageAnalysis(scanID, imageAnalysis)
	a.setAnalysis(scanID, analysis)

	a.log.Info("Dive analysis completed",
		"image", scanID,
		"efficiency", fmt.Sprintf("%.2f%%", analysis.Efficiency),
		"wasted_mb", analysis.WastedBytes/1024/1024,
		"layers", len(analysis.Layers),
		"elapsed", time.Since(start),
	)

	return analysis, nil
}

// Stop performs cleanup
func (a *diveAnalyzer) Stop() {
	a.mx.Lock()
	defer a.mx.Unlock()

	if !a.initialized {
		return
	}

	a.log.Info("Stopping dive analyzer")

	// Clear cached scans
	a.scans = nil
	a.initialized = false

	a.log.Info("Dive analyzer stopped")
}

// convertImageAnalysis converts dive's image.AnalysisResult to our DiveAnalysis format
func convertImageAnalysis(imageName string, analysis *image.AnalysisResult) *DiveAnalysis {
	// Convert layers
	layers := make([]DiveLayer, 0, len(analysis.Layers))
	for _, layer := range analysis.Layers {
		fileCount := 0
		if layer.Tree != nil {
			fileCount = layer.Tree.Size
		}

		layers = append(layers, DiveLayer{
			Index:     layer.Index,
			ID:        layer.Id,
			DigestID:  layer.Digest,
			Command:   layer.Command,
			SizeBytes: layer.Size,
			FileCount: fileCount,
		})
	}

	// Convert inefficiencies
	inefficiencies := make([]DiveInefficiency, 0, len(analysis.Inefficiencies))
	for _, ineff := range analysis.Inefficiencies {
		// Note: EfficiencyData doesn't have RemovedOperations in v0.13.1
		// We use the number of nodes as a proxy for how many times the file appears
		inefficiencies = append(inefficiencies, DiveInefficiency{
			Path:              ineff.Path,
			SizeBytes:         uint64(ineff.CumulativeSize),
			RemovedOperations: len(ineff.Nodes),
		})
	}

	return &DiveAnalysis{
		Image:             imageName,
		Layers:            layers,
		Efficiency:        analysis.Efficiency,
		SizeBytes:         analysis.SizeBytes,
		UserSizeBytes:     analysis.UserSizeByes, // Note: typo in dive v0.13.1 struct
		WastedBytes:       analysis.WastedBytes,
		WastedUserPercent: analysis.WastedUserPercent,
		Inefficiencies:    inefficiencies,
		AnalysisTime:      time.Now(),
		Status:            "completed",
	}
}
