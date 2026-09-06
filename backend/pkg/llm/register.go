// Package llm is the stable import path for Ace LLM module contracts.
//
// External ace-llm-* modules import github.com/aceobservability/ace/backend/pkg/llm
// and call RegisterLLM from init. In-tree providers in internal/handlers register
// the same way. Do not import internal/handlers from a module.
package llm

import (
	"errors"
	"fmt"
	"sync"
)

// ErrUnknownProviderType is returned when provider_type has no registered
// factory. Callers must fail closed: never fall through to OpenAI-compat.
var ErrUnknownProviderType = errors.New("unknown provider_type")

var pluginRegistry = struct {
	mu sync.RWMutex
	m  map[string]LLMFactory
}{
	m: map[string]LLMFactory{},
}

// RegisterLLM records a factory for provider_type.
func RegisterLLM(providerType string, factory LLMFactory) {
	if providerType == "" {
		panic("RegisterLLM: empty provider_type")
	}
	if factory == nil {
		panic("RegisterLLM: nil factory")
	}
	pluginRegistry.mu.Lock()
	defer pluginRegistry.mu.Unlock()
	if _, exists := pluginRegistry.m[providerType]; exists {
		panic(fmt.Sprintf("RegisterLLM: provider_type already registered: %s", providerType))
	}
	pluginRegistry.m[providerType] = factory
}

func lookup(providerType string) (LLMFactory, bool) {
	pluginRegistry.mu.RLock()
	defer pluginRegistry.mu.RUnlock()
	f, ok := pluginRegistry.m[providerType]
	return f, ok
}

// UnregisterLLM removes a factory. Tests use this to isolate probe types.
func UnregisterLLM(providerType string) {
	pluginRegistry.mu.Lock()
	defer pluginRegistry.mu.Unlock()
	delete(pluginRegistry.m, providerType)
}

// New constructs a provider from the registry. Unknown types fail closed.
func New(providerType string, cfg LLMConfig) (AIProvider, error) {
	factory, ok := lookup(providerType)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrUnknownProviderType, providerType)
	}
	return factory(cfg)
}

// RequireKnown reports ErrUnknownProviderType when provider_type is unregistered.
func RequireKnown(providerType string) error {
	if _, ok := lookup(providerType); !ok {
		return fmt.Errorf("%w: %s", ErrUnknownProviderType, providerType)
	}
	return nil
}
