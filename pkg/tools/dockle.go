package tools

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"sync"
	"time"

	deckodertypes "github.com/goodwithtech/deckoder/types"
	"github.com/goodwithtech/dockle/config"
	"github.com/goodwithtech/dockle/pkg/assessor/credential"
	"github.com/goodwithtech/dockle/pkg/assessor/manifest"
	"github.com/goodwithtech/dockle/pkg/scanner"
	"github.com/goodwithtech/dockle/pkg/types"
)

const (
	dockleScanTimeout = 5 * time.Minute
)

// Global dockle scanner instance
var DockleScn *DockleScanner

// DockleConfig holds configuration for the dockle scanner
type DockleConfig struct {
	Enable         bool          `json:"enable"`
	Timeout        time.Duration `json:"timeout"`
	IgnoreCodes    []string      `json:"ignoreCodes"`
	AcceptFiles    []string      `json:"acceptFiles"`
	AcceptExts     []string      `json:"acceptExts"`
	SensitiveWords []string      `json:"sensitiveWords"`
	SensitiveFiles []string      `json:"sensitiveFiles"`
}

// DockleScanner wraps dockle's scanning functionality
type DockleScanner struct {
	mx          sync.RWMutex
	initialized bool
	config      DockleConfig
	scans       map[string]*DockleScan
	ignoreMap   map[string]struct{}
	log         *slog.Logger
}

// DockleScan holds results from a single image scan
type DockleScan struct {
	Image       string             `json:"image"`
	Assessments []DockleAssessment `json:"assessments"`
	Summary     DockleSummary      `json:"summary"`
	ScanTime    time.Time          `json:"scanTime"`
	Status      string             `json:"status"` // "completed", "queued", "error"
	Error       string             `json:"error,omitempty"`
}

// DockleAssessment represents a single checkpoint result
type DockleAssessment struct {
	Code     string   `json:"code"`
	Title    string   `json:"title"`
	Level    string   `json:"level"`
	LevelInt int      `json:"levelInt"`
	Alerts   []string `json:"alerts,omitempty"`
}

// DockleSummary holds counts by severity
type DockleSummary struct {
	Fatal int `json:"fatal"`
	Warn  int `json:"warn"`
	Info  int `json:"info"`
	Skip  int `json:"skip"`
	Pass  int `json:"pass"`
	Total int `json:"total"`
}

// NewDockleScanner creates a new dockle scanner
func NewDockleScanner(cfg DockleConfig, l *slog.Logger) *DockleScanner {
	if cfg.Timeout == 0 {
		cfg.Timeout = dockleScanTimeout
	}
	return &DockleScanner{
		config: cfg,
		scans:  make(map[string]*DockleScan),
		log:    l.With("subsys", "dockle"),
	}
}

// Init configures the dockle assessor settings and global config
func (s *DockleScanner) Init() {
	s.mx.Lock()
	defer s.mx.Unlock()

	// Build ignore map
	s.ignoreMap = make(map[string]struct{})
	for _, code := range s.config.IgnoreCodes {
		s.ignoreMap[code] = struct{}{}
	}

	// Set the global dockle config (no CLI dependency)
	config.Conf = config.Config{
		IgnoreMap: s.ignoreMap,
		ExitCode:  0,
		ExitLevel: types.WarnLevel,
	}

	// Configure assessor-specific settings
	if len(s.config.SensitiveWords) > 0 {
		manifest.AddSensitiveWords(s.config.SensitiveWords)
	}
	if len(s.config.SensitiveFiles) > 0 {
		credential.AddSensitiveFiles(s.config.SensitiveFiles)
	}
	if len(s.config.AcceptFiles) > 0 {
		scanner.AddAcceptanceFiles(s.config.AcceptFiles)
	}
	if len(s.config.AcceptExts) > 0 {
		scanner.AddAcceptanceExtensions(s.config.AcceptExts)
	}

	s.initialized = true
	s.log.Info("Dockle scanner initialized")
}

// IsEnabled returns whether the scanner is enabled
func (s *DockleScanner) IsEnabled() bool {
	return s.config.Enable
}

// IsInitialized returns whether the scanner has been initialized
func (s *DockleScanner) IsInitialized() bool {
	s.mx.RLock()
	defer s.mx.RUnlock()
	return s.initialized
}

// GetScan retrieves a cached scan result
func (s *DockleScanner) GetScan(img string) (*DockleScan, bool) {
	s.mx.RLock()
	defer s.mx.RUnlock()
	sc, ok := s.scans[img]
	return sc, ok
}

func (s *DockleScanner) setScan(img string, sc *DockleScan) {
	s.mx.Lock()
	defer s.mx.Unlock()
	s.scans[img] = sc
}

// ScanImage scans a container image by name and returns results
func (s *DockleScanner) ScanImage(ctx context.Context, imageName string) (*DockleScan, error) {
	if !s.IsInitialized() {
		return nil, fmt.Errorf("dockle scanner not initialized")
	}
	if imageName == "" {
		return nil, fmt.Errorf("image name is required")
	}
	return s.doScan(ctx, imageName, "")
}

