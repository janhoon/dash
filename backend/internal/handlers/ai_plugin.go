package handlers

import "github.com/aceobservability/ace/backend/pkg/llm"

type (
	AIProvider  = llm.AIProvider
	AIModel     = llm.AIModel
	ChatRequest = llm.ChatRequest
	LLMConfig   = llm.LLMConfig
	LLMFactory  = llm.LLMFactory
)

var ErrUnknownProviderType = llm.ErrUnknownProviderType

func RegisterLLM(providerType string, factory LLMFactory) {
	llm.RegisterLLM(providerType, factory)
}

func unregisterLLM(providerType string) {
	llm.UnregisterLLM(providerType)
}

func newLLMProvider(providerType string, cfg LLMConfig) (AIProvider, error) {
	return llm.New(providerType, cfg)
}

func requireKnownLLMType(providerType string) error {
	return llm.RequireKnown(providerType)
}

func init() {
	openaiCompat := func(cfg LLMConfig) (AIProvider, error) {
		return NewOpenAICompatibleProvider(cfg.BaseURL, cfg.APIKey, cfg.DisplayName), nil
	}
	RegisterLLM("openai", openaiCompat)
	RegisterLLM("openrouter", openaiCompat)
	RegisterLLM("ollama", openaiCompat)
	RegisterLLM("custom", openaiCompat)
	RegisterLLM("anthropic", NewAnthropic)
	// Copilot's live chat path remains provider_id=="copilot" (user GH token).
	// Registering the type means a stored provider_type=copilot does not
	// impersonate OpenAI-compat. LLMConfig.APIKey is ciphertext; the factory
	// must not receive a decrypted key (CopilotProvider decrypts on use).
	RegisterLLM("copilot", func(cfg LLMConfig) (AIProvider, error) {
		return &CopilotProvider{EncryptedGHToken: cfg.APIKey}, nil
	})
}
