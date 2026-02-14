package flows

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/siddhantprateek/reefline/pkg/crypto"
	"github.com/siddhantprateek/reefline/pkg/database"
	"github.com/siddhantprateek/reefline/pkg/models"
)

// ─── Provider Constants ──────────────────────────────────────────────────────

// Provider identifies which AI provider backend is being used.
type Provider string

const (
	ProviderOpenAI     Provider = "openai"
	ProviderAnthropic  Provider = "anthropic"
	ProviderGoogle     Provider = "google"
	ProviderOpenRouter Provider = "openrouter"
)

// OpenAI-compatible base URLs for each provider.
// All providers now expose an OpenAI-compatible chat completions endpoint,
// so we can use the openai-go SDK for all of them.
var providerBaseURLs = map[Provider]string{
	ProviderOpenAI:     "https://api.openai.com/v1/",
	ProviderAnthropic:  "https://api.anthropic.com/v1/",
	ProviderGoogle:     "https://generativelanguage.googleapis.com/v1beta/openai/",
	ProviderOpenRouter: "https://openrouter.ai/api/v1/",
}

// ─── Provider Client ─────────────────────────────────────────────────────────

// providerClient wraps an openai-go client bound to a specific provider.
type providerClient struct {
	provider Provider
	client   *openai.Client
	model    ModelInfo
}

// ─── Multi-Provider Client ───────────────────────────────────────────────────

// Client is the BYOK (Bring Your Own Key) multi-provider AI client.
//
// It discovers which providers are configured by reading encrypted API keys
// from the database. Each configured provider gets its own openai-go client
// pointed at the provider's OpenAI-compatible base URL.
//
// Usage:
//
//	client, err := flows.NewClient("user-123")
//	resp, err := client.ChatCompletion(ctx, flows.ChatRequest{
//	    Messages: []flows.Message{{Role: "user", Content: "Hello"}},
//	})
type Client struct {
	mu sync.RWMutex

	userID string

	// providers maps each configured provider to its client.
	providers map[Provider]*providerClient

	// active is the currently selected provider.
	active Provider

	// fallback is the fallback provider if the active one fails.
	fallback Provider
}

// NewClient creates a multi-provider AI client for the given user.
// It reads all configured AI integrations from the database, decrypts API keys,
// and initializes an openai-go client for each configured provider.
//
// Provider selection logic:
//   - If only one provider is configured, it becomes the active provider.
//   - If multiple providers are configured, the first one found in priority order
//     (OpenAI → Anthropic → Google → OpenRouter) becomes active, and the second
//     becomes the fallback.
//   - Users can override this via SetActiveProvider / SetFallbackProvider.
func NewClient(userID string) (*Client, error) {
	c := &Client{
		userID:    userID,
		providers: make(map[Provider]*providerClient),
	}

	if err := c.loadProviders(); err != nil {
		return nil, fmt.Errorf("failed to initialize AI client: %w", err)
	}

	if len(c.providers) == 0 {
		return nil, fmt.Errorf("no AI providers configured — connect at least one AI integration (OpenAI, Anthropic, Google, or OpenRouter)")
	}

	// Auto-select active and fallback providers
	c.autoSelectProviders()

	return c, nil
}

// loadProviders reads stored AI integrations from the database and creates
// an openai-go client for each configured provider.
func (c *Client) loadProviders() error {
	aiProviderIDs := []string{"openai", "anthropic", "google", "openrouter"}

	var integrations []models.Integration
	result := database.DB.Where(
		"user_id = ? AND integration_id IN ? AND status = ?",
		c.userID, aiProviderIDs, "connected",
	).Find(&integrations)

	if result.Error != nil {
		return fmt.Errorf("failed to query integrations: %w", result.Error)
	}

	for _, integration := range integrations {
		provider := Provider(integration.IntegrationID)

		apiKey, err := decryptAPIKey(integration.Credentials)
		if err != nil {
			log.Printf("[flows] failed to decrypt %s credentials for user %s: %v", provider, c.userID, err)
			continue
		}

		baseURL, ok := providerBaseURLs[provider]
		if !ok {
			log.Printf("[flows] unknown provider %s — skipping", provider)
			continue
		}

		// Create an openai-go client pointed at this provider's base URL
		oaiClient := openai.NewClient(
			option.WithAPIKey(apiKey),
			option.WithBaseURL(baseURL),
		)

		model := defaultModels[provider]

		c.providers[provider] = &providerClient{
			provider: provider,
			client:   &oaiClient,
			model:    model,
		}

		log.Printf("[flows] initialized %s provider for user %s (model: %s)", provider, c.userID, model.ID)
	}

	return nil
}

