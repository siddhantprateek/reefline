package harbor

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Config holds the configuration for a Harbor integration.
// Harbor uses URL + basic auth (username/password).
type Config struct {
	URL      string `json:"url"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// Client provides methods to interact with the Harbor API v2
type Client struct {
	config     Config
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a new Harbor integration client
func NewClient(config Config) *Client {
	baseURL := config.URL
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		baseURL = "https://" + baseURL
	}
	if len(baseURL) > 0 && baseURL[len(baseURL)-1] == '/' {
		baseURL = baseURL[:len(baseURL)-1]
	}

	return &Client{
		config: config,
		httpClient: &http.Client{
			Transport: &basicAuthTransport{
				username:  config.Username,
				password:  config.Password,
				transport: http.DefaultTransport,
			},
		},
		baseURL: baseURL,
	}
}

// basicAuthTransport adds Basic auth credentials to every request
type basicAuthTransport struct {
	username  string
	password  string
	transport http.RoundTripper
}

func (t *basicAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.SetBasicAuth(t.username, t.password)
	req.Header.Set("Accept", "application/json")
	return t.transport.RoundTrip(req)
}

// Project represents a Harbor project
type Project struct {
	ProjectID int64  `json:"project_id"`
	Name      string `json:"name"`
	RepoCount int    `json:"repo_count"`
	OwnerName string `json:"owner_name"`
	CreatedAt string `json:"creation_time"`
	UpdatedAt string `json:"update_time"`
}

// HarborRepository represents a repository within a Harbor project
type HarborRepository struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	ArtifactCount int    `json:"artifact_count"`
	PullCount     int64  `json:"pull_count"`
	CreatedAt     string `json:"creation_time"`
	UpdatedAt     string `json:"update_time"`
}

// Artifact represents a container image artifact in Harbor
type Artifact struct {
	ID           int64                  `json:"id"`
	Digest       string                 `json:"digest"`
	Size         int64                  `json:"size"`
	Tags         []Tag                  `json:"tags"`
	PushTime     string                 `json:"push_time"`
	PullTime     string                 `json:"pull_time"`
	ProjectID    int64                  `json:"project_id"`
	RepositoryID int64                  `json:"repository_id"`
	MediaType    string                 `json:"media_type"`
	ExtraAttrs   map[string]interface{} `json:"extra_attrs"`
}

// Tag represents a tag attached to an artifact
type Tag struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	PushTime   string `json:"push_time"`
	ArtifactID int64  `json:"artifact_id"`
}

// VulnerabilitySummary represents Harbor's built-in vulnerability scan results
type VulnerabilitySummary struct {
	Total      int    `json:"total"`
	Critical   int    `json:"critical"`
	High       int    `json:"high"`
	Medium     int    `json:"medium"`
	Low        int    `json:"low"`
	Negligible int    `json:"negligible"`
	ScanStatus string `json:"scan_status"`
}

// doRequest is a helper that executes an HTTP request and returns the response body.
func (c *Client) doRequest(ctx context.Context, method, url string, body io.Reader) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("failed to read response: %w", err)
	}

	return data, resp.StatusCode, nil
}

// ValidateCredentials checks if the Harbor URL, username, and password are valid.
// Returns the Harbor version on success.
func (c *Client) ValidateCredentials(ctx context.Context) (string, error) {
	// Check credentials against a protected endpoint (401 if invalid)
	permURL := fmt.Sprintf("%s/api/v2.0/users/current/permissions", c.baseURL)
	_, status, err := c.doRequest(ctx, http.MethodGet, permURL, nil)
	if err != nil {
		return "", fmt.Errorf("cannot reach Harbor at %s: %w", c.baseURL, err)
	}

	if status == http.StatusUnauthorized {
		return "", fmt.Errorf("invalid credentials: 401 Unauthorized")
	}
	if status != http.StatusOK {
		return "", fmt.Errorf("unexpected status checking permissions: %d", status)
	}

	// Get version info via systeminfo (which might be public, but we already validated auth)
	url := fmt.Sprintf("%s/api/v2.0/systeminfo", c.baseURL)
	data, status, err := c.doRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to reach system info: %w", err)
	}
	// We expect 200 here too
	if status != http.StatusOK {
		return "", fmt.Errorf("unexpected status fetch system info: %d", status)
	}

	var info struct {
		HarborVersion string `json:"harbor_version"`
	}
	if err := json.Unmarshal(data, &info); err != nil {
		return "", fmt.Errorf("failed to parse system info: %w", err)
	}

	return info.HarborVersion, nil
}

// ListProjects returns all projects accessible to the authenticated user.
func (c *Client) ListProjects(ctx context.Context, page, pageSize int) ([]Project, error) {
	url := fmt.Sprintf("%s/api/v2.0/projects?page=%d&page_size=%d", c.baseURL, page, pageSize)
	data, status, err := c.doRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", status, string(data))
	}

	var projects []Project
	if err := json.Unmarshal(data, &projects); err != nil {
		return nil, fmt.Errorf("failed to parse projects: %w", err)
	}
	return projects, nil
}

// GetProject returns details of a specific project.
func (c *Client) GetProject(ctx context.Context, projectName string) (*Project, error) {
	url := fmt.Sprintf("%s/api/v2.0/projects?name=%s", c.baseURL, projectName)
	data, status, err := c.doRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", status, string(data))
	}

	var projects []Project
	if err := json.Unmarshal(data, &projects); err != nil {
		return nil, fmt.Errorf("failed to parse projects: %w", err)
	}
	if len(projects) == 0 {
		return nil, fmt.Errorf("project not found: %s", projectName)
	}
	return &projects[0], nil
}

// ListRepositories returns repositories within a Harbor project.
func (c *Client) ListRepositories(ctx context.Context, projectName string, page, pageSize int) ([]HarborRepository, error) {
	url := fmt.Sprintf("%s/api/v2.0/projects/%s/repositories?page=%d&page_size=%d", c.baseURL, projectName, page, pageSize)
	data, status, err := c.doRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", status, string(data))
	}

	var repos []HarborRepository
	if err := json.Unmarshal(data, &repos); err != nil {
		return nil, fmt.Errorf("failed to parse repos: %w", err)
	}
	return repos, nil
}

// ListArtifacts returns artifacts for a repository.
func (c *Client) ListArtifacts(ctx context.Context, projectName, repoName string, page, pageSize int) ([]Artifact, error) {
	url := fmt.Sprintf("%s/api/v2.0/projects/%s/repositories/%s/artifacts?page=%d&page_size=%d&with_tag=true",
		c.baseURL, projectName, repoName, page, pageSize)
	data, status, err := c.doRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", status, string(data))
	}

	var artifacts []Artifact
	if err := json.Unmarshal(data, &artifacts); err != nil {
		return nil, fmt.Errorf("failed to parse artifacts: %w", err)
	}
	return artifacts, nil
}

// GetArtifact returns details of a specific artifact by tag or digest.
func (c *Client) GetArtifact(ctx context.Context, projectName, repoName, reference string) (*Artifact, error) {
	url := fmt.Sprintf("%s/api/v2.0/projects/%s/repositories/%s/artifacts/%s",
		c.baseURL, projectName, repoName, reference)
	data, status, err := c.doRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", status, string(data))
	}

	var artifact Artifact
	if err := json.Unmarshal(data, &artifact); err != nil {
		return nil, fmt.Errorf("failed to parse artifact: %w", err)
	}
	return &artifact, nil
}

// GetVulnerabilitySummary returns the vulnerability scan results for an artifact.
func (c *Client) GetVulnerabilitySummary(ctx context.Context, projectName, repoName, reference string) (*VulnerabilitySummary, error) {
	url := fmt.Sprintf("%s/api/v2.0/projects/%s/repositories/%s/artifacts/%s?with_scan_overview=true",
		c.baseURL, projectName, repoName, reference)
	data, status, err := c.doRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", status, string(data))
	}

	// The scan overview is nested inside the artifact response
	var artifact struct {
		ScanOverview map[string]struct {
			Summary *VulnerabilitySummary `json:"summary"`
		} `json:"scan_overview"`
	}
	if err := json.Unmarshal(data, &artifact); err != nil {
		return nil, fmt.Errorf("failed to parse scan overview: %w", err)
	}

	// Return the first scan report found
	for _, overview := range artifact.ScanOverview {
		if overview.Summary != nil {
			return overview.Summary, nil
		}
	}

	return nil, fmt.Errorf("no scan results available for this artifact")
}

// TriggerScan triggers a vulnerability scan for an artifact.
func (c *Client) TriggerScan(ctx context.Context, projectName, repoName, reference string) error {
	url := fmt.Sprintf("%s/api/v2.0/projects/%s/repositories/%s/artifacts/%s/scan",
		c.baseURL, projectName, repoName, reference)
	_, status, err := c.doRequest(ctx, http.MethodPost, url, nil)
	if err != nil {
		return err
	}
	if status != http.StatusAccepted && status != http.StatusOK {
		return fmt.Errorf("failed to trigger scan: status %d", status)
	}
	return nil
}

// ImageRef returns the full Harbor image reference.
// Format: {harbor_host}/{project}/{repo}:{tag}
func (c *Client) ImageRef(projectName, repoName, tag string) string {
	if tag == "" {
		tag = "latest"
	}
	host := c.baseURL
	if len(host) > 8 && host[:8] == "https://" {
		host = host[8:]
	} else if len(host) > 7 && host[:7] == "http://" {
		host = host[7:]
	}
	return fmt.Sprintf("%s/%s/%s:%s", host, projectName, repoName, tag)
}
