package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Provider identifies which AI provider is being used
type Provider string

const (
	ProviderOpenAI     Provider = "openai"
	ProviderAnthropic  Provider = "anthropic"
	ProviderGoogleAI   Provider = "google"
	ProviderOpenRouter Provider = "openrouter"
)

// BaseURLs for each AI provider
var providerBaseURLs = map[Provider]string{
	ProviderOpenAI:     "https://api.openai.com/v1",
	ProviderAnthropic:  "https://api.anthropic.com/v1",
	ProviderGoogleAI:   "https://generativelanguage.googleapis.com/v1beta",
	ProviderOpenRouter: "https://openrouter.ai/api/v1",
}

// Config holds the configuration for an AI provider integration
type Config struct {
	Provider Provider `json:"provider"`
	APIKey   string   `json:"apiKey"`
}

// Client provides methods to interact with AI provider APIs.
type Client struct {
	config     Config
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a new AI provider client
func NewClient(config Config) *Client {
	baseURL, ok := providerBaseURLs[config.Provider]
	if !ok {
		baseURL = providerBaseURLs[ProviderOpenAI]
	}

	return &Client{
		config: config,
		httpClient: &http.Client{
			Transport: &apiKeyTransport{
				provider:  config.Provider,
				apiKey:    config.APIKey,
				transport: http.DefaultTransport,
			},
		},
		baseURL: baseURL,
	}
}

// apiKeyTransport adds the appropriate auth header for each AI provider
type apiKeyTransport struct {
	provider  Provider
	apiKey    string
	transport http.RoundTripper
}

func (t *apiKeyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	switch t.provider {
	case ProviderAnthropic:
		req.Header.Set("x-api-key", t.apiKey)
		req.Header.Set("anthropic-version", "2023-06-01")
	case ProviderGoogleAI:
		q := req.URL.Query()
		q.Set("key", t.apiKey)
		req.URL.RawQuery = q.Encode()
	default:
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.apiKey))
	}
	req.Header.Set("Content-Type", "application/json")
	return t.transport.RoundTrip(req)
}

// Model represents an AI model available from the provider
type Model struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Provider string `json:"provider"`
}

// ChatMessage represents a message in a chat completion request
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionRequest represents a request to the chat completions API
type ChatCompletionRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Temperature float64       `json:"temperature,omitempty"`
	Stream      bool          `json:"stream,omitempty"`
}

// ChatCompletionResponse represents the response from a chat completions API
type ChatCompletionResponse struct {
	ID      string `json:"id"`
	Model   string `json:"model"`
	Content string `json:"content"`
	Usage   Usage  `json:"usage"`
}

// Usage tracks token usage for billing/monitoring
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
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

// ValidateCredentials checks if the API key is valid by making a lightweight API call.
func (c *Client) ValidateCredentials(ctx context.Context) (string, error) {
	switch c.config.Provider {
	case ProviderOpenAI, ProviderOpenRouter:
		// GET /models — lightweight check that the key is valid
		url := c.baseURL + "/models"
		_, status, err := c.doRequest(ctx, http.MethodGet, url, nil)
		if err != nil {
			return "", fmt.Errorf("failed to validate: %w", err)
		}
		if status == http.StatusUnauthorized {
			return "", fmt.Errorf("invalid API key: 401 Unauthorized")
		}
		if status != http.StatusOK {
			return "", fmt.Errorf("unexpected status %d", status)
		}
		return string(c.config.Provider), nil

	case ProviderAnthropic:
		// POST /messages with a minimal request to check the key
		payload := map[string]interface{}{
			"model":      "claude-3-haiku-20240307",
			"max_tokens": 5,
			"messages": []map[string]string{
				{"role": "user", "content": "hi"},
			},
		}
		payloadJSON, _ := json.Marshal(payload)
		url := c.baseURL + "/messages"
		_, status, err := c.doRequest(ctx, http.MethodPost, url, bytes.NewReader(payloadJSON))
		if err != nil {
			return "", fmt.Errorf("failed to validate: %w", err)
		}
		if status == http.StatusUnauthorized {
			return "", fmt.Errorf("invalid API key: 401 Unauthorized")
		}
		// Both 200 (success) and 400 (bad model) mean the key is valid
		if status == http.StatusOK || status == http.StatusBadRequest {
			return string(c.config.Provider), nil
		}
		return "", fmt.Errorf("unexpected status %d", status)

	case ProviderGoogleAI:
		// GET /models — with API key in query param (handled by transport)
		url := c.baseURL + "/models"
		_, status, err := c.doRequest(ctx, http.MethodGet, url, nil)
		if err != nil {
			return "", fmt.Errorf("failed to validate: %w", err)
		}
		if status == http.StatusBadRequest || status == http.StatusForbidden {
			return "", fmt.Errorf("invalid API key: %d", status)
		}
		if status != http.StatusOK {
			return "", fmt.Errorf("unexpected status %d", status)
		}
		return string(c.config.Provider), nil

	default:
		return "", fmt.Errorf("unknown provider: %s", c.config.Provider)
	}
}