// autoSelectProviders picks the active and fallback providers based on priority.
func (c *Client) autoSelectProviders() {
	priority := []Provider{ProviderOpenAI, ProviderAnthropic, ProviderGoogle, ProviderOpenRouter}

	var configured []Provider
	for _, p := range priority {
		if _, ok := c.providers[p]; ok {
			configured = append(configured, p)
		}
	}

	if len(configured) > 0 {
		c.active = configured[0]
	}
	if len(configured) > 1 {
		c.fallback = configured[1]
	}
}

// ─── Configuration Methods ───────────────────────────────────────────────────

// SetActiveProvider changes the currently active AI provider.
func (c *Client) SetActiveProvider(provider Provider) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.providers[provider]; !ok {
		return fmt.Errorf("provider %s is not configured", provider)
	}
	c.active = provider
	return nil
}

// SetFallbackProvider changes the fallback AI provider.
func (c *Client) SetFallbackProvider(provider Provider) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.providers[provider]; !ok {
		return fmt.Errorf("provider %s is not configured", provider)
	}
	if provider == c.active {
		return fmt.Errorf("fallback provider cannot be the same as the active provider")
	}
	c.fallback = provider
	return nil
}

// SetModel changes the model used for a specific provider.
func (c *Client) SetModel(provider Provider, modelID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	pc, ok := c.providers[provider]
	if !ok {
		return fmt.Errorf("provider %s is not configured", provider)
	}

	// Validate the model exists in our registry
	for _, m := range availableModels[provider] {
		if m.ID == modelID {
			pc.model = m
			return nil
		}
	}

	// Allow custom/unknown models — the provider API will validate
	pc.model = ModelInfo{
		ID:       modelID,
		Name:     modelID,
		Provider: provider,
	}
	return nil
}

// ActiveProvider returns the current active provider.
func (c *Client) ActiveProvider() Provider {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.active
}

// FallbackProvider returns the current fallback provider.
func (c *Client) FallbackProvider() Provider {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.fallback
}

// ConfiguredProviders returns a list of all configured providers.
func (c *Client) ConfiguredProviders() []Provider {
	c.mu.RLock()
	defer c.mu.RUnlock()

	providers := make([]Provider, 0, len(c.providers))
	for p := range c.providers {
		providers = append(providers, p)
	}
	return providers
}

// AvailableModels returns models available for a specific provider.
func (c *Client) AvailableModels(provider Provider) []ModelInfo {
	if models, ok := availableModels[provider]; ok {
		return models
	}
	return nil
}

// ─── Chat Completion ─────────────────────────────────────────────────────────

// Message represents a single message in a conversation.
type Message struct {
	Role    string `json:"role"`    // "system", "user", "assistant"
	Content string `json:"content"` // message text
}

// ChatRequest contains the parameters for a chat completion request.
type ChatRequest struct {
	Messages    []Message `json:"messages"`
	Model       string    `json:"model,omitempty"`       // override the default model
	Provider    Provider  `json:"provider,omitempty"`    // override the active provider
	MaxTokens   int64     `json:"max_tokens,omitempty"`  // max tokens in the response
	Temperature float64   `json:"temperature,omitempty"` // sampling temperature (0.0-2.0)
	TopP        float64   `json:"top_p,omitempty"`       // nucleus sampling parameter
}

// ChatResponse contains the result of a chat completion.
type ChatResponse struct {
	Content  string    `json:"content"`
	Model    string    `json:"model"`
	Provider Provider  `json:"provider"`
	Usage    UsageInfo `json:"usage"`
}

