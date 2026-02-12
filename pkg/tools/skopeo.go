package tools

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/containers/image/v5/docker"
	"github.com/containers/image/v5/manifest"
	"github.com/containers/image/v5/types"
	"github.com/opencontainers/go-digest"
)

const (
	inspectTimeout = 2 * time.Minute
)

// Global image inspector instance
var ImgInspector *ImageInspector

// ImageInspectorConfig holds configuration for the image inspector
type ImageInspectorConfig struct {
	Enable                bool          `json:"enable"`
	Timeout               time.Duration `json:"timeout"`
	InsecureSkipTLSVerify bool          `json:"insecureSkipTLSVerify"`
}

// ImageAuth holds per-request authentication credentials
type ImageAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// ImageInspector wraps containers/image inspect functionality
type ImageInspector struct {
	mx          sync.RWMutex
	initialized bool
	config      ImageInspectorConfig
	cache       map[string]*InspectResult
	log         *slog.Logger
}

// InspectResult holds the result of an image inspection
type InspectResult struct {
	Image         string             `json:"image"`
	Digest        string             `json:"digest"`
	MediaType     string             `json:"mediaType"`
	Architecture  string             `json:"architecture"`
	Variant       string             `json:"variant,omitempty"`
	Os            string             `json:"os"`
	Created       *time.Time         `json:"created,omitempty"`
	DockerVersion string             `json:"dockerVersion,omitempty"`
	Author        string             `json:"author,omitempty"`
	Labels        map[string]string  `json:"labels,omitempty"`
	Env           []string           `json:"env,omitempty"`
	Layers        []InspectLayerInfo `json:"layers"`
	RawManifest   []byte             `json:"rawManifest,omitempty"`
	RawConfig     []byte             `json:"rawConfig,omitempty"`
	InspectTime   time.Time          `json:"inspectTime"`
	Status        string             `json:"status"` // "completed", "error"
	Error         string             `json:"error,omitempty"`
}