// ListModels returns available models from the AI provider.
func (c *Client) ListModels(ctx context.Context) ([]Model, error) {
	switch c.config.Provider {
	case ProviderAnthropic:
		// Anthropic doesn't have a list models endpoint — return known models
		return []Model{
			{ID: "claude-sonnet-4-20250514", Name: "Claude Sonnet 4", Provider: "anthropic"},
			{ID: "claude-3-5-haiku-20241022", Name: "Claude 3.5 Haiku", Provider: "anthropic"},
		}, nil

	default:
		// OpenAI, OpenRouter, Google AI all have /models
		url := c.baseURL + "/models"
		data, status, err := c.doRequest(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}
		if status != http.StatusOK {
			return nil, fmt.Errorf("unexpected status %d: %s", status, string(data))
		}

		var response struct {
			Data []struct {
				ID string `json:"id"`
			} `json:"data"`
			// Google uses "models" key
			Models []struct {
				Name string `json:"name"`
			} `json:"models"`
		}
		if err := json.Unmarshal(data, &response); err != nil {
			return nil, fmt.Errorf("failed to parse models: %w", err)
		}

		var models []Model
		for _, m := range response.Data {
			models = append(models, Model{ID: m.ID, Name: m.ID, Provider: string(c.config.Provider)})
		}
		for _, m := range response.Models {
			models = append(models, Model{ID: m.Name, Name: m.Name, Provider: string(c.config.Provider)})
		}
		return models, nil
	}
}

// ChatCompletion sends a chat completion request (non-streaming).
func (c *Client) ChatCompletion(ctx context.Context, req ChatCompletionRequest) (*ChatCompletionResponse, error) {
	switch c.config.Provider {
	case ProviderAnthropic:
		return c.anthropicChatCompletion(ctx, req)
	case ProviderGoogleAI:
		return c.googleChatCompletion(ctx, req)
	default:
		return c.openAIChatCompletion(ctx, req)
	}
}

// openAIChatCompletion handles OpenAI/OpenRouter format
func (c *Client) openAIChatCompletion(ctx context.Context, req ChatCompletionRequest) (*ChatCompletionResponse, error) {
	payloadJSON, _ := json.Marshal(req)
	url := c.baseURL + "/chat/completions"

	data, status, err := c.doRequest(ctx, http.MethodPost, url, bytes.NewReader(payloadJSON))
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("chat completion failed (status %d): %s", status, string(data))
	}

	var response struct {
		ID      string `json:"id"`
		Model   string `json:"model"`
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage Usage `json:"usage"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	content := ""
	if len(response.Choices) > 0 {
		content = response.Choices[0].Message.Content
	}

	return &ChatCompletionResponse{
		ID:      response.ID,
		Model:   response.Model,
		Content: content,
		Usage:   response.Usage,
	}, nil
}

// anthropicChatCompletion handles Anthropic's Messages API format
func (c *Client) anthropicChatCompletion(ctx context.Context, req ChatCompletionRequest) (*ChatCompletionResponse, error) {
	// Convert to Anthropic's format — separate system from messages
	var systemPrompt string
	var messages []map[string]string
	for _, m := range req.Messages {
		if m.Role == "system" {
			systemPrompt = m.Content
		} else {
			messages = append(messages, map[string]string{
				"role":    m.Role,
				"content": m.Content,
			})
		}
	}

	payload := map[string]interface{}{
		"model":      req.Model,
		"max_tokens": req.MaxTokens,
		"messages":   messages,
	}
	if systemPrompt != "" {
		payload["system"] = systemPrompt
	}
	if req.Temperature > 0 {
		payload["temperature"] = req.Temperature
	}

	payloadJSON, _ := json.Marshal(payload)
	url := c.baseURL + "/messages"

	data, status, err := c.doRequest(ctx, http.MethodPost, url, bytes.NewReader(payloadJSON))
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("chat completion failed (status %d): %s", status, string(data))
	}

	var response struct {
		ID      string `json:"id"`
		Model   string `json:"model"`
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	content := ""
	if len(response.Content) > 0 {
		content = response.Content[0].Text
	}

	return &ChatCompletionResponse{
		ID:      response.ID,
		Model:   response.Model,
		Content: content,
		Usage: Usage{
			PromptTokens:     response.Usage.InputTokens,
			CompletionTokens: response.Usage.OutputTokens,
			TotalTokens:      response.Usage.InputTokens + response.Usage.OutputTokens,
		},
	}, nil
}

// googleChatCompletion handles Google AI's generateContent format
func (c *Client) googleChatCompletion(ctx context.Context, req ChatCompletionRequest) (*ChatCompletionResponse, error) {
	// Convert messages to Google's format
	var parts []map[string]interface{}
	for _, m := range req.Messages {
		role := m.Role
		if role == "assistant" {
			role = "model"
		}
		parts = append(parts, map[string]interface{}{
			"role":  role,
			"parts": []map[string]string{{"text": m.Content}},
		})
	}

	payload := map[string]interface{}{
		"contents": parts,
	}
	payloadJSON, _ := json.Marshal(payload)

	url := fmt.Sprintf("%s/models/%s:generateContent", c.baseURL, req.Model)
	data, status, err := c.doRequest(ctx, http.MethodPost, url, bytes.NewReader(payloadJSON))
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("chat completion failed (status %d): %s", status, string(data))
	}

	var response struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	content := ""
	if len(response.Candidates) > 0 && len(response.Candidates[0].Content.Parts) > 0 {
		content = response.Candidates[0].Content.Parts[0].Text
	}

	return &ChatCompletionResponse{
		Model:   req.Model,
		Content: content,
	}, nil
}

// ChatCompletionStream sends a streaming chat completion request.
// Returns a channel that receives response chunks as they arrive.
func (c *Client) ChatCompletionStream(ctx context.Context, req ChatCompletionRequest) (<-chan string, error) {
	// TODO: Implement streaming for each provider
	// For now, fall back to non-streaming and send it all at once
	resp, err := c.ChatCompletion(ctx, req)
	if err != nil {
		return nil, err
	}

	ch := make(chan string, 1)
	go func() {
		defer close(ch)
		ch <- resp.Content
	}()

	return ch, nil
}
