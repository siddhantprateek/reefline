package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/siddhantprateek/reefline/internal/integration/ai"
	"github.com/siddhantprateek/reefline/internal/integration/dockerhub"
	"github.com/siddhantprateek/reefline/internal/integration/github"
	"github.com/siddhantprateek/reefline/internal/integration/harbor"
	k8s "github.com/siddhantprateek/reefline/internal/integration/kubernetes"
	"github.com/siddhantprateek/reefline/pkg/crypto"
	"github.com/siddhantprateek/reefline/pkg/database"
	"github.com/siddhantprateek/reefline/pkg/models"
)

// IntegrationHandler handles CRUD operations for user integrations
type IntegrationHandler struct{}

// NewIntegrationHandler creates a new IntegrationHandler instance
func NewIntegrationHandler() *IntegrationHandler {
	return &IntegrationHandler{}
}

// knownIntegrations lists all supported integration IDs
var knownIntegrations = []string{
	"docker", "harbor", "github", "kubernetes",
	"openai", "anthropic", "google", "openrouter",
}

// integrationStatusResponse is the API response for a single integration
type integrationStatusResponse struct {
	ID          string                 `json:"id"`
	Status      string                 `json:"status"`
	ConnectedAt *time.Time             `json:"connected_at,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// List returns all integrations with their connection status.
//
// GET /api/v1/integrations
func (h *IntegrationHandler) List(c *fiber.Ctx) error {
	// For now, use a default user ID until auth is implemented
	userID := getUserID(c)

	// Fetch all stored integrations for this user
	var stored []models.Integration
	if err := database.DB.Where("user_id = ?", userID).Find(&stored).Error; err != nil {
		log.Printf("Failed to query integrations: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch integrations",
		})
	}

	// Build a map of stored integrations by their integration_id
	storedMap := make(map[string]*models.Integration)
	for i := range stored {
		storedMap[stored[i].IntegrationID] = &stored[i]
	}

	// Check if Kubernetes in-cluster is available (no credentials needed)
	k8sAvailable := k8s.IsAvailable()

	// Build response list — includes all known integrations, merged with stored status
	integrations := make([]integrationStatusResponse, 0, len(knownIntegrations))
	for _, id := range knownIntegrations {
		resp := integrationStatusResponse{
			ID:     id,
			Status: "disconnected",
		}
		// Kubernetes is auto-detected; no stored credentials required
		if id == "kubernetes" {
			if k8sAvailable {
				resp.Status = "connected"
			}
			integrations = append(integrations, resp)
			continue
		}
		if s, ok := storedMap[id]; ok {
			resp.Status = s.Status
			resp.ConnectedAt = s.ConnectedAt
			if s.Metadata != "" {
				_ = json.Unmarshal([]byte(s.Metadata), &resp.Metadata)
			}
		}
		integrations = append(integrations, resp)
	}

	return c.JSON(fiber.Map{
		"integrations": integrations,
	})
}

// Get returns details of a specific integration.
//
// GET /api/v1/integrations/:id
func (h *IntegrationHandler) Get(c *fiber.Ctx) error {
	integrationID := c.Params("id")
	userID := getUserID(c)

	var integration models.Integration
	result := database.DB.Where("user_id = ? AND integration_id = ?", userID, integrationID).First(&integration)

	if result.Error != nil {
		return c.JSON(integrationStatusResponse{
			ID:     integrationID,
			Status: "disconnected",
		})
	}

	resp := integrationStatusResponse{
		ID:          integrationID,
		Status:      integration.Status,
		ConnectedAt: integration.ConnectedAt,
	}
	if integration.Metadata != "" {
		_ = json.Unmarshal([]byte(integration.Metadata), &resp.Metadata)
	}

	return c.JSON(resp)
}

// Connect saves integration credentials after validating them.
//
// POST /api/v1/integrations/:id/connect
func (h *IntegrationHandler) Connect(c *fiber.Ctx) error {
	integrationID := c.Params("id")
	userID := getUserID(c)

	// Parse the envelope: { "data": "<base64-encoded credentials JSON>", "test_only": bool }
	var envelope struct {
		Data     string `json:"data"`
		TestOnly bool   `json:"test_only"`
	}
	if err := c.BodyParser(&envelope); err != nil || envelope.Data == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body — expected { data: string, test_only: bool }",
		})
	}

	// Base64-decode the credentials
	decoded, err := base64.StdEncoding.DecodeString(envelope.Data)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid base64-encoded credentials",
		})
	}

	var credentials map[string]string
	if err := json.Unmarshal(decoded, &credentials); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid credentials format",
		})
	}

	testOnly := envelope.TestOnly

	// Validate credentials against the provider
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	metadata, err := validateProviderCredentials(ctx, integrationID, credentials)
	if err != nil {
		return c.JSON(fiber.Map{
			"id":     integrationID,
			"status": "error",
			"error":  fmt.Sprintf("Credential validation failed: %v", err),
		})
	}

	// If test-only, return success without saving
	if testOnly {
		return c.JSON(fiber.Map{
			"id":       integrationID,
			"status":   "connected",
			"metadata": metadata,
		})
	}

	// Encrypt credentials before storing (AES-256-GCM)
	credJSON, _ := json.Marshal(credentials)
	encryptedCreds, err := crypto.Encrypt(credJSON)
	if err != nil {
		log.Printf("Failed to encrypt credentials: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to secure credentials",
		})
	}

	// Serialize metadata for storage
	metaJSON, _ := json.Marshal(metadata)

	now := time.Now()

	// Upsert: create or update the integration record
	var existing models.Integration
	result := database.DB.Where("user_id = ? AND integration_id = ?", userID, integrationID).First(&existing)

	if result.Error != nil {
		// Create new
		integration := models.Integration{
			UserID:        userID,
			IntegrationID: integrationID,
			Status:        "connected",
			Credentials:   encryptedCreds,
			Metadata:      string(metaJSON),
			ConnectedAt:   &now,
		}
		if err := database.DB.Create(&integration).Error; err != nil {
			log.Printf("Failed to create integration: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to save integration",
			})
		}
	} else {
		// Update existing
		existing.Status = "connected"
		existing.Credentials = encryptedCreds
		existing.Metadata = string(metaJSON)
		existing.ConnectedAt = &now
		if err := database.DB.Save(&existing).Error; err != nil {
			log.Printf("Failed to update integration: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to update integration",
			})
		}
	}

	return c.JSON(fiber.Map{
		"id":       integrationID,
		"status":   "connected",
		"metadata": metadata,
	})
}

// Disconnect removes credentials for an integration.
//
// POST /api/v1/integrations/:id/disconnect
func (h *IntegrationHandler) Disconnect(c *fiber.Ctx) error {
	integrationID := c.Params("id")
	userID := getUserID(c)

	// Delete the integration record
	result := database.DB.Where("user_id = ? AND integration_id = ?", userID, integrationID).Delete(&models.Integration{})
	if result.Error != nil {
		log.Printf("Failed to delete integration: %v", result.Error)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to disconnect integration",
		})
	}

	return c.JSON(fiber.Map{
		"id":     integrationID,
		"status": "disconnected",
	})
}

// TestConnection re-validates existing stored credentials.
//
// POST /api/v1/integrations/:id/test
func (h *IntegrationHandler) TestConnection(c *fiber.Ctx) error {
	integrationID := c.Params("id")
	userID := getUserID(c)

	// Fetch stored credentials
	var integration models.Integration
	result := database.DB.Where("user_id = ? AND integration_id = ?", userID, integrationID).First(&integration)
	if result.Error != nil {
		return c.JSON(fiber.Map{
			"id":     integrationID,
			"status": "error",
			"error":  "Integration not found — connect it first",
		})
	}

	// Decrypt and parse stored credentials (AES-256-GCM)
	decryptedJSON, err := crypto.Decrypt(integration.Credentials)
	if err != nil {
		log.Printf("Failed to decrypt credentials for %s: %v", integrationID, err)
		return c.JSON(fiber.Map{
			"id":     integrationID,
			"status": "error",
			"error":  "Failed to decrypt stored credentials",
		})
	}

	var credentials map[string]string
	if err := json.Unmarshal(decryptedJSON, &credentials); err != nil {
		return c.JSON(fiber.Map{
			"id":     integrationID,
			"status": "error",
			"error":  "Failed to read stored credentials",
		})
	}

	// Validate against the provider
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	start := time.Now()
	_, err = validateProviderCredentials(ctx, integrationID, credentials)
	latencyMs := time.Since(start).Milliseconds()

	if err != nil {
		// Update status to error in DB
		database.DB.Model(&integration).Update("status", "error")

		return c.JSON(fiber.Map{
			"id":         integrationID,
			"status":     "error",
			"error":      fmt.Sprintf("Connection test failed: %v", err),
			"latency_ms": latencyMs,
		})
	}

	return c.JSON(fiber.Map{
		"id":         integrationID,
		"status":     "connected",
		"latency_ms": latencyMs,
	})
}

// === GitHub-specific endpoints ===

// ListGitHubRepos lists GitHub repositories for the connected account.
//
// GET /api/v1/integrations/github/repos
func (h *IntegrationHandler) ListGitHubRepos(c *fiber.Ctx) error {
	client, err := getGitHubClient(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 20)

	repos, err := client.ListRepositories(c.Context(), page, perPage)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to list repositories: %v", err),
		})
	}

	return c.JSON(repos)
}

// GetGitHubDockerfile fetches a Dockerfile from a GitHub repository.
//
// GET /api/v1/integrations/github/repos/:owner/:repo/dockerfile
func (h *IntegrationHandler) GetGitHubDockerfile(c *fiber.Ctx) error {
	client, err := getGitHubClient(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	owner := c.Params("owner")
	repo := c.Params("repo")
	path := c.Query("path", "")
	ref := c.Query("ref", "")

	content, err := client.GetDockerfile(c.Context(), owner, repo, path, ref)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": fmt.Sprintf("Dockerfile not found: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"content": content,
		"path":    path,
	})
}

// ListGitHubContainerImages lists GHCR images.
//
// GET /api/v1/integrations/github/images
func (h *IntegrationHandler) ListGitHubContainerImages(c *fiber.Ctx) error {
	client, err := getGitHubClient(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	owner := c.Query("owner", "")
	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 20)

	if owner == "" {
		// Use the authenticated user
		owner, _ = client.ValidateCredentials(c.Context())
	}

	images, err := client.ListContainerImages(c.Context(), owner, page, perPage)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to list container images: %v", err),
		})
	}

	return c.JSON(images)
}

// CreateGitHubIssue creates a GitHub issue with optimization results.
//
// POST /api/v1/integrations/github/repos/:owner/:repo/issues
func (h *IntegrationHandler) CreateGitHubIssue(c *fiber.Ctx) error {
	client, err := getGitHubClient(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	owner := c.Params("owner")
	repo := c.Params("repo")

	var body struct {
		JobID  string   `json:"job_id"`
		Title  string   `json:"title"`
		Labels []string `json:"labels"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// TODO: Fetch report from DB using body.JobID and format as markdown
	reportSummary := fmt.Sprintf("Container image optimization report for job %s", body.JobID)
	recommendations := []string{"Report details will be populated when the analysis pipeline is implemented."}

	issue, err := client.CreateOptimizationIssue(c.Context(), owner, repo, reportSummary, recommendations)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to create issue: %v", err),
		})
	}

	return c.JSON(issue)
}

