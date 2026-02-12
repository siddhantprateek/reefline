package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestValidateCredentials(t *testing.T) {
	// Helper to restore base URLs after test
	originalBaseURLs := make(map[Provider]string)
	for k, v := range providerBaseURLs {
		originalBaseURLs[k] = v
	}
	defer func() {
		providerBaseURLs = originalBaseURLs
	}()

	tests := []struct {
		name           string
		provider       Provider
		apiKey         string
		mockHandler    func(w http.ResponseWriter, r *http.Request)
		expectedError  bool
		expectedResult string
	}{
		{
			name:     "OpenAI Success",
			provider: ProviderOpenAI,
			apiKey:   "test-openai-key",
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/models" {
					t.Errorf("expected path /models, got %s", r.URL.Path)
				}
				if r.Header.Get("Authorization") != "Bearer test-openai-key" {
					t.Errorf("expected auth header Bearer test-openai-key, got %s", r.Header.Get("Authorization"))
				}
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"data": []}`))
			},
			expectedResult: "openai",
			expectedError:  false,
		},
		{
			name:     "OpenAI Invalid Key",
			provider: ProviderOpenAI,
			apiKey:   "invalid-key",
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
			},
			expectedError: true,
		},
		{
			name:     "Anthropic Success",
			provider: ProviderAnthropic,
			apiKey:   "test-anthropic-key",
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/messages" {
					t.Errorf("expected path /messages, got %s", r.URL.Path)
				}
				if r.Header.Get("x-api-key") != "test-anthropic-key" {
					t.Errorf("expected x-api-key test-anthropic-key, got %s", r.Header.Get("x-api-key"))
				}
				if r.Header.Get("anthropic-version") != "2023-06-01" {
					t.Errorf("expected version 2023-06-01, got %s", r.Header.Get("anthropic-version"))
				}
				// Mock a success response (or even a bad request due to dummy body is considered "auth success" in our logic)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"content": []}`))
			},
			expectedResult: "anthropic",
			expectedError:  false,
		},
		{
			name:     "Google AI Success",
			provider: ProviderGoogleAI,
			apiKey:   "test-google-key",
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/models" {
					t.Errorf("expected path /models, got %s", r.URL.Path)
				}
				// Check query param
				if r.URL.Query().Get("key") != "test-google-key" {
					t.Errorf("expected key query param test-google-key, got %s", r.URL.Query().Get("key"))
				}
				w.WriteHeader(http.StatusOK)
			},
			expectedResult: "google",
			expectedError:  false,
		},
		{
			name:     "OpenRouter Success",
			provider: ProviderOpenRouter,
			apiKey:   "test-openrouter-key",
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/models" {
					t.Errorf("expected path /models, got %s", r.URL.Path)
				}
				if r.Header.Get("Authorization") != "Bearer test-openrouter-key" {
					t.Errorf("expected auth header Bearer test-openrouter-key, got %s", r.Header.Get("Authorization"))
				}
				w.WriteHeader(http.StatusOK)
			},
			expectedResult: "openrouter",
			expectedError:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mock server
			server := httptest.NewServer(http.HandlerFunc(tc.mockHandler))
			defer server.Close()

			// Override base URL for the provider
			providerBaseURLs[tc.provider] = server.URL

			client := NewClient(Config{
				Provider: tc.provider,
				APIKey:   tc.apiKey,
			})

			result, err := client.ValidateCredentials(context.Background())

			if tc.expectedError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result != tc.expectedResult {
					t.Errorf("expected result %s, got %s", tc.expectedResult, result)
				}
			}
		})
	}
}

func TestListModels(t *testing.T) {
	originalBaseURLs := make(map[Provider]string)
	for k, v := range providerBaseURLs {
		originalBaseURLs[k] = v
	}
	defer func() {
		providerBaseURLs = originalBaseURLs
	}()

	// Test OpenAI model listing
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"data": [
				{"id": "gpt-4"}
			]
		}`))
	}))
	defer server.Close()

	providerBaseURLs[ProviderOpenAI] = server.URL

	client := NewClient(Config{Provider: ProviderOpenAI, APIKey: "test"})
	models, err := client.ListModels(context.Background())
	if err != nil {
		t.Fatalf("ListModels failed: %v", err)
	}

	if len(models) == 0 {
		t.Error("expected models, got 0")
	}
	if models[0].ID != "gpt-4" {
		t.Errorf("expected model gpt-4, got %s", models[0].ID)
	}
}

func TestChatCompletion_Anthropic(t *testing.T) {
	originalBaseURLs := make(map[Provider]string)
	for k, v := range providerBaseURLs {
		originalBaseURLs[k] = v
	}
	defer func() {
		providerBaseURLs = originalBaseURLs
	}()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify conversion of messages
		var payload map[string]interface{}
		json.NewDecoder(r.Body).Decode(&payload)

		// Check system prompt extraction
		if payload["system"] != "System prompt" {
			t.Errorf("expected system prompt 'System prompt', got %v", payload["system"])
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"id": "msg_123",
			"content": [{"text": "Hello world"}],
			"usage": {"input_tokens": 10, "output_tokens": 5}
		}`))
	}))
	defer server.Close()

	providerBaseURLs[ProviderAnthropic] = server.URL

	client := NewClient(Config{Provider: ProviderAnthropic, APIKey: "test"})
	resp, err := client.ChatCompletion(context.Background(), ChatCompletionRequest{
		Model: "claude-3-opus",
		Messages: []ChatMessage{
			{Role: "system", Content: "System prompt"},
			{Role: "user", Content: "Hi"},
		},
	})

	if err != nil {
		t.Fatalf("ChatCompletion failed: %v", err)
	}
	if resp.Content != "Hello world" {
		t.Errorf("expected 'Hello world', got '%s'", resp.Content)
	}
}
