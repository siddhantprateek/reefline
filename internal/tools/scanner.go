package tools

import (
	"context"
	"fmt"
	"time"

	"github.com/siddhantprateek/reefline/pkg/tools"
)

// ScannerStatus represents the current status of the vulnerability scanner
type ScannerStatus struct {
	Available   bool `json:"available"`
	Initialized bool `json:"initialized"`
}

// ScanRequest represents a request to scan images
type ScanRequest struct {
	Images []string `json:"images" binding:"required"`
}

// ScanResult represents the result of a vulnerability scan
type ScanResult struct {
	Image           string          `json:"image"`
	Vulnerabilities []Vulnerability `json:"vulnerabilities,omitempty"`
	Summary         Summary         `json:"summary"`
	ScanTime        string          `json:"scanTime"`
	Status          string          `json:"status"` // "completed", "queued", "not_found", "error"
}

// Vulnerability represents a single vulnerability finding
type Vulnerability struct {
	ID                     string                  `json:"id"`
	Severity               string                  `json:"severity"`
	PackageName            string                  `json:"packageName"`
	Version                string                  `json:"version"`
	FixVersion             string                  `json:"fixVersion"`
	PackageType            string                  `json:"packageType"`
	DataSource             string                  `json:"dataSource,omitempty"`
	Description            string                  `json:"description,omitempty"`
	PublishedDate          string                  `json:"publishedDate,omitempty"`
	LastModifiedDate       string                  `json:"lastModifiedDate,omitempty"`
	CVSSScore              *float64                `json:"cvssScore,omitempty"`
	CVSSVector             string                  `json:"cvssVector,omitempty"`
	CWEIDs                 []string                `json:"cweIds,omitempty"`
	Namespace              string                  `json:"namespace,omitempty"`
	PURL                   string                  `json:"purl,omitempty"`
	URLs                   []string                `json:"urls,omitempty"`
	Locations              []VulnerabilityLocation `json:"locations,omitempty"`
	RelatedVulnerabilities []RelatedVulnerability  `json:"relatedVulnerabilities,omitempty"`
}

// VulnerabilityLocation represents where a vulnerability was found
type VulnerabilityLocation struct {
	Path    string `json:"path"`
	LayerID string `json:"layerID,omitempty"`
}

// RelatedVulnerability represents related CVEs
type RelatedVulnerability struct {
	ID        string `json:"id"`
	Namespace string `json:"namespace,omitempty"`
}

// Summary represents the count of vulnerabilities by severity
type Summary struct {
	Critical int `json:"critical"`
	High     int `json:"high"`
	Medium   int `json:"medium"`
	Low      int `json:"low"`
	Unknown  int `json:"unknown"`
	Total    int `json:"total"`
}

// GetScannerStatus returns the current status of the vulnerability scanner
func GetScannerStatus() ScannerStatus {
	if tools.ImgScanner == nil {
		return ScannerStatus{
			Available:   false,
			Initialized: false,
		}
	}

	return ScannerStatus{
		Available:   true,
		Initialized: tools.ImgScanner.IsEnabled(),
	}
}

// ScanImages initiates vulnerability scans for the provided images
func ScanImages(ctx context.Context, images []string) ([]ScanResult, error) {
	if tools.ImgScanner == nil {
		return nil, fmt.Errorf("vulnerability scanner not initialized")
	}

	// Enqueue images for scanning (non-blocking)
	tools.ImgScanner.Enqueue(ctx, images...)

	var results []ScanResult

	// Check current scan status
	for _, img := range images {
		scan, found := tools.ImgScanner.GetScan(img)
		if found && scan != nil {
			result := ScanResult{
				Image:           img,
				Vulnerabilities: convertVulnerabilities(scan),
				Summary: Summary{
					Critical: scan.Tally.Critical,
					High:     scan.Tally.High,
					Medium:   scan.Tally.Medium,
					Low:      scan.Tally.Low,
					Unknown:  scan.Tally.Unknown,
					Total:    scan.Tally.Total,
				},
				ScanTime: time.Now().Format(time.RFC3339),
				Status:   "completed",
			}
			results = append(results, result)
		} else {
			// Scan is queued/in progress
			result := ScanResult{
				Image:    img,
				ScanTime: time.Now().Format(time.RFC3339),
				Status:   "queued",
			}
			results = append(results, result)
		}
	}

	return results, nil
}