// === Docker Hub-specific endpoints ===

// ListDockerHubRepos lists Docker Hub repositories.
//
// GET /api/v1/integrations/docker/repos
func (h *IntegrationHandler) ListDockerHubRepos(c *fiber.Ctx) error {
	client, err := getDockerHubClient(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("page_size", 20)

	repos, err := client.ListRepositories(c.Context(), page, pageSize)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to list repositories: %v", err),
		})
	}

	return c.JSON(repos)
}

// ListDockerHubTags lists tags for a Docker Hub repository.
//
// GET /api/v1/integrations/docker/repos/:namespace/:repo/tags
func (h *IntegrationHandler) ListDockerHubTags(c *fiber.Ctx) error {
	client, err := getDockerHubClient(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	namespace := c.Params("namespace")
	repo := c.Params("repo")
	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("page_size", 20)

	tags, err := client.ListTags(c.Context(), namespace, repo, page, pageSize)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to list tags: %v", err),
		})
	}

	return c.JSON(tags)
}

// === Harbor-specific endpoints ===

// ListHarborProjects lists Harbor projects.
//
// GET /api/v1/integrations/harbor/projects
func (h *IntegrationHandler) ListHarborProjects(c *fiber.Ctx) error {
	client, err := getHarborClient(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("page_size", 20)

	projects, err := client.ListProjects(c.Context(), page, pageSize)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to list projects: %v", err),
		})
	}

	return c.JSON(projects)
}