// UsageInfo tracks token usage for billing and monitoring.
type UsageInfo struct {
	PromptTokens     int64 `json:"prompt_tokens"`
	CompletionTokens int64 `json:"completion_tokens"`
	TotalTokens      int64 `json:"total_tokens"`
}

// ChatCompletion sends a chat completion request to the active (or specified) provider.
// If the active provider fails and a fallback is configured, it automatically retries
// with the fallback provider.
func (c *Client) ChatCompletion(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	c.mu.RLock()
	activeProvider := c.active
	fallbackProvider := c.fallback
	c.mu.RUnlock()

	// Allow request-level provider override
	if req.Provider != "" {
		activeProvider = req.Provider
	}

	// Try the active provider
	resp, err := c.chatCompletionWithProvider(ctx, activeProvider, req)
	if err == nil {
		return resp, nil
	}

	log.Printf("[flows] %s provider failed: %v", activeProvider, err)

	// Try fallback if available and different from active
	if fallbackProvider != "" && fallbackProvider != activeProvider {
		log.Printf("[flows] falling back to %s provider", fallbackProvider)
		resp, fallbackErr := c.chatCompletionWithProvider(ctx, fallbackProvider, req)
		if fallbackErr == nil {
			return resp, nil
		}
		log.Printf("[flows] fallback %s also failed: %v", fallbackProvider, fallbackErr)
		return nil, fmt.Errorf("all providers failed — primary (%s): %w, fallback (%s): %v",
			activeProvider, err, fallbackProvider, fallbackErr)
	}

	return nil, fmt.Errorf("provider %s failed: %w", activeProvider, err)
}

// chatCompletionWithProvider executes a chat completion with a specific provider.
func (c *Client) chatCompletionWithProvider(ctx context.Context, provider Provider, req ChatRequest) (*ChatResponse, error) {
	c.mu.RLock()
	pc, ok := c.providers[provider]
	c.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("provider %s is not configured", provider)
	}

	// Determine which model to use
	modelID := pc.model.ID
	if req.Model != "" {
		modelID = req.Model
	}

	// Convert our Message type to openai-go params
	messages := make([]openai.ChatCompletionMessageParamUnion, 0, len(req.Messages))
	for _, msg := range req.Messages {
		switch msg.Role {
		case "system":
			messages = append(messages, openai.SystemMessage(msg.Content))
		case "user":
			messages = append(messages, openai.UserMessage(msg.Content))
		case "assistant":
			messages = append(messages, openai.AssistantMessage(msg.Content))
		default:
			messages = append(messages, openai.UserMessage(msg.Content))
		}
	}

	// Build the completion request
	params := openai.ChatCompletionNewParams{
		Model:    modelID,
		Messages: messages,
	}

	if req.MaxTokens > 0 {
		params.MaxTokens = openai.Int(req.MaxTokens)
	}
	if req.Temperature > 0 {
		params.Temperature = openai.Float(req.Temperature)
	}
	if req.TopP > 0 {
		params.TopP = openai.Float(req.TopP)
	}

	// Execute the completion
	completion, err := pc.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("chat completion failed: %w", err)
	}

	// Extract content from the response
	content := ""
	if len(completion.Choices) > 0 {
		content = completion.Choices[0].Message.Content
	}

	// Build usage info
	usage := UsageInfo{}
	if completion.Usage.PromptTokens > 0 || completion.Usage.CompletionTokens > 0 {
		usage.PromptTokens = completion.Usage.PromptTokens
		usage.CompletionTokens = completion.Usage.CompletionTokens
		usage.TotalTokens = completion.Usage.TotalTokens
	}

	return &ChatResponse{
		Content:  content,
		Model:    completion.Model,
		Provider: provider,
		Usage:    usage,
	}, nil
}

// ─── Streaming ───────────────────────────────────────────────────────────────

// StreamHandler is a callback invoked for each chunk during streaming.
type StreamHandler func(chunk string)

// ChatCompletionStream sends a streaming chat completion request and calls the
// handler for each content chunk as it arrives.
func (c *Client) ChatCompletionStream(ctx context.Context, req ChatRequest, handler StreamHandler) (*ChatResponse, error) {
	c.mu.RLock()
	activeProvider := c.active
	c.mu.RUnlock()

	if req.Provider != "" {
		activeProvider = req.Provider
	}

	return c.chatStreamWithProvider(ctx, activeProvider, req, handler)
}