// InspectLayerInfo holds metadata for a single image layer
type InspectLayerInfo struct {
	MIMEType    string            `json:"mimeType,omitempty"`
	Digest      string            `json:"digest"`
	Size        int64             `json:"size"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

// NewImageInspector creates a new image inspector
func NewImageInspector(cfg ImageInspectorConfig, l *slog.Logger) *ImageInspector {
	if cfg.Timeout == 0 {
		cfg.Timeout = inspectTimeout
	}
	return &ImageInspector{
		config: cfg,
		cache:  make(map[string]*InspectResult),
		log:    l.With("subsys", "image-inspector"),
	}
}

// Init marks the inspector as ready
func (i *ImageInspector) Init() {
	i.mx.Lock()
	defer i.mx.Unlock()
	i.initialized = true
	i.log.Info("Image inspector initialized")
}

// IsInitialized returns whether the inspector is ready
func (i *ImageInspector) IsInitialized() bool {
	i.mx.RLock()
	defer i.mx.RUnlock()
	return i.initialized
}

// IsEnabled returns whether the inspector is enabled
func (i *ImageInspector) IsEnabled() bool {
	return i.config.Enable
}

// GetInspection retrieves a cached inspection result
func (i *ImageInspector) GetInspection(img string) (*InspectResult, bool) {
	i.mx.RLock()
	defer i.mx.RUnlock()
	r, ok := i.cache[img]
	return r, ok
}

func (i *ImageInspector) setInspection(img string, r *InspectResult) {
	i.mx.Lock()
	defer i.mx.Unlock()
	i.cache[img] = r
}

// InspectImage inspects a remote container image and returns its metadata
func (i *ImageInspector) InspectImage(ctx context.Context, imageName string, auth *ImageAuth) (*InspectResult, error) {
	if !i.IsInitialized() {
		return nil, fmt.Errorf("image inspector not initialized")
	}
	if imageName == "" {
		return nil, fmt.Errorf("image name is required")
	}

	start := time.Now()
	i.log.Info("Inspecting image", "image", imageName)

	ctx, cancel := context.WithTimeout(ctx, i.config.Timeout)
	defer cancel()

	// Parse image reference
	ref, err := parseImageReference(imageName)
	if err != nil {
		result := &InspectResult{
			Image:       imageName,
			InspectTime: time.Now(),
			Status:      "error",
			Error:       err.Error(),
		}
		i.setInspection(imageName, result)
		return result, fmt.Errorf("failed to parse image reference %s: %w", imageName, err)
	}

	// Build system context with auth
	sysCtx := i.buildSystemContext(auth)

	// Create image source (for raw manifest)
	imgSrc, err := ref.NewImageSource(ctx, sysCtx)
	if err != nil {
		result := &InspectResult{
			Image:       imageName,
			InspectTime: time.Now(),
			Status:      "error",
			Error:       err.Error(),
		}
		i.setInspection(imageName, result)
		return result, fmt.Errorf("failed to create image source for %s: %w", imageName, err)
	}
	defer imgSrc.Close()

	// Get raw manifest
	manifestBytes, mimeType, err := imgSrc.GetManifest(ctx, nil)
	if err != nil {
		result := &InspectResult{
			Image:       imageName,
			InspectTime: time.Now(),
			Status:      "error",
			Error:       err.Error(),
		}
		i.setInspection(imageName, result)
		return result, fmt.Errorf("failed to get manifest for %s: %w", imageName, err)
	}

	// Compute digest
	dgst, err := manifest.Digest(manifestBytes)
	if err != nil {
		dgst = digest.FromBytes(manifestBytes)
	}

	// Create full image (parses config)
	img, err := ref.NewImage(ctx, sysCtx)
	if err != nil {
		result := &InspectResult{
			Image:       imageName,
			Digest:      dgst.String(),
			MediaType:   mimeType,
			RawManifest: manifestBytes,
			InspectTime: time.Now(),
			Status:      "error",
			Error:       fmt.Sprintf("manifest retrieved but failed to parse image config: %s", err.Error()),
		}
		i.setInspection(imageName, result)
		return result, fmt.Errorf("failed to create image for %s: %w", imageName, err)
	}
	defer img.Close()

	// Get image inspect info
	inspectInfo, err := img.Inspect(ctx)
	if err != nil {
		result := &InspectResult{
			Image:       imageName,
			Digest:      dgst.String(),
			MediaType:   mimeType,
			RawManifest: manifestBytes,
			InspectTime: time.Now(),
			Status:      "error",
			Error:       err.Error(),
		}
		i.setInspection(imageName, result)
		return result, fmt.Errorf("failed to inspect image %s: %w", imageName, err)
	}

	// Get config blob
	configBlob, err := img.ConfigBlob(ctx)
	if err != nil {
		i.log.Warn("Failed to get config blob", "image", imageName, "error", err)
	}

	// Build layers info
	var layers []InspectLayerInfo
	if inspectInfo.LayersData != nil {
		for _, ld := range inspectInfo.LayersData {
			layers = append(layers, InspectLayerInfo{
				MIMEType:    ld.MIMEType,
				Digest:      ld.Digest.String(),
				Size:        ld.Size,
				Annotations: ld.Annotations,
			})
		}
	} else {
		// Fallback to layer digests
		for _, l := range inspectInfo.Layers {
			layers = append(layers, InspectLayerInfo{
				Digest: l,
			})
		}
	}

	result := &InspectResult{
		Image:         imageName,
		Digest:        dgst.String(),
		MediaType:     mimeType,
		Architecture:  inspectInfo.Architecture,
		Variant:       inspectInfo.Variant,
		Os:            inspectInfo.Os,
		Created:       inspectInfo.Created,
		DockerVersion: inspectInfo.DockerVersion,
		Author:        inspectInfo.Author,
		Labels:        inspectInfo.Labels,
		Env:           inspectInfo.Env,
		Layers:        layers,
		RawManifest:   manifestBytes,
		RawConfig:     configBlob,
		InspectTime:   time.Now(),
		Status:        "completed",
	}

	i.setInspection(imageName, result)

	i.log.Info("Image inspection completed",
		"image", imageName,
		"digest", dgst.String(),
		"arch", inspectInfo.Architecture,
		"os", inspectInfo.Os,
		"layers", len(layers),
		"elapsed", time.Since(start),
	)

	return result, nil
}

// GetRawManifest retrieves only the raw manifest bytes for an image
func (i *ImageInspector) GetRawManifest(ctx context.Context, imageName string, auth *ImageAuth) ([]byte, string, error) {
	if !i.IsInitialized() {
		return nil, "", fmt.Errorf("image inspector not initialized")
	}
	if imageName == "" {
		return nil, "", fmt.Errorf("image name is required")
	}

	ctx, cancel := context.WithTimeout(ctx, i.config.Timeout)
	defer cancel()

	ref, err := parseImageReference(imageName)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse image reference %s: %w", imageName, err)
	}

	sysCtx := i.buildSystemContext(auth)

	imgSrc, err := ref.NewImageSource(ctx, sysCtx)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create image source for %s: %w", imageName, err)
	}
	defer imgSrc.Close()

	manifestBytes, mimeType, err := imgSrc.GetManifest(ctx, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get manifest for %s: %w", imageName, err)
	}

	return manifestBytes, mimeType, nil
}

// Stop cleans up the inspector
func (i *ImageInspector) Stop() {
	i.mx.Lock()
	defer i.mx.Unlock()
	i.initialized = false
	i.log.Info("Image inspector stopped")
}

// buildSystemContext creates a types.SystemContext with optional auth and TLS config
func (i *ImageInspector) buildSystemContext(auth *ImageAuth) *types.SystemContext {
	sysCtx := &types.SystemContext{}

	if auth != nil && auth.Username != "" {
		sysCtx.DockerAuthConfig = &types.DockerAuthConfig{
			Username: auth.Username,
			Password: auth.Password,
		}
	}

	if i.config.InsecureSkipTLSVerify {
		sysCtx.DockerInsecureSkipTLSVerify = types.OptionalBoolTrue
	}

	// Force Linux/AMD64 platform for consistent analysis
	sysCtx.OSChoice = "linux"
	sysCtx.ArchitectureChoice = "amd64"

	return sysCtx
}

// parseImageReference converts an image name like "docker.io/library/alpine:3.19"
// into a docker transport reference. Handles various formats:
//   - alpine:3.19           -> docker://docker.io/library/alpine:3.19
//   - myregistry.io/img:tag -> docker://myregistry.io/img:tag
//   - docker.io/lib/img     -> docker://docker.io/lib/img
func parseImageReference(imageName string) (types.ImageReference, error) {
	// docker.ParseReference expects "//" prefix (no "docker:" scheme)
	refStr := "//" + imageName

	ref, err := docker.ParseReference(refStr)
	if err != nil {
		return nil, fmt.Errorf("invalid image reference %q: %w", imageName, err)
	}
	return ref, nil
}

// NormalizeImageName ensures the image name has a registry prefix
func NormalizeImageName(imageName string) string {
	// If no "/" in the name, it's a Docker Hub official image
	if !strings.Contains(imageName, "/") {
		return "docker.io/library/" + imageName
	}
	// If only one "/" and no "." in the first part, it's a Docker Hub user image
	parts := strings.SplitN(imageName, "/", 2)
	if !strings.Contains(parts[0], ".") && !strings.Contains(parts[0], ":") {
		return "docker.io/" + imageName
	}
	return imageName
}
