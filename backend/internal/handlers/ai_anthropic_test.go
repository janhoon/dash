package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
)

const anthropicStreamFixture = "" +
	"event: message_start\n" +
	"data: {\"type\":\"message_start\",\"message\":{\"id\":\"msg_01\",\"role\":\"assistant\",\"content\":[]}}\n" +
	"\n" +
	"event: content_block_start\n" +
	"data: {\"type\":\"content_block_start\",\"index\":0,\"content_block\":{\"type\":\"text\",\"text\":\"\"}}\n" +
	"\n" +
	"event: ping\n" +
	"data: {\"type\":\"ping\"}\n" +
	"\n" +
	"event: content_block_delta\n" +
	"data: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\"Hello\"}}\n" +
	"\n" +
	"event: content_block_delta\n" +
	"data: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"thinking_delta\",\"thinking\":\"scratch\"}}\n" +
	"\n" +
	"event: content_block_delta\n" +
	"data: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\" world\"}}\n" +
	"\n" +
	"event: content_block_stop\n" +
	"data: {\"type\":\"content_block_stop\",\"index\":0}\n" +
	"\n" +
	"event: message_delta\n" +
	"data: {\"type\":\"message_delta\",\"delta\":{\"stop_reason\":\"end_turn\"}}\n" +
	"\n" +
	"event: message_stop\n" +
	"data: {\"type\":\"message_stop\"}\n" +
	"\n"

func TestAnthropicListModels_StaticAllowList(t *testing.T) {
	p, err := NewAnthropic(LLMConfig{DisplayName: "Anthropic"})
	if err != nil {
		t.Fatalf("NewAnthropic: %v", err)
	}
	models, err := p.ListModels(context.Background())
	if err != nil {
		t.Fatalf("ListModels: %v", err)
	}
	if len(models) == 0 {
		t.Fatal("expected static allow-list, got none")
	}
	seen := map[string]bool{}
	for _, m := range models {
		if m.ID == "" {
			t.Fatalf("model missing id: %+v", m)
		}
		if m.Vendor != "anthropic" {
			t.Errorf("expected vendor anthropic for %s, got %q", m.ID, m.Vendor)
		}
		if seen[m.ID] {
			t.Errorf("duplicate model id %q", m.ID)
		}
		seen[m.ID] = true
	}
	for _, id := range []string{"claude-opus-5", "claude-sonnet-5", "claude-haiku-4-5"} {
		if !seen[id] {
			t.Errorf("allow-list missing %s", id)
		}
	}
}

func TestAnthropicChat_StreamMapsToOpenAISSE(t *testing.T) {
	var gotPath, gotAPIKey, gotVersion, gotContentType string
	var gotBody map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAPIKey = r.Header.Get("x-api-key")
		gotVersion = r.Header.Get("anthropic-version")
		gotContentType = r.Header.Get("Content-Type")
		raw, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(raw, &gotBody); err != nil {
			t.Errorf("upstream body: %v", err)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(anthropicStreamFixture))
	}))
	defer server.Close()

	p, err := NewAnthropic(LLMConfig{
		BaseURL: server.URL,
		APIKey:  "sk-ant-test",
	})
	if err != nil {
		t.Fatalf("NewAnthropic: %v", err)
	}

	rr := httptest.NewRecorder()
	err = p.Chat(context.Background(), ChatRequest{
		Model: "claude-sonnet-5",
		Messages: []json.RawMessage{
			json.RawMessage(`{"role":"system","content":"You are Ace."}`),
			json.RawMessage(`{"role":"user","content":"hi"}`),
		},
		Stream: true,
	}, rr)
	if err != nil {
		t.Fatalf("Chat: %v", err)
	}

	if gotPath != "/v1/messages" {
		t.Errorf("expected POST /v1/messages, got %s", gotPath)
	}
	if gotAPIKey != "sk-ant-test" {
		t.Errorf("expected x-api-key, got %q", gotAPIKey)
	}
	if gotVersion != "2023-06-01" {
		t.Errorf("expected anthropic-version 2023-06-01, got %q", gotVersion)
	}
	if gotContentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", gotContentType)
	}
	if gotBody["model"] != "claude-sonnet-5" {
		t.Errorf("expected model claude-sonnet-5, got %#v", gotBody["model"])
	}
	if gotBody["stream"] != true {
		t.Errorf("expected stream true, got %#v", gotBody["stream"])
	}
	if gotBody["system"] != "You are Ace." {
		t.Errorf("expected system extracted, got %#v", gotBody["system"])
	}
	if _, ok := gotBody["tools"]; ok {
		t.Errorf("tools must not be forwarded, body=%v", gotBody)
	}

	if ct := rr.Header().Get("Content-Type"); ct != "text/event-stream" {
		t.Errorf("expected Content-Type text/event-stream, got %q", ct)
	}

	out := rr.Body.String()
	if strings.Contains(out, "content_block_delta") {
		t.Errorf("must not leak Anthropic event JSON, got %s", out)
	}
	if strings.Contains(out, "thinking") || strings.Contains(out, "scratch") {
		t.Errorf("must ignore thinking deltas, got %s", out)
	}
	if !strings.Contains(out, `data: {"choices":[{"index":0,"delta":{"content":"Hello"}}]}`) {
		t.Errorf("missing Hello OpenAI SSE delta, got %s", out)
	}
	if !strings.Contains(out, `data: {"choices":[{"index":0,"delta":{"content":" world"}}]}`) {
		t.Errorf("missing world OpenAI SSE delta, got %s", out)
	}
	if !strings.Contains(out, "data: [DONE]") {
		t.Errorf("missing OpenAI [DONE], got %s", out)
	}

	if got := collectOpenAISSEContent(out); got != "Hello world" {
		t.Errorf("concatenated deltas: got %q want %q", got, "Hello world")
	}
}

