package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	// GitHubAPIBaseURL is the base URL for GitHub REST API v3
	GitHubAPIBaseURL = "https://api.github.com"
	// GHCRBaseURL is the base URL for GitHub Container Registry
	GHCRBaseURL = "ghcr.io"
)

// Config holds the configuration for a GitHub integration
type Config struct {
	// PersonalAccessToken (PAT) for authenticating with GitHub API and GHCR
	// Required scopes: repo, read:packages, write:packages, read:org (optional)
	PersonalAccessToken string `json:"patToken"`
}

// Client provides methods to interact with GitHub API and GHCR
type Client struct {
	config     Config
	httpClient *http.Client
}

// NewClient creates a new GitHub integration client
func NewClient(config Config) *Client {
	return &Client{
		config: config,
		httpClient: &http.Client{
			Transport: &tokenTransport{
				token:     config.PersonalAccessToken,
				transport: http.DefaultTransport,
			},
		},
	}
}

// tokenTransport adds the PAT token to every request as a Bearer token
type tokenTransport struct {
	token     string
	transport http.RoundTripper
}

func (t *tokenTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.token))
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	return t.transport.RoundTrip(req)
}

// Repository represents a GitHub repository
type Repository struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	FullName      string `json:"full_name"`
	Description   string `json:"description"`
	Private       bool   `json:"private"`
	HTMLURL       string `json:"html_url"`
	CloneURL      string `json:"clone_url"`
	DefaultBranch string `json:"default_branch"`
}

// ContainerImage represents an image in GitHub Container Registry (GHCR)
type ContainerImage struct {
	ID          int64    `json:"id"`
	Name        string   `json:"name"`
	PackageType string   `json:"package_type"`
	HTMLURL     string   `json:"html_url"`
	Tags        []string `json:"tags"`
}

// FileContent represents the content of a file fetched from a repo
type FileContent struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Content string `json:"content"` // base64-encoded content
	SHA     string `json:"sha"`
}

// Issue represents a GitHub issue
type Issue struct {
	ID      int64  `json:"id"`
	Number  int    `json:"number"`
	Title   string `json:"title"`
	Body    string `json:"body"`
	State   string `json:"state"`
	HTMLURL string `json:"html_url"`
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

// ValidateCredentials checks if the PAT is valid by calling the /user endpoint.
// Returns the authenticated username or an error.
func (c *Client) ValidateCredentials(ctx context.Context) (string, error) {
	data, status, err := c.doRequest(ctx, http.MethodGet, GitHubAPIBaseURL+"/user", nil)
	if err != nil {
		return "", fmt.Errorf("failed to validate: %w", err)
	}

	if status == http.StatusUnauthorized {
		return "", fmt.Errorf("invalid token: 401 Unauthorized")
	}
	if status != http.StatusOK {
		return "", fmt.Errorf("unexpected status %d: %s", status, string(data))
	}

	var user struct {
		Login string `json:"login"`
	}
	if err := json.Unmarshal(data, &user); err != nil {
		return "", fmt.Errorf("failed to parse user response: %w", err)
	}

	return user.Login, nil
}

// ListRepositories returns repositories accessible to the authenticated user.
// Supports pagination via page and perPage parameters.
func (c *Client) ListRepositories(ctx context.Context, page, perPage int) ([]Repository, error) {
	url := fmt.Sprintf("%s/user/repos?page=%d&per_page=%d&sort=updated", GitHubAPIBaseURL, page, perPage)
	data, status, err := c.doRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", status, string(data))
	}

	var repos []Repository
	if err := json.Unmarshal(data, &repos); err != nil {
		return nil, fmt.Errorf("failed to parse repos: %w", err)
	}
	return repos, nil
}

// GetRepository returns details of a specific repository.
func (c *Client) GetRepository(ctx context.Context, owner, repo string) (*Repository, error) {
	url := fmt.Sprintf("%s/repos/%s/%s", GitHubAPIBaseURL, owner, repo)
	data, status, err := c.doRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", status, string(data))
	}

	var repository Repository
	if err := json.Unmarshal(data, &repository); err != nil {
		return nil, fmt.Errorf("failed to parse repo: %w", err)
	}
	return &repository, nil
}

// GetFileContent retrieves a file's content from a repository.
func (c *Client) GetFileContent(ctx context.Context, owner, repo, path, ref string) (*FileContent, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/contents/%s", GitHubAPIBaseURL, owner, repo, path)
	if ref != "" {
		url += "?ref=" + ref
	}

	data, status, err := c.doRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if status == http.StatusNotFound {
		return nil, fmt.Errorf("file not found: %s", path)
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", status, string(data))
	}

	var fc FileContent
	if err := json.Unmarshal(data, &fc); err != nil {
		return nil, fmt.Errorf("failed to parse file content: %w", err)
	}
	return &fc, nil
}

