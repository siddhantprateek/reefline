package flows

// ─── Model Registry ──────────────────────────────────────────────────────────

// ModelInfo describes an AI model available from a provider.
type ModelInfo struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Provider Provider `json:"provider"`
}

// Default models per provider — used when the user hasn't explicitly configured a model.
var defaultModels = map[Provider]ModelInfo{
	ProviderOpenAI: {
		ID:       "gpt-4o",
		Name:     "GPT-4o",
		Provider: ProviderOpenAI,
	},
	ProviderAnthropic: {
		ID:       "claude-sonnet-4-20250514",
		Name:     "Claude Sonnet 4",
		Provider: ProviderAnthropic,
	},
	ProviderGoogle: {
		ID:       "gemini-2.0-flash",
		Name:     "Gemini 2.0 Flash",
		Provider: ProviderGoogle,
	},
	ProviderOpenRouter: {
		ID:       "openai/gpt-4o",
		Name:     "GPT-4o (via OpenRouter)",
		Provider: ProviderOpenRouter,
	},
}

// availableModels lists the well-known models per provider.
var availableModels = map[Provider][]ModelInfo{
	ProviderOpenAI: {
		{ID: "gpt-4o", Name: "GPT-4o", Provider: ProviderOpenAI},
		{ID: "gpt-4o-mini", Name: "GPT-4o Mini", Provider: ProviderOpenAI},
		{ID: "gpt-4.1", Name: "GPT-4.1", Provider: ProviderOpenAI},
		{ID: "gpt-4.1-mini", Name: "GPT-4.1 Mini", Provider: ProviderOpenAI},
		{ID: "o3-mini", Name: "o3-mini", Provider: ProviderOpenAI},
	},
	ProviderAnthropic: {
		{ID: "claude-sonnet-4-20250514", Name: "Claude Sonnet 4", Provider: ProviderAnthropic},
		{ID: "claude-3-5-haiku-20241022", Name: "Claude 3.5 Haiku", Provider: ProviderAnthropic},
	},
	ProviderGoogle: {
		{ID: "gemini-2.0-flash", Name: "Gemini 2.0 Flash", Provider: ProviderGoogle},
		{ID: "gemini-2.5-pro-preview-06-05", Name: "Gemini 2.5 Pro", Provider: ProviderGoogle},
		{ID: "gemini-2.5-flash-preview-05-20", Name: "Gemini 2.5 Flash", Provider: ProviderGoogle},
	},
	ProviderOpenRouter: {
		{ID: "openai/gpt-4o", Name: "GPT-4o (OR)", Provider: ProviderOpenRouter},
		{ID: "anthropic/claude-sonnet-4-20250514", Name: "Claude Sonnet 4 (OR)", Provider: ProviderOpenRouter},
		{ID: "google/gemini-2.0-flash-001", Name: "Gemini 2.0 Flash (OR)", Provider: ProviderOpenRouter},
	},
}