func TestAnthropicChat_NonStreamMapsToOpenAIJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"msg_01","type":"message","role":"assistant","content":[{"type":"text","text":"Hello world"}],"stop_reason":"end_turn"}`))
	}))
	defer server.Close()

	p, err := NewAnthropic(LLMConfig{BaseURL: server.URL, APIKey: "sk-ant-test"})
	if err != nil {
		t.Fatalf("NewAnthropic: %v", err)
	}

	rr := httptest.NewRecorder()
	err = p.Chat(context.Background(), ChatRequest{
		Model:    "claude-sonnet-5",
		Messages: []json.RawMessage{json.RawMessage(`{"role":"user","content":"hi"}`)},
	}, rr)
	if err != nil {
		t.Fatalf("Chat: %v", err)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json, got %q", ct)
	}
	var payload struct {
		Choices []struct {
			Message struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("openai json: %v body=%s", err, rr.Body.String())
	}
	if len(payload.Choices) != 1 || payload.Choices[0].Message.Content != "Hello world" {
		t.Fatalf("unexpected openai json: %s", rr.Body.String())
	}
	if payload.Choices[0].Message.Role != "assistant" {
		t.Errorf("expected assistant role, got %q", payload.Choices[0].Message.Role)
	}
}

func TestAnthropicChat_MergesSystemAndConsecutiveRoles(t *testing.T) {
	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"content":[{"type":"text","text":"ok"}]}`))
	}))
	defer server.Close()

	p, err := NewAnthropic(LLMConfig{BaseURL: server.URL, APIKey: "k"})
	if err != nil {
		t.Fatalf("NewAnthropic: %v", err)
	}
	err = p.Chat(context.Background(), ChatRequest{
		Model: "claude-sonnet-5",
		Messages: []json.RawMessage{
			json.RawMessage(`{"role":"system","content":"A"}`),
			json.RawMessage(`{"role":"system","content":"B"}`),
			json.RawMessage(`{"role":"user","content":"one"}`),
			json.RawMessage(`{"role":"user","content":"two"}`),
			json.RawMessage(`{"role":"assistant","content":"prev"}`),
			json.RawMessage(`{"role":"tool","content":"ignored"}`),
			json.RawMessage(`{"role":"user","content":"three"}`),
		},
	}, httptest.NewRecorder())
	if err != nil {
		t.Fatalf("Chat: %v", err)
	}
	if gotBody["system"] != "A\n\nB" {
		t.Errorf("system concat: %#v", gotBody["system"])
	}
	msgs, _ := gotBody["messages"].([]any)
	if len(msgs) != 3 {
		t.Fatalf("expected 3 anthropic messages (merged + skip tool), got %#v", gotBody["messages"])
	}
	first, _ := msgs[0].(map[string]any)
	if first["role"] != "user" || first["content"] != "one\n\ntwo" {
		t.Errorf("first message: %#v", first)
	}
}

func TestAnthropicChat_UpstreamError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":{"message":"invalid x-api-key"}}`))
	}))
	defer server.Close()

	p, err := NewAnthropic(LLMConfig{BaseURL: server.URL, APIKey: "bad"})
	if err != nil {
		t.Fatalf("NewAnthropic: %v", err)
	}
	err = p.Chat(context.Background(), ChatRequest{
		Model:    "claude-sonnet-5",
		Messages: []json.RawMessage{json.RawMessage(`{"role":"user","content":"hi"}`)},
		Stream:   true,
	}, httptest.NewRecorder())
	if err == nil {
		t.Fatal("expected upstream error")
	}
	if !strings.Contains(err.Error(), "401") {
		t.Errorf("expected 401 in error, got %v", err)
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

func collectOpenAISSEContent(body string) string {
	var out strings.Builder
	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "data: ") || trimmed == "data: [DONE]" {
			continue
		}
		var payload struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
			} `json:"choices"`
		}
		if err := json.Unmarshal([]byte(trimmed[6:]), &payload); err != nil {
			continue
		}
		if len(payload.Choices) > 0 {
			out.WriteString(payload.Choices[0].Delta.Content)
		}
	}
	return out.String()
}
