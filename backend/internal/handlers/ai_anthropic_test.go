package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestChat_Anthropic_ToolsTriggerHostRetry(t *testing.T) {
	hits := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		raw, _ := io.ReadAll(r.Body)
		var body map[string]any
		if err := json.Unmarshal(raw, &body); err != nil {
			t.Errorf("upstream body: %v", err)
		}
		if _, ok := body["tools"]; ok {
			t.Errorf("retry must not forward tools, body=%v", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"content":[{"type":"text","text":"retried-without-tools"}]}`))
	}))
	defer server.Close()

	h := NewAIHandler(nil, nil)
	rowID := uuid.New()
	orgID := uuid.New()
	h.testProviderRow = &providerRow{
		ID:           rowID,
		ProviderType: "anthropic",
		DisplayName:  "Anthropic",
		BaseURL:      server.URL,
	}

	body, _ := json.Marshal(map[string]any{
		"provider_id": rowID.String(),
		"model":       "claude-sonnet-5",
		"messages":    []map[string]string{{"role": "user", "content": "hi"}},
		"tools": []map[string]any{
			{"type": "function", "function": map[string]string{"name": "query_prometheus"}},
		},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/orgs/"+orgID.String()+"/ai/chat", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(ctxWithUserAndOrg(uuid.New(), orgID))

	rr := httptest.NewRecorder()
	h.Chat(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 after tool retry, got %d: %s", rr.Code, rr.Body.String())
	}
	if rr.Header().Get("X-Tools-Unsupported") != "true" {
		t.Error("expected X-Tools-Unsupported after Anthropic tool rejection")
	}
	if hits != 1 {
		t.Fatalf("expected one upstream call (retry only), got %d", hits)
	}
	if !strings.Contains(rr.Body.String(), "retried-without-tools") {
		t.Errorf("expected retried content, got %s", rr.Body.String())
	}
}

func TestChat_AnthropicProviderType_Dispatches(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"content":[{"type":"text","text":"from-anthropic"}]}`))
	}))
	defer server.Close()

	h := NewAIHandler(nil, nil)
	rowID := uuid.New()
	orgID := uuid.New()
	h.testProviderRow = &providerRow{
		ID:           rowID,
		ProviderType: "anthropic",
		DisplayName:  "Anthropic",
		BaseURL:      server.URL,
	}

	body, _ := json.Marshal(map[string]any{
		"provider_id": rowID.String(),
		"model":       "claude-sonnet-5",
		"messages":    []map[string]string{{"role": "user", "content": "hi"}},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/orgs/"+orgID.String()+"/ai/chat", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(ctxWithUserAndOrg(uuid.New(), orgID))

	rr := httptest.NewRecorder()
	h.Chat(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 for registered anthropic, got %d: %s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "from-anthropic") {
		t.Errorf("expected mapped content, got %s", rr.Body.String())
	}
}