// ScanImageFromFile scans a container image from a local tar archive
func (s *DockleScanner) ScanImageFromFile(ctx context.Context, filePath string) (*DockleScan, error) {
	if !s.IsInitialized() {
		return nil, fmt.Errorf("dockle scanner not initialized")
	}
	if filePath == "" {
		return nil, fmt.Errorf("file path is required")
	}
	return s.doScan(ctx, "", filePath)
}

func (s *DockleScanner) doScan(ctx context.Context, imageName, filePath string) (*DockleScan, error) {
	scanID := imageName
	if scanID == "" {
		scanID = filePath
	}

	start := time.Now()
	s.log.Info("Starting dockle scan", "image", scanID)

	ctx, cancel := context.WithTimeout(ctx, s.config.Timeout)
	defer cancel()

	dockerOption := deckodertypes.DockerOption{
		Timeout:  s.config.Timeout,
		SkipPing: true,
	}

	// Call dockle's scanner
	assessments, err := scanner.ScanImage(ctx, imageName, filePath, dockerOption)
	if err != nil {
		scan := &DockleScan{
			Image:    scanID,
			ScanTime: time.Now(),
			Status:   "error",
			Error:    err.Error(),
		}
		s.setScan(scanID, scan)
		return scan, fmt.Errorf("dockle scan failed for %s: %w", scanID, err)
	}

	// Check for latest tag usage (same logic as dockle's pkg/run.go)
	if imageName != "" && isLatestTag(imageName) {
		assessments = append(assessments, &types.Assessment{
			Code:     types.AvoidLatestTag,
			Filename: imageName,
			Desc:     "Avoid 'latest' tag",
		})
	}

	// Create assessment map with ignore rules
	assessmentMap := types.CreateAssessmentMap(assessments, s.ignoreMap, false)

	// Convert to our result format
	scan := convertAssessmentMap(scanID, assessmentMap)
	s.setScan(scanID, scan)

	s.log.Info("Dockle scan completed",
		"image", scanID,
		"fatal", scan.Summary.Fatal,
		"warn", scan.Summary.Warn,
		"info", scan.Summary.Info,
		"pass", scan.Summary.Pass,
		"elapsed", time.Since(start),
	)

	return scan, nil
}

// Stop cleans up the scanner
func (s *DockleScanner) Stop() {
	s.mx.Lock()
	defer s.mx.Unlock()
	s.initialized = false
	s.log.Info("Dockle scanner stopped")
}

// convertAssessmentMap transforms dockle's AssessmentMap into our DockleScan format
func convertAssessmentMap(image string, am types.AssessmentMap) *DockleScan {
	scan := &DockleScan{
		Image:    image,
		ScanTime: time.Now(),
		Status:   "completed",
	}

	// Build ordered code list sorted by severity (highest first)
	codeOrder := make(types.ByLevel, 0, len(types.DefaultLevelMap))
	for code, level := range types.DefaultLevelMap {
		codeOrder = append(codeOrder, types.Assessment{Code: code, Level: level})
	}
	sort.Sort(codeOrder)

	for _, ordered := range codeOrder {
		codeInfo, found := am[ordered.Code]
		if !found {
			// Checkpoint passed
			scan.Assessments = append(scan.Assessments, DockleAssessment{
				Code:     ordered.Code,
				Title:    types.TitleMap[ordered.Code],
				Level:    "PASS",
				LevelInt: types.PassLevel,
			})
			scan.Summary.Pass++
			scan.Summary.Total++
			continue
		}

		// Collect alert descriptions
		var alerts []string
		for _, a := range codeInfo.Assessments {
			if a.Desc != "" {
				alerts = append(alerts, a.Desc)
			}
		}

		scan.Assessments = append(scan.Assessments, DockleAssessment{
			Code:     codeInfo.Code,
			Title:    types.TitleMap[codeInfo.Code],
			Level:    levelToString(codeInfo.Level),
			LevelInt: codeInfo.Level,
			Alerts:   alerts,
		})

		switch codeInfo.Level {
		case types.FatalLevel:
			scan.Summary.Fatal++
		case types.WarnLevel:
			scan.Summary.Warn++
		case types.InfoLevel:
			scan.Summary.Info++
		case types.SkipLevel:
			scan.Summary.Skip++
		default:
			scan.Summary.Pass++
		}
		scan.Summary.Total++
	}

	return scan
}

func levelToString(level int) string {
	switch level {
	case types.FatalLevel:
		return "FATAL"
	case types.WarnLevel:
		return "WARN"
	case types.InfoLevel:
		return "INFO"
	case types.PassLevel:
		return "PASS"
	case types.SkipLevel:
		return "SKIP"
	case types.IgnoreLevel:
		return "IGNORE"
	default:
		return "UNKNOWN"
	}
}

func isLatestTag(imageName string) bool {
	// If no tag specified, Docker defaults to :latest
	if !strings.Contains(imageName, ":") {
		return true
	}
	return strings.HasSuffix(imageName, ":latest")
}
