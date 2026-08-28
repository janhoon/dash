package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/aceobservability/ace/backend/internal/crypto"
)

type probeProvider struct {
	id string
	mockAIProvider
}

func TestNewLLMProvider_UnknownTypeFailsClosed(t *testing.T) {
	p, err := newLLMProvider("nope", LLMConfig{
		BaseURL:     "https://api.openai.com/v1",
		APIKey:      "sk-should-not-be-used",
		DisplayName: "Nope",
	})
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

func TestNewLLMProvider_AnthropicNotRegistered(t *testing.T) {
	// Anthropic is #416. Until that lands, a stored type must 400, not impersonate OpenAI.
	p, err := newLLMProvider("anthropic", LLMConfig{
		BaseURL:     "https://api.anthropic.com",
		APIKey:      "sk-ant",
		DisplayName: "Anthropic",
	})
	if !errors.Is(err, ErrUnknownProviderType) {
		t.Fatalf("expected ErrUnknownProviderType for anthropic, got %v", err)
	}
	if p != nil {
		t.Fatalf("anthropic must not construct a provider yet, got %T", p)
	}
}

func TestNewLLMProvider_OpenAICompatListsAndChats(t *testing.T) {
	modelsJSON := `{"data":[{"id":"gpt-4o","owned_by":"openai"}]}`
	chatJSON := `{"id":"chatcmpl-1","choices":[{"index":0,"message":{"role":"assistant","content":"hi"},"finish_reason":"stop"}]}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/models":
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(modelsJSON))
		case r.Method == http.MethodPost && r.URL.Path == "/chat/completions":
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(chatJSON))
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	for _, providerType := range []string{"openai", "openrouter", "ollama", "custom"} {
		t.Run(providerType, func(t *testing.T) {
			p, err := newLLMProvider(providerType, LLMConfig{
				BaseURL:     server.URL,
				APIKey:      "test-api-key",
				DisplayName: providerType,
			})
			if err != nil {
				t.Fatalf("newLLMProvider(%q): %v", providerType, err)
			}
			if _, ok := p.(*OpenAICompatibleProvider); !ok {
				t.Fatalf("expected *OpenAICompatibleProvider, got %T", p)
			}

			models, err := p.ListModels(context.Background())
			if err != nil {
				t.Fatalf("ListModels: %v", err)
			}
			if len(models) != 1 || models[0].ID != "gpt-4o" {
				t.Fatalf("unexpected models: %+v", models)
			}

			rr := httptest.NewRecorder()
			err = p.Chat(context.Background(), ChatRequest{
				Model:    "gpt-4o",
				Messages: []json.RawMessage{json.RawMessage(`{"role":"user","content":"hi"}`)},
			}, rr)
			if err != nil {
				t.Fatalf("Chat: %v", err)
			}
			if !strings.Contains(rr.Body.String(), `"content":"hi"`) {
				t.Fatalf("chat body missing content, got %s", rr.Body.String())
			}
		})
	}
}

func TestNewLLMProvider_CopilotRegistered(t *testing.T) {
	p, err := newLLMProvider("copilot", LLMConfig{APIKey: "encrypted-gh-token"})
	if err != nil {
		t.Fatalf("copilot should be registered: %v", err)
	}
	cp, ok := p.(*CopilotProvider)
	if !ok {
		t.Fatalf("expected *CopilotProvider, got %T", p)
	}
	if cp.EncryptedGHToken != "encrypted-gh-token" {
		t.Errorf("expected EncryptedGHToken from LLMConfig.APIKey, got %q", cp.EncryptedGHToken)
	}
}

func TestInstantiateDBProvider_CopilotKeepsCiphertext(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret-for-plugin-tests")
	// DecryptToken rejects this blob (invalid encoding). If build/instantiate
	// decrypted before the factory, this would fail instead of constructing.
	ciphertext := "not-valid-aes-gcm-ciphertext"
	if _, err := crypto.DecryptToken(ciphertext); err == nil {
		t.Fatal("sanity: blob must fail DecryptToken so a double-decrypt would be visible")
	}

	p, _, err := instantiateDBProvider(providerRow{
		ID:           uuid.New(),
		ProviderType: "copilot",
		DisplayName:  "Copilot",
		APIKey:       &ciphertext,
	})
	if err != nil {
		t.Fatalf("copilot DB path must not decrypt API key: %v", err)
	}
	cp, ok := p.(*CopilotProvider)
	if !ok {
		t.Fatalf("expected *CopilotProvider, got %T", p)
	}
	if cp.EncryptedGHToken != ciphertext {
		t.Errorf("expected ciphertext EncryptedGHToken, got %q", cp.EncryptedGHToken)
	}
}

func TestInstantiateDBProvider_CopilotEncryptedKeyListsAndChats(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret-for-copilot-tests")

	copilotToken := "tid=uuid-copilot-row-token"
	expiresAt := time.Now().Unix() + 3600
	chatJSON := `{"id":"chatcmpl-uuid","choices":[{"index":0,"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}]}`

	copilotAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/models":
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(copilotModelsPayload()))
		case r.Method == http.MethodPost && r.URL.Path == "/chat/completions":
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(chatJSON))
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	defer copilotAPI.Close()

	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"token":      copilotToken,
			"expires_at": expiresAt,
			"endpoints":  map[string]string{"api": copilotAPI.URL},
		})
	}))
	defer tokenServer.Close()

	enc := encryptTestToken(t, "ghp_uuid_copilot_row")
	row := providerRow{
		ID:           uuid.New(),
		ProviderType: "copilot",
		DisplayName:  "Copilot",
		APIKey:       &enc,
	}

	copilotTokenCache.Range(func(key, value any) bool {
		copilotTokenCache.Delete(key)
		return true
	})

	p, _, err := instantiateDBProvider(row)
	if err != nil {
		t.Fatalf("instantiateDBProvider: %v", err)
	}
	cp, ok := p.(*CopilotProvider)
	if !ok {
		t.Fatalf("expected *CopilotProvider, got %T", p)
	}
	if cp.EncryptedGHToken != enc {
		t.Fatal("EncryptedGHToken must still be the stored ciphertext")
	}
	cp.tokenEndpoint = tokenServer.URL + "/copilot_internal/v2/token"

	models, err := cp.ListModels(context.Background())
	if err != nil {
		t.Fatalf("ListModels after one decrypt: %v", err)
	}
	if len(models) == 0 {
		t.Fatal("expected copilot models")
	}

	rr := httptest.NewRecorder()
	err = cp.Chat(context.Background(), ChatRequest{
		Model:    "gpt-4o",
		Messages: []json.RawMessage{json.RawMessage(`{"role":"user","content":"hi"}`)},
	}, rr)
	if err != nil {
		t.Fatalf("Chat after one decrypt: %v", err)
	}
	if !strings.Contains(rr.Body.String(), `"content":"ok"`) {
		t.Fatalf("chat body missing content, got %s", rr.Body.String())
	}
}

