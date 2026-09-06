package llm

import (
	"context"
	"encoding/json"
	"net/http"
)

// AIProvider is the LLM module contract. In-tree providers and future
// ace-llm-* modules implement this interface.
type AIProvider interface {
	ListModels(ctx context.Context) ([]AIModel, error)
	Chat(ctx context.Context, req ChatRequest, w http.ResponseWriter) error
}

// AIModel is a model advertised by a provider.
type AIModel struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Vendor   string                 `json:"vendor"`
	Category string                 `json:"category"`
	Meta     map[string]interface{} `json:"meta,omitempty"`
}

// ChatRequest is the in-process chat completion request.
type ChatRequest struct {
	Model    string            `json:"model"`
	Messages []json.RawMessage `json:"messages"`
	Tools    []json.RawMessage `json:"tools,omitempty"`
	Stream   bool              `json:"stream"`
}

// LLMConfig is the in-process config passed to an LLM factory.
type LLMConfig struct {
	BaseURL string
	// APIKey is decrypted plaintext for openai/openrouter/ollama/custom/anthropic.
	// For copilot it is still-encrypted GH token; CopilotProvider decrypts
	// EncryptedGHToken on ListModels/Chat.
	APIKey      string
	DisplayName string
}

// LLMFactory constructs an AIProvider from LLMConfig.
type LLMFactory func(LLMConfig) (AIProvider, error)
