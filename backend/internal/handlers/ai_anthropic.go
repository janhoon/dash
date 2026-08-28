package handlers

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	anthropicAPIVersion = "2023-06-01"
	anthropicMaxTokens  = 4096
)

// Outbound HTTP uses Go's default http.Client (timeout only). Base URLs are
// checked at save time by validateBaseURL, not by ssrf.SafeClient or
// ssrf.DatasourceClient — see docs/adr/0003-outbound-http-ssrf-policy-seams.md.
type AnthropicProvider struct {
	BaseURL     string
	APIKey      string
	DisplayName string
}

func NewAnthropic(cfg LLMConfig) (AIProvider, error) {
	return &AnthropicProvider{
		BaseURL:     strings.TrimRight(cfg.BaseURL, "/"),
		APIKey:      cfg.APIKey,
		DisplayName: cfg.DisplayName,
	}, nil
}

var anthropicAllowList = []AIModel{
	{ID: "claude-opus-5", Name: "Claude Opus 5", Vendor: "anthropic"},
	{ID: "claude-sonnet-5", Name: "Claude Sonnet 5", Vendor: "anthropic"},
	{ID: "claude-haiku-4-5", Name: "Claude Haiku 4.5", Vendor: "anthropic"},
	{ID: "claude-fable-5", Name: "Claude Fable 5", Vendor: "anthropic"},
}

func (p *AnthropicProvider) ListModels(ctx context.Context) ([]AIModel, error) {
	out := make([]AIModel, len(anthropicAllowList))
	copy(out, anthropicAllowList)
	return out, nil
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicStreamEvent struct {
	Type  string `json:"type"`
	Delta struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"delta"`
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

type openAIStreamChunk struct {
	Choices []openAIStreamChoice `json:"choices"`
}

type openAIStreamChoice struct {
	Index int         `json:"index"`
	Delta openAIDelta `json:"delta"`
}

type openAIDelta struct {
	Content string `json:"content"`
}

// Chat sends Messages API requests and writes OpenAI-shaped JSON or SSE to w.
func (p *AnthropicProvider) Chat(ctx context.Context, chatReq ChatRequest, w http.ResponseWriter) error {
	if len(chatReq.Tools) > 0 {
		return fmt.Errorf("provider returned 400: tools not supported")
	}

	system, messages, err := openaiToAnthropicMessages(chatReq.Messages)
	if err != nil {
		return err
	}
	if len(messages) == 0 {
		return fmt.Errorf("anthropic chat requires at least one user or assistant message")
	}
	if messages[0].Role != "user" {
		return fmt.Errorf("anthropic chat requires the first message to be from the user")
	}

	body := map[string]any{
		"model":      chatReq.Model,
		"max_tokens": anthropicMaxTokens,
		"stream":     chatReq.Stream,
		"messages":   messages,
	}
	if system != "" {
		body["system"] = system
	}
	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.BaseURL+"/v1/messages", bytes.NewReader(bodyJSON))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.APIKey)
	req.Header.Set("anthropic-version", anthropicAPIVersion)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to reach provider: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("provider returned %d: %s", resp.StatusCode, string(respBody))
	}

	if chatReq.Stream {
		return writeAnthropicStreamAsOpenAI(resp.Body, w)
	}
	return writeAnthropicJSONAsOpenAI(resp.Body, w)
}

func openaiToAnthropicMessages(raw []json.RawMessage) (system string, messages []anthropicMessage, err error) {
	var systemParts []string
	for _, item := range raw {
		role, text, convErr := openaiMessageText(item)
		if convErr != nil {
			return "", nil, convErr
		}
		switch role {
		case "system":
			if text != "" {
				systemParts = append(systemParts, text)
			}
		case "user", "assistant":
			if text == "" {
				continue
			}
			if len(messages) > 0 && messages[len(messages)-1].Role == role {
				messages[len(messages)-1].Content += "\n\n" + text
				continue
			}
			messages = append(messages, anthropicMessage{Role: role, Content: text})
		default:
			continue
		}
	}
	return strings.Join(systemParts, "\n\n"), messages, nil
}

func openaiMessageText(raw json.RawMessage) (role, text string, err error) {
	var m struct {
		Role    string          `json:"role"`
		Content json.RawMessage `json:"content"`
	}
	if err := json.Unmarshal(raw, &m); err != nil {
		return "", "", fmt.Errorf("invalid chat message: %w", err)
	}
	if len(m.Content) == 0 || string(m.Content) == "null" {
		return m.Role, "", nil
	}
	var asString string
	if err := json.Unmarshal(m.Content, &asString); err == nil {
		return m.Role, asString, nil
	}
	var parts []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if err := json.Unmarshal(m.Content, &parts); err == nil {
		var b strings.Builder
		for _, part := range parts {
			if part.Type == "text" || part.Type == "" {
				b.WriteString(part.Text)
			}
		}
		return m.Role, b.String(), nil
	}
	return m.Role, "", nil
}

func writeAnthropicStreamAsOpenAI(r io.Reader, w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	flusher, canFlush := w.(http.Flusher)

	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	wroteDone := false
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if payload == "" {
			continue
		}
		var ev anthropicStreamEvent
		if err := json.Unmarshal([]byte(payload), &ev); err != nil {
			continue
		}
		switch ev.Type {
		case "content_block_delta":
			if ev.Delta.Type != "text_delta" || ev.Delta.Text == "" {
				continue
			}
			if err := writeOpenAIDelta(w, ev.Delta.Text); err != nil {
				return err
			}
			if canFlush {
				flusher.Flush()
			}
		case "message_stop":
			if _, err := io.WriteString(w, "data: [DONE]\n\n"); err != nil {
				return err
			}
			wroteDone = true
			if canFlush {
				flusher.Flush()
			}
		case "error":
			msg := ev.Error.Message
			if msg == "" {
				msg = "stream error"
			}
			return fmt.Errorf("anthropic stream error: %s", msg)
		default:
			continue
		}
	}
	if err := sc.Err(); err != nil {
		return err
	}
	if !wroteDone {
		if _, err := io.WriteString(w, "data: [DONE]\n\n"); err != nil {
			return err
		}
		if canFlush {
			flusher.Flush()
		}
	}
	return nil
}

func writeOpenAIDelta(w http.ResponseWriter, text string) error {
	chunk := openAIStreamChunk{
		Choices: []openAIStreamChoice{{
			Index: 0,
			Delta: openAIDelta{Content: text},
		}},
	}
	payload, err := json.Marshal(chunk)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "data: %s\n\n", payload)
	return err
}

func writeAnthropicJSONAsOpenAI(r io.Reader, w http.ResponseWriter) error {
	var msg struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.NewDecoder(r).Decode(&msg); err != nil {
		return fmt.Errorf("failed to decode anthropic response: %w", err)
	}
	var text strings.Builder
	for _, block := range msg.Content {
		if block.Type == "text" {
			text.WriteString(block.Text)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(map[string]any{
		"choices": []map[string]any{
			{
				"index": 0,
				"message": map[string]any{
					"role":    "assistant",
					"content": text.String(),
				},
				"finish_reason": "stop",
			},
		},
	})
}