// GetImageScanResults retrieves scan results for a specific image
func GetImageScanResults(image string) (*ScanResult, error) {
	if tools.ImgScanner == nil {
		return nil, fmt.Errorf("vulnerability scanner not initialized")
	}

	scan, found := tools.ImgScanner.GetScan(image)
	if !found {
		return &ScanResult{
			Image:    image,
			ScanTime: time.Now().Format(time.RFC3339),
			Status:   "not_found",
		}, nil
	}

	result := &ScanResult{
		Image:           image,
		Vulnerabilities: convertVulnerabilities(scan),
		Summary: Summary{
			Critical: scan.Tally.Critical,
			High:     scan.Tally.High,
			Medium:   scan.Tally.Medium,
			Low:      scan.Tally.Low,
			Unknown:  scan.Tally.Unknown,
			Total:    scan.Tally.Total,
		},
		ScanTime: time.Now().Format(time.RFC3339),
		Status:   "completed",
	}

	return result, nil
}

// convertVulnerabilities converts internal scan results to API format
func convertVulnerabilities(scan *tools.Scan) []Vulnerability {
	var vulns []Vulnerability

	for i, row := range scan.Table.Rows {
		if len(row) >= 6 {
			vuln := Vulnerability{
				ID:          row.Vulnerability(),
				Severity:    row.Severity(),
				PackageName: row.Name(),
				Version:     row.Version(),
				FixVersion:  row.Fix(),
				PackageType: row.Type(),
			}

			// Add enhanced metadata if available
			if i < len(scan.Table.Metadata) {
				meta := scan.Table.Metadata[i]
				if meta.VulnMetadata != nil {
					vuln.DataSource = meta.VulnMetadata.DataSource
					vuln.Description = meta.VulnMetadata.Description
					vuln.Namespace = meta.VulnMetadata.Namespace
					vuln.URLs = meta.VulnMetadata.URLs

					// Add CVSS information
					if len(meta.VulnMetadata.Cvss) > 0 {
						// Get the first (typically highest priority) CVSS score
						for _, cvss := range meta.VulnMetadata.Cvss {
							if cvss.Metrics.BaseScore > 0 {
								score := cvss.Metrics.BaseScore
								vuln.CVSSScore = &score
								vuln.CVSSVector = cvss.Vector
								break
							}
						}
					}

					// Add CISA KEV date if available
					if len(meta.VulnMetadata.KnownExploited) > 0 && meta.VulnMetadata.KnownExploited[0].DateAdded != nil {
						vuln.PublishedDate = meta.VulnMetadata.KnownExploited[0].DateAdded.Format("2006-01-02T15:04:05Z")
					}

					// Add EPSS date if available
					if len(meta.VulnMetadata.EPSS) > 0 {
						vuln.LastModifiedDate = meta.VulnMetadata.EPSS[0].Date.Format("2006-01-02T15:04:05Z")
					}
				}

				if meta.Match != nil {
					// Add package URL if available
					if meta.Match.Package.PURL != "" {
						vuln.PURL = meta.Match.Package.PURL
					}

					// Add locations
					for _, location := range meta.Match.Package.Locations.ToSlice() {
						vuln.Locations = append(vuln.Locations, VulnerabilityLocation{
							Path:    location.RealPath,
							LayerID: "",
						})
					}

					// Add related vulnerabilities
					for _, related := range meta.Match.Vulnerability.RelatedVulnerabilities {
						vuln.RelatedVulnerabilities = append(vuln.RelatedVulnerabilities, RelatedVulnerability{
							ID:        related.ID,
							Namespace: related.Namespace,
						})
					}
				}
			}

			vulns = append(vulns, vuln)
		}
	}

	return vulns
}