// ListHarborArtifacts lists artifacts for a Harbor repository.
//
// GET /api/v1/integrations/harbor/projects/:project/repos/:repo/artifacts
func (h *IntegrationHandler) ListHarborArtifacts(c *fiber.Ctx) error {
	client, err := getHarborClient(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	project := c.Params("project")
	repo := c.Params("repo")
	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("page_size", 20)

	artifacts, err := client.ListArtifacts(c.Context(), project, repo, page, pageSize)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to list artifacts: %v", err),
		})
	}

	return c.JSON(artifacts)
}

// === Kubernetes-specific endpoints ===

// GetKubernetesStatus returns whether the app is running in-cluster and basic cluster info.
//
// GET /api/v1/integrations/kubernetes/status
func (h *IntegrationHandler) GetKubernetesStatus(c *fiber.Ctx) error {
	if !k8s.IsAvailable() {
		return c.JSON(fiber.Map{
			"id":        "kubernetes",
			"status":    "disconnected",
			"available": false,
			"message":   "Not running inside a Kubernetes cluster",
		})
	}

	client, err := k8s.NewInClusterClient()
	if err != nil {
		return c.JSON(fiber.Map{
			"id":        "kubernetes",
			"status":    "error",
			"available": false,
			"error":     err.Error(),
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	info, err := client.GetClusterInfo(ctx)
	if err != nil {
		return c.JSON(fiber.Map{
			"id":        "kubernetes",
			"status":    "error",
			"available": true,
			"error":     err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"id":        "kubernetes",
		"status":    "connected",
		"available": true,
		"metadata":  info,
	})
}

// ListKubernetesImages lists all container images running in the cluster.
//
// GET /api/v1/integrations/kubernetes/images
func (h *IntegrationHandler) ListKubernetesImages(c *fiber.Ctx) error {
	if !k8s.IsAvailable() {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "Not running inside a Kubernetes cluster",
		})
	}

	client, err := k8s.NewInClusterClient()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to create Kubernetes client: %v", err),
		})
	}

	namespace := c.Query("namespace", "")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	images, err := client.ListContainerImages(ctx, namespace)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to list container images: %v", err),
		})
	}

	return c.JSON(images)
}

