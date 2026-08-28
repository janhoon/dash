package handlers

import (
	"errors"
	"fmt"
	"sync"
)

const pluginKindLLM = "llm"

// LLMConfig is the in-process config passed to an LLM plugin factory.
type LLMConfig struct {
	BaseURL     string
	APIKey      string
	DisplayName string
}

// LLMFactory constructs an AIProvider from LLMConfig.
type LLMFactory func(LLMConfig) (AIProvider, error)

// ErrUnknownProviderType is returned when ai_providers.provider_type has no
// registered factory. Callers must fail closed: never fall through to
// OpenAI-compat.
var ErrUnknownProviderType = errors.New("unknown provider_type")

// pluginRegistry is kind → id → factory. v1 wires kind=llm only.
var pluginRegistry = struct {
	mu sync.RWMutex
	m  map[string]map[string]LLMFactory
}{
	m: map[string]map[string]LLMFactory{
		pluginKindLLM: {},
	},
}

// RegisterLLM records a factory for provider_type under kind=llm.
func RegisterLLM(providerType string, factory LLMFactory) {
	if providerType == "" {
		panic("RegisterLLM: empty provider_type")
	}
	if factory == nil {
		panic("RegisterLLM: nil factory")
	}
	pluginRegistry.mu.Lock()
	defer pluginRegistry.mu.Unlock()
	pluginRegistry.m[pluginKindLLM][providerType] = factory
}

func lookupLLM(providerType string) (LLMFactory, bool) {
	pluginRegistry.mu.RLock()
	defer pluginRegistry.mu.RUnlock()
	kind, ok := pluginRegistry.m[pluginKindLLM]
	if !ok {
		return nil, false
	}
	f, ok := kind[providerType]
	return f, ok
}

func unregisterLLM(providerType string) {
	pluginRegistry.mu.Lock()
	defer pluginRegistry.mu.Unlock()
	delete(pluginRegistry.m[pluginKindLLM], providerType)
}

func newLLMProvider(providerType string, cfg LLMConfig) (AIProvider, error) {
	factory, ok := lookupLLM(providerType)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrUnknownProviderType, providerType)
	}
	return factory(cfg)
}

func requireKnownLLMType(providerType string) error {
	if _, ok := lookupLLM(providerType); !ok {
		return fmt.Errorf("%w: %s", ErrUnknownProviderType, providerType)
	}
	return nil
}

func init() {
	openaiCompat := func(cfg LLMConfig) (AIProvider, error) {
		return NewOpenAICompatibleProvider(cfg.BaseURL, cfg.APIKey, cfg.DisplayName), nil
	}
	RegisterLLM("openai", openaiCompat)
	RegisterLLM("openrouter", openaiCompat)
	RegisterLLM("ollama", openaiCompat)
	RegisterLLM("custom", openaiCompat)
	// Copilot's live chat path remains provider_id=="copilot" (user GH token).
	// Registering the type means a stored provider_type=copilot does not
	// impersonate OpenAI-compat. LLMConfig.APIKey is the encrypted GH token.
	RegisterLLM("copilot", func(cfg LLMConfig) (AIProvider, error) {
		return &CopilotProvider{EncryptedGHToken: cfg.APIKey}, nil
	})
}
