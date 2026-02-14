package flows

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
