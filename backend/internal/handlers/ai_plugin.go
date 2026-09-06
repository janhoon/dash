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

func init() {
	openaiCompat := func(cfg LLMConfig) (AIProvider, error) {
		return NewOpenAICompatibleProvider(cfg.BaseURL, cfg.APIKey, cfg.DisplayName), nil
	}
	llm.RegisterLLM("openai", openaiCompat)
	llm.RegisterLLM("openrouter", openaiCompat)
	llm.RegisterLLM("ollama", openaiCompat)
	llm.RegisterLLM("custom", openaiCompat)
	llm.RegisterLLM("anthropic", NewAnthropic)
	// Copilot's live chat path remains provider_id=="copilot" (user GH token).
	// Registering the type means a stored provider_type=copilot does not
	// impersonate OpenAI-compat. LLMConfig.APIKey is ciphertext; the factory
	// must not receive a decrypted key (CopilotProvider decrypts on use).
	llm.RegisterLLM("copilot", func(cfg LLMConfig) (AIProvider, error) {
		return &CopilotProvider{EncryptedGHToken: cfg.APIKey}, nil
	})
}
