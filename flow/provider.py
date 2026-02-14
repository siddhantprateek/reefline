from dataclasses import dataclass
from openai import AsyncOpenAI

PROVIDER_BASE_URLS: dict[str, str] = {
    "openai":     "https://api.openai.com/v1",
    "anthropic":  "https://api.anthropic.com/v1",
    "google":     "https://generativelanguage.googleapis.com/v1beta/openai",
    "openrouter": "https://openrouter.ai/api/v1",
}

DEFAULT_MODELS: dict[str, str] = {
    "openai":     "gpt-5-mini",
    "anthropic":  "claude-sonnet-4-20250514",
    "google":     "gemini-2.0-flash",
    "openrouter": "openai/gpt-4o",
}


@dataclass
class ProviderConfig:
    provider: str
    api_key: str
    model_id: str

    @classmethod
    def from_db_row(cls, row: dict) -> "ProviderConfig":
        provider = row["integration_id"]  # e.g. "openai"
        model_id = row.get("model_id") or DEFAULT_MODELS.get(provider, "gpt-5-mini")
        return cls(provider=provider, api_key=row["api_key"], model_id=model_id)

    @property
    def base_url(self) -> str:
        return PROVIDER_BASE_URLS.get(self.provider, PROVIDER_BASE_URLS["openai"])

    def openai_client(self) -> AsyncOpenAI:
        """All providers are accessed via the OpenAI-compatible client with the provider's base URL."""
        return AsyncOpenAI(api_key=self.api_key, base_url=self.base_url)