// GetDockerfile is a convenience method that fetches a Dockerfile from a repository.
// It tries common Dockerfile paths: Dockerfile, docker/Dockerfile, .docker/Dockerfile
func (c *Client) GetDockerfile(ctx context.Context, owner, repo, path, ref string) (string, error) {
	paths := []string{path}
	if path == "" {
		paths = []string{"Dockerfile", "docker/Dockerfile", ".docker/Dockerfile"}
	}

	for _, p := range paths {
		fc, err := c.GetFileContent(ctx, owner, repo, p, ref)
		if err == nil {
			return fc.Content, nil
		}
	}
	return "", fmt.Errorf("no Dockerfile found in repository %s/%s", owner, repo)
}

// ListContainerImages lists container images published to GHCR for a user/org.
func (c *Client) ListContainerImages(ctx context.Context, owner string, page, perPage int) ([]ContainerImage, error) {
	url := fmt.Sprintf("%s/users/%s/packages?package_type=container&page=%d&per_page=%d", GitHubAPIBaseURL, owner, page, perPage)
	data, status, err := c.doRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", status, string(data))
	}

	var images []ContainerImage
	if err := json.Unmarshal(data, &images); err != nil {
		return nil, fmt.Errorf("failed to parse images: %w", err)
	}
	return images, nil
}

// GetContainerImageTags returns available tags for a GHCR image.
func (c *Client) GetContainerImageTags(ctx context.Context, owner, imageName string) ([]string, error) {
	url := fmt.Sprintf("%s/users/%s/packages/container/%s/versions", GitHubAPIBaseURL, owner, imageName)
	data, status, err := c.doRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", status, string(data))
	}

	var versions []struct {
		Metadata struct {
			Container struct {
				Tags []string `json:"tags"`
			} `json:"container"`
		} `json:"metadata"`
	}
	if err := json.Unmarshal(data, &versions); err != nil {
		return nil, fmt.Errorf("failed to parse versions: %w", err)
	}

	var tags []string
	for _, v := range versions {
		tags = append(tags, v.Metadata.Container.Tags...)
	}
	return tags, nil
}

// GHCRImageRef returns the full GHCR image reference for use with analysis tools.
// Format: ghcr.io/{owner}/{image}:{tag}
func GHCRImageRef(owner, image, tag string) string {
	if tag == "" {
		tag = "latest"
	}
	return fmt.Sprintf("%s/%s/%s:%s", GHCRBaseURL, owner, image, tag)
}

// CreateIssue creates a new issue in a GitHub repository.
func (c *Client) CreateIssue(ctx context.Context, owner, repo, title, body string, labels []string) (*Issue, error) {
	payload := map[string]interface{}{
		"title":  title,
		"body":   body,
		"labels": labels,
	}
	payloadJSON, _ := json.Marshal(payload)

	url := fmt.Sprintf("%s/repos/%s/%s/issues", GitHubAPIBaseURL, owner, repo)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, io.NopCloser(
		io.Reader(
			jsonReader(payloadJSON),
		),
	))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create issue: %w", err)
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("failed to create issue (status %d): %s", resp.StatusCode, string(data))
	}

	var issue Issue
	if err := json.Unmarshal(data, &issue); err != nil {
		return nil, fmt.Errorf("failed to parse issue: %w", err)
	}
	return &issue, nil
}

// CreateOptimizationIssue creates a GitHub issue with the analysis results
// formatted as a markdown report with recommendations.
func (c *Client) CreateOptimizationIssue(ctx context.Context, owner, repo string, reportSummary string, recommendations []string) (*Issue, error) {
	body := "## Container Image Optimization Report\n\n"
	body += reportSummary + "\n\n"
	body += "### Recommendations\n\n"
	for _, rec := range recommendations {
		body += "- [ ] " + rec + "\n"
	}
	body += "\n---\n*Generated by Reefline*\n"

	return c.CreateIssue(ctx, owner, repo, "Container Image Optimization Report", body, []string{"optimization"})
}

// jsonReader is a helper to create an io.Reader from a byte slice
func jsonReader(data []byte) io.Reader {
	return io.NopCloser(
		readerFromBytes(data),
	)
}

type bytesReader struct {
	data []byte
	pos  int
}

func readerFromBytes(data []byte) *bytesReader {
	return &bytesReader{data: data}
}

func (r *bytesReader) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n = copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}