// chatStreamWithProvider executes a streaming chat completion with a specific provider.
func (c *Client) chatStreamWithProvider(ctx context.Context, provider Provider, req ChatRequest, handler StreamHandler) (*ChatResponse, error) {
	c.mu.RLock()
	pc, ok := c.providers[provider]
	c.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("provider %s is not configured", provider)
	}

	modelID := pc.model.ID
	if req.Model != "" {
		modelID = req.Model
	}

	messages := make([]openai.ChatCompletionMessageParamUnion, 0, len(req.Messages))
	for _, msg := range req.Messages {
		switch msg.Role {
		case "system":
			messages = append(messages, openai.SystemMessage(msg.Content))
		case "user":
			messages = append(messages, openai.UserMessage(msg.Content))
		case "assistant":
			messages = append(messages, openai.AssistantMessage(msg.Content))
		default:
			messages = append(messages, openai.UserMessage(msg.Content))
		}
	}

	params := openai.ChatCompletionNewParams{
		Model:    modelID,
		Messages: messages,
	}

	if req.MaxTokens > 0 {
		params.MaxTokens = openai.Int(req.MaxTokens)
	}
	if req.Temperature > 0 {
		params.Temperature = openai.Float(req.Temperature)
	}

	stream := pc.client.Chat.Completions.NewStreaming(ctx, params)

	// Use the accumulator to track progressive deltas
	acc := openai.ChatCompletionAccumulator{}
	for stream.Next() {
		chunk := stream.Current()
		acc.AddChunk(chunk)

		// Extract and emit any content delta
		if len(chunk.Choices) > 0 {
			delta := chunk.Choices[0].Delta.Content
			if delta != "" && handler != nil {
				handler(delta)
			}
		}
	}

	if err := stream.Err(); err != nil {
		return nil, fmt.Errorf("streaming failed: %w", err)
	}

	// Build final response from accumulated result
	content := ""
	if len(acc.Choices) > 0 {
		content = acc.Choices[0].Message.Content
	}

	return &ChatResponse{
		Content:  content,
		Model:    acc.Model,
		Provider: provider,
		Usage: UsageInfo{
			PromptTokens:     acc.Usage.PromptTokens,
			CompletionTokens: acc.Usage.CompletionTokens,
			TotalTokens:      acc.Usage.TotalTokens,
		},
	}, nil
}

// ─── Provider Info (for settings UI) ─────────────────────────────────────────

// ProviderStatus describes the configuration status of a provider.
type ProviderStatus struct {
	Provider   Provider  `json:"provider"`
	Configured bool      `json:"configured"`
	IsActive   bool      `json:"is_active"`
	IsFallback bool      `json:"is_fallback"`
	Model      ModelInfo `json:"model"`
}

// Status returns the configuration status of all providers.
func (c *Client) Status() []ProviderStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()

	allProviders := []Provider{ProviderOpenAI, ProviderAnthropic, ProviderGoogle, ProviderOpenRouter}
	statuses := make([]ProviderStatus, 0, len(allProviders))

	for _, p := range allProviders {
		status := ProviderStatus{
			Provider:   p,
			IsActive:   p == c.active,
			IsFallback: p == c.fallback,
		}
		if pc, ok := c.providers[p]; ok {
			status.Configured = true
			status.Model = pc.model
		}
		statuses = append(statuses, status)
	}

	return statuses
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

// decryptAPIKey decrypts stored credentials and extracts the API key.
func decryptAPIKey(encryptedCreds string) (string, error) {
	decryptedJSON, err := crypto.Decrypt(encryptedCreds)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt credentials: %w", err)
	}

	var credentials map[string]string
	if err := json.Unmarshal(decryptedJSON, &credentials); err != nil {
		return "", fmt.Errorf("failed to parse credentials: %w", err)
	}

	apiKey, ok := credentials["apiKey"]
	if !ok || apiKey == "" {
		return "", fmt.Errorf("API key not found in credentials")
	}

	return apiKey, nil
}