func TestInstantiateDBProvider_OpenAIDecryptsAPIKey(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret-for-plugin-tests")
	plain := "sk-test-openai-key"
	enc, err := crypto.EncryptToken(plain)
	if err != nil {
		t.Fatalf("EncryptToken: %v", err)
	}

	p, _, err := instantiateDBProvider(providerRow{
		ID:           uuid.New(),
		ProviderType: "openai",
		DisplayName:  "OpenAI",
		BaseURL:      "https://api.openai.com/v1",
		APIKey:       &enc,
	})
	if err != nil {
		t.Fatalf("openai DB path: %v", err)
	}
	op, ok := p.(*OpenAICompatibleProvider)
	if !ok {
		t.Fatalf("expected *OpenAICompatibleProvider, got %T", p)
	}
	if op.APIKey != plain {
		t.Errorf("openai factory must receive plaintext, got %q", op.APIKey)
	}
}

func TestRegisterLLM_DispatchUsesRegisteredFactory(t *testing.T) {
	const providerType = "probe-dispatch"
	RegisterLLM(providerType, func(cfg LLMConfig) (AIProvider, error) {
		return &probeProvider{id: cfg.DisplayName}, nil
	})
	t.Cleanup(func() { unregisterLLM(providerType) })

	p, err := newLLMProvider(providerType, LLMConfig{DisplayName: "probe-one"})
	if err != nil {
		t.Fatalf("dispatch registered type: %v", err)
	}
	got, ok := p.(*probeProvider)
	if !ok {
		t.Fatalf("expected *probeProvider, got %T", p)
	}
	if got.id != "probe-one" {
		t.Errorf("factory did not receive LLMConfig, id=%q", got.id)
	}
}

func TestChat_UnknownProviderType_Returns400(t *testing.T) {
	h := NewAIHandler(nil, nil)
	h.testProviderRow = &providerRow{
		ID:           uuid.New(),
		ProviderType: "nope",
		DisplayName:  "Nope",
		BaseURL:      "https://example.com/v1",
	}

	userID := uuid.New()
	orgID := uuid.New()
	body, _ := json.Marshal(map[string]any{
		"provider_id": h.testProviderRow.ID.String(),
		"model":       "gpt-4o",
		"messages":    []map[string]string{{"role": "user", "content": "hi"}},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/orgs/"+orgID.String()+"/ai/chat", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(ctxWithUserAndOrg(userID, orgID))

	rr := httptest.NewRecorder()
	h.Chat(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for unknown provider_type, got %d: %s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "unknown provider_type") {
		t.Errorf("expected unknown provider_type in body, got %s", rr.Body.String())
	}
}

func TestChat_RegisteredTypeIsDispatched(t *testing.T) {
	const providerType = "probe-chat-dispatch"
	mock := &mockAIProvider{chatBody: `{"choices":[{"index":0,"delta":{"content":"from-probe"}}]}`}
	RegisterLLM(providerType, func(LLMConfig) (AIProvider, error) {
		return mock, nil
	})
	t.Cleanup(func() { unregisterLLM(providerType) })

	h := NewAIHandler(nil, nil)
	h.testProviderRow = &providerRow{
		ID:           uuid.New(),
		ProviderType: providerType,
		DisplayName:  "Probe",
		BaseURL:      "https://example.com/v1",
	}

	userID := uuid.New()
	orgID := uuid.New()
	body, _ := json.Marshal(map[string]any{
		"provider_id": h.testProviderRow.ID.String(),
		"model":       "probe-model",
		"messages":    []map[string]string{{"role": "user", "content": "hi"}},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/orgs/"+orgID.String()+"/ai/chat", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(ctxWithUserAndOrg(userID, orgID))

	rr := httptest.NewRecorder()
	h.Chat(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	if !mock.chatCalled {
		t.Fatal("registered factory was not used for chat")
	}
	if !strings.Contains(rr.Body.String(), "from-probe") {
		t.Errorf("expected probe chat body, got %s", rr.Body.String())
	}
}

func TestRequireKnownLLMType(t *testing.T) {
	if err := requireKnownLLMType("openai"); err != nil {
		t.Errorf("openai should be known: %v", err)
	}
	if err := requireKnownLLMType("nope"); !errors.Is(err, ErrUnknownProviderType) {
		t.Errorf("nope should fail closed, got %v", err)
	}
}