// ListKubernetesNamespaces lists all namespaces in the cluster.
//
// GET /api/v1/integrations/kubernetes/namespaces
func (h *IntegrationHandler) ListKubernetesNamespaces(c *fiber.Ctx) error {
	if !k8s.IsAvailable() {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "Not running inside a Kubernetes cluster",
		})
	}

	client, err := k8s.NewInClusterClient()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to create Kubernetes client: %v", err),
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	namespaces, err := client.ListNamespaces(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to list namespaces: %v", err),
		})
	}

	return c.JSON(fiber.Map{"namespaces": namespaces})
}

// ─── Helper functions ────────────────────────────────────────────────────────

// getUserID extracts the user ID from the request context.
// Falls back to a default user for development until auth middleware is implemented.
func getUserID(c *fiber.Ctx) string {
	// TODO: Extract from JWT/session once auth middleware is in place
	userID := c.Get("X-User-ID")
	if userID == "" {
		userID = "default-user"
	}
	return userID
}

// getStoredCredentials retrieves and decrypts stored credentials for an integration.
func getStoredCredentials(c *fiber.Ctx, integrationID string) (map[string]string, error) {
	userID := getUserID(c)

	var integration models.Integration
	result := database.DB.Where("user_id = ? AND integration_id = ? AND status = ?", userID, integrationID, "connected").First(&integration)
	if result.Error != nil {
		return nil, fmt.Errorf("%s is not connected — set it up first", integrationID)
	}

	// Decrypt stored credentials (AES-256-GCM)
	decryptedJSON, err := crypto.Decrypt(integration.Credentials)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt stored credentials")
	}

	var credentials map[string]string
	if err := json.Unmarshal(decryptedJSON, &credentials); err != nil {
		return nil, fmt.Errorf("failed to read stored credentials")
	}

	return credentials, nil
}

