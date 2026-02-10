package dockerhub

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	// DockerHubAPIBaseURL is the base URL for Docker Hub API v2
	DockerHubAPIBaseURL = "https://hub.docker.com/v2"
	// DockerRegistryBaseURL is the Docker Registry base for image references
	DockerRegistryBaseURL = "docker.io"
)

// Config holds the configuration for a Docker Hub integration
type Config struct {
	// PersonalAccessToken for authenticating with Docker Hub API
	PersonalAccessToken string `json:"patToken"`
	// Username is the Docker Hub username
	Username string `json:"username"`
}

// Client provides methods to interact with Docker Hub API
type Client struct {
	config     Config
	httpClient *http.Client
}

// NewClient creates a new Docker Hub integration client
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

// tokenTransport adds the PAT token to every Docker Hub API request
type tokenTransport struct {
	token     string
	transport http.RoundTripper
}

func (t *tokenTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.token))
	req.Header.Set("Content-Type", "application/json")
	return t.transport.RoundTrip(req)
}

// DockerRepository represents a repository on Docker Hub
type DockerRepository struct {
	Name           string `json:"name"`
	Namespace      string `json:"namespace"`
	Description    string `json:"description"`
	IsPrivate      bool   `json:"is_private"`
	StarCount      int    `json:"star_count"`
	PullCount      int64  `json:"pull_count"`
	LastUpdated    string `json:"last_updated"`
	RepositoryType string `json:"repository_type"`
}

// ImageTag represents a tag/version of a Docker Hub image
type ImageTag struct {
	Name        string      `json:"name"`
	FullSize    int64       `json:"full_size"`
	LastUpdated string      `json:"last_updated"`
	Digest      string      `json:"digest"`
	Images      []ImageArch `json:"images"`
}

// ImageArch represents architecture-specific image info within a tag
type ImageArch struct {
	Architecture string `json:"architecture"`
	OS           string `json:"os"`
	Size         int64  `json:"size"`
	Digest       string `json:"digest"`
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

// ValidateCredentials checks if the PAT and username are valid.
// Returns the authenticated user info or an error.
func (c *Client) ValidateCredentials(ctx context.Context) (string, error) {
	// Docker Hub v2 API: GET /v2/users/{username}
	url := fmt.Sprintf("%s/users/%s", DockerHubAPIBaseURL, c.config.Username)
	data, status, err := c.doRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to validate: %w", err)
	}

	if status == http.StatusUnauthorized || status == http.StatusForbidden {
		return "", fmt.Errorf("invalid credentials: %d", status)
	}
	if status == http.StatusNotFound {
		return "", fmt.Errorf("user not found: %s", c.config.Username)
	}
	if status != http.StatusOK {
		return "", fmt.Errorf("unexpected status %d: %s", status, string(data))
	}

	var user struct {
		ID       string `json:"id"`
		Username string `json:"username"`
	}
	if err := json.Unmarshal(data, &user); err != nil {
		return "", fmt.Errorf("failed to parse user response: %w", err)
	}

	return user.Username, nil
}

// ListRepositories returns repositories belonging to the authenticated user's namespace.
func (c *Client) ListRepositories(ctx context.Context, page, pageSize int) ([]DockerRepository, error) {
	url := fmt.Sprintf("%s/repositories/%s/?page=%d&page_size=%d", DockerHubAPIBaseURL, c.config.Username, page, pageSize)
	data, status, err := c.doRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", status, string(data))
	}

	var response struct {
		Results []DockerRepository `json:"results"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse repos: %w", err)
	}
	return response.Results, nil
}

// GetRepository returns details of a specific repository.
func (c *Client) GetRepository(ctx context.Context, namespace, repo string) (*DockerRepository, error) {
	url := fmt.Sprintf("%s/repositories/%s/%s/", DockerHubAPIBaseURL, namespace, repo)
	data, status, err := c.doRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", status, string(data))
	}

	var repository DockerRepository
	if err := json.Unmarshal(data, &repository); err != nil {
		return nil, fmt.Errorf("failed to parse repo: %w", err)
	}
	return &repository, nil
}

// ListTags returns available tags for a repository.
func (c *Client) ListTags(ctx context.Context, namespace, repo string, page, pageSize int) ([]ImageTag, error) {
	url := fmt.Sprintf("%s/repositories/%s/%s/tags/?page=%d&page_size=%d", DockerHubAPIBaseURL, namespace, repo, page, pageSize)
	data, status, err := c.doRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", status, string(data))
	}

	var response struct {
		Results []ImageTag `json:"results"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse tags: %w", err)
	}
	return response.Results, nil
}

// GetTag returns details of a specific tag.
func (c *Client) GetTag(ctx context.Context, namespace, repo, tag string) (*ImageTag, error) {
	url := fmt.Sprintf("%s/repositories/%s/%s/tags/%s", DockerHubAPIBaseURL, namespace, repo, tag)
	data, status, err := c.doRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", status, string(data))
	}

	var imageTag ImageTag
	if err := json.Unmarshal(data, &imageTag); err != nil {
		return nil, fmt.Errorf("failed to parse tag: %w", err)
	}
	return &imageTag, nil
}

// ImageRef returns the full Docker Hub image reference.
// Format: docker.io/{namespace}/{repo}:{tag}
func ImageRef(namespace, repo, tag string) string {
	if tag == "" {
		tag = "latest"
	}
	return fmt.Sprintf("%s/%s/%s:%s", DockerRegistryBaseURL, namespace, repo, tag)
}

// SearchImages searches Docker Hub for public images matching the query.
func (c *Client) SearchImages(ctx context.Context, query string, page, pageSize int) ([]DockerRepository, error) {
	url := fmt.Sprintf("%s/search/repositories/?query=%s&page=%d&page_size=%d", DockerHubAPIBaseURL, query, page, pageSize)
	data, status, err := c.doRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", status, string(data))
	}

	var response struct {
		Results []DockerRepository `json:"results"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse search results: %w", err)
	}
	return response.Results, nil
}

// GetImageDigest returns the digest for a specific image:tag.
func (c *Client) GetImageDigest(ctx context.Context, namespace, repo, tag string) (string, error) {
	imageTag, err := c.GetTag(ctx, namespace, repo, tag)
	if err != nil {
		return "", err
	}
	return imageTag.Digest, nil
}
