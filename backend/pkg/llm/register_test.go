package llm

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
)

type stubProvider struct {
	id string
}

func (s *stubProvider) ListModels(context.Context) ([]AIModel, error) {
	return []AIModel{{ID: s.id, Name: s.id}}, nil
}

func (s *stubProvider) Chat(context.Context, ChatRequest, http.ResponseWriter) error {
	return nil
}

func TestNew_UnknownTypeFailsClosed(t *testing.T) {
	p, err := New("nope", LLMConfig{APIKey: "sk-should-not-be-used"})
	if err == nil {
		t.Fatal("expected error for unknown provider_type")
	}
	if !errors.Is(err, ErrUnknownProviderType) {
		t.Fatalf("expected ErrUnknownProviderType, got %v", err)
	}
	if p != nil {
		t.Fatalf("unknown type must not construct a provider, got %T", p)
	}
}

func TestRegisterLLM_DispatchUsesRegisteredFactory(t *testing.T) {
	const providerType = "probe-pkg-llm"
	RegisterLLM(providerType, func(cfg LLMConfig) (AIProvider, error) {
		return &stubProvider{id: cfg.DisplayName}, nil
	})
	t.Cleanup(func() { UnregisterLLM(providerType) })

	p, err := New(providerType, LLMConfig{DisplayName: "probe-one"})
	if err != nil {
		t.Fatalf("dispatch registered type: %v", err)
	}
	got, ok := p.(*stubProvider)
	if !ok {
		t.Fatalf("expected *stubProvider, got %T", p)
	}
	if got.id != "probe-one" {
		t.Errorf("factory did not receive LLMConfig, id=%q", got.id)
	}

	models, err := p.ListModels(context.Background())
	if err != nil {
		t.Fatalf("ListModels: %v", err)
	}
	if len(models) != 1 || models[0].ID != "probe-one" {
		t.Fatalf("unexpected models: %+v", models)
	}

	if err := p.Chat(context.Background(), ChatRequest{
		Model:    "probe",
		Messages: []json.RawMessage{json.RawMessage(`{"role":"user"}`)},
	}, nil); err != nil {
		t.Fatalf("Chat: %v", err)
	}
}

func TestRequireKnown(t *testing.T) {
	if err := RequireKnown("nope"); !errors.Is(err, ErrUnknownProviderType) {
		t.Fatalf("expected ErrUnknownProviderType, got %v", err)
	}

	const providerType = "probe-require-known"
	RegisterLLM(providerType, func(LLMConfig) (AIProvider, error) {
		return &stubProvider{id: providerType}, nil
	})
	t.Cleanup(func() { UnregisterLLM(providerType) })

	if err := RequireKnown(providerType); err != nil {
		t.Fatalf("registered type should be known: %v", err)
	}
}