// getGitHubClient creates a GitHub client from stored credentials.
func getGitHubClient(c *fiber.Ctx) (*github.Client, error) {
	creds, err := getStoredCredentials(c, "github")
	if err != nil {
		return nil, err
	}

	return github.NewClient(github.Config{
		PersonalAccessToken: creds["patToken"],
	}), nil
}

// getDockerHubClient creates a Docker Hub client from stored credentials.
func getDockerHubClient(c *fiber.Ctx) (*dockerhub.Client, error) {
	creds, err := getStoredCredentials(c, "docker")
	if err != nil {
		return nil, err
	}

	return dockerhub.NewClient(dockerhub.Config{
		PersonalAccessToken: creds["patToken"],
		Username:            creds["username"],
	}), nil
}

// getHarborClient creates a Harbor client from stored credentials.
func getHarborClient(c *fiber.Ctx) (*harbor.Client, error) {
	creds, err := getStoredCredentials(c, "harbor")
	if err != nil {
		return nil, err
	}

	return harbor.NewClient(harbor.Config{
		URL:      creds["url"],
		Username: creds["username"],
		Password: creds["password"],
	}), nil
}

// validateProviderCredentials validates credentials against the provider API.
// Returns metadata about the connection (e.g., username) on success.
func validateProviderCredentials(ctx context.Context, integrationID string, credentials map[string]string) (map[string]interface{}, error) {
	metadata := make(map[string]interface{})

	switch integrationID {
	case "github":
		client := github.NewClient(github.Config{
			PersonalAccessToken: credentials["patToken"],
		})
		username, err := client.ValidateCredentials(ctx)
		if err != nil {
			return nil, err
		}
		metadata["username"] = username

	case "docker":
		client := dockerhub.NewClient(dockerhub.Config{
			PersonalAccessToken: credentials["patToken"],
			Username:            credentials["username"],
		})
		username, err := client.ValidateCredentials(ctx)
		if err != nil {
			return nil, err
		}
		metadata["username"] = username

	case "harbor":
		client := harbor.NewClient(harbor.Config{
			URL:      credentials["url"],
			Username: credentials["username"],
			Password: credentials["password"],
		})
		version, err := client.ValidateCredentials(ctx)
		if err != nil {
			return nil, err
		}
		metadata["version"] = version
		metadata["url"] = credentials["url"]

	case "openai", "anthropic", "google", "openrouter":
		provider := ai.Provider(integrationID)
		client := ai.NewClient(ai.Config{
			Provider: provider,
			APIKey:   credentials["apiKey"],
		})
		providerName, err := client.ValidateCredentials(ctx)
		if err != nil {
			return nil, err
		}
		metadata["provider"] = providerName

	default:
		return nil, fmt.Errorf("unknown integration: %s", integrationID)
	}

	return metadata, nil
}
