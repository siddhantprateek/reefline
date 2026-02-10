package dockerhub

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
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
	jwt        string
	mu         sync.RWMutex
}

// NewClient creates a new Docker Hub integration client
func NewClient(config Config) *Client {
	return &Client{
		config:     config,
		httpClient: http.DefaultClient,
	}
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

// login authenticates with Docker Hub using the PAT to get a JWT.
func (c *Client) login(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// If already have a token, assume it's valid for now
	// TODO: Check expiration if needed
	if c.jwt != "" {
		return nil
	}

	url := fmt.Sprintf("%s/users/login", DockerHubAPIBaseURL)
	payload := map[string]string{
		"username": c.config.Username,
		"password": c.config.PersonalAccessToken,
	}
	bodyBytes, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create login request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("login failed (status %d): %s", resp.StatusCode, string(body))
	}

	var response struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(readBody(resp.Body), &response); err != nil {
		return fmt.Errorf("failed to parse login response: %w", err)
	}

	c.jwt = response.Token
	return nil
}

// readBody reads and restores the body for unmarshal
func readBody(r io.ReadCloser) []byte {
	data, _ := io.ReadAll(r)
	return data
}

// doRequest executes an HTTP request with automatic authentication.
func (c *Client) doRequest(ctx context.Context, method, url string, body io.Reader) ([]byte, int, error) {
	// Ensure we are logged in (unless we are already doing a login, handled inside login())
	if err := c.login(ctx); err != nil {
		return nil, 0, err
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %w", err)
	}

	c.mu.RLock()
	token := c.jwt
	c.mu.RUnlock()

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")

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

// ValidateCredentials checks if the PAT and username are valid by attempting to login.
func (c *Client) ValidateCredentials(ctx context.Context) (string, error) {
	// Trying to login is sufficient validation
	if err := c.login(ctx); err != nil {
		return "", err
	}
	return c.config.Username, nil
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

	// Docker Hub search structure is slightly different, but let's assume standard pagination for now
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
