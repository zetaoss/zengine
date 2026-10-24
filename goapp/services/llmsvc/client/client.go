package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/app/config"
)

type LLMClient struct {
	Endpoint string
	Client   *http.Client
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type ChatResponse struct {
	Content string
	Model   string
	Usage   Usage
}

func New(cfg *config.Config) *LLMClient {
	endpoint := ""
	if cfg != nil && cfg.API.BobEndpoint != "" {
		endpoint = cfg.API.BobEndpoint + "/aigate"
	}
	return NewClientWithEndpoint(endpoint)
}

func NewClientWithEndpoint(endpoint string) *LLMClient {
	return &LLMClient{Endpoint: normalizeEndpoint(endpoint), Client: &http.Client{Timeout: 60 * time.Second}}
}

func (s *LLMClient) ChatCompletion(ctx context.Context, model string, messages []Message) (string, string, error) {
	resp, err := s.ChatCompletionFull(ctx, model, messages)
	if err != nil {
		return "", "", err
	}
	return resp.Content, resp.Model, nil
}

func (s *LLMClient) ChatCompletionFull(ctx context.Context, model string, messages []Message) (ChatResponse, error) {
	if strings.TrimSpace(s.Endpoint) == "" {
		return ChatResponse{}, fmt.Errorf("BOB_ENDPOINT is required")
	}
	if s.Client == nil {
		s.Client = &http.Client{Timeout: 60 * time.Second}
	}
	url := s.Endpoint + "/v1/chat/completions"
	body := app.H{
		"messages": messages,
	}
	if model != "" {
		body["model"] = model
	}
	b, _ := json.Marshal(body)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(string(b)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.Client.Do(req)
	if err != nil {
		return ChatResponse{}, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	var payload struct {
		Model   string `json:"model"`
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage Usage `json:"usage"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return ChatResponse{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ChatResponse{}, fmt.Errorf("llm request failed: status=%d", resp.StatusCode)
	}
	if len(payload.Choices) == 0 || strings.TrimSpace(payload.Choices[0].Message.Content) == "" {
		return ChatResponse{}, fmt.Errorf("empty llm output")
	}
	return ChatResponse{
		Content: payload.Choices[0].Message.Content,
		Model:   payload.Model,
		Usage:   payload.Usage,
	}, nil
}

// GetJSON fetches an aigate GET path such as "/v1/models?all" or "/v1/usage" and returns its body and status.
func (s *LLMClient) GetJSON(ctx context.Context, path string) (json.RawMessage, int, error) {
	if strings.TrimSpace(s.Endpoint) == "" {
		return nil, 0, fmt.Errorf("BOB_ENDPOINT is required")
	}
	if s.Client == nil {
		s.Client = &http.Client{Timeout: 60 * time.Second}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.Endpoint+path, nil)
	if err != nil {
		return nil, 0, err
	}
	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	var body json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, resp.StatusCode, err
	}
	return body, resp.StatusCode, nil
}

func normalizeEndpoint(raw string) string {
	v := strings.TrimSpace(raw)
	if v == "" {
		return ""
	}
	if !strings.Contains(v, "://") {
		v = "http://" + v
	}
	u, err := url.Parse(v)
	if err != nil {
		return strings.TrimRight(v, "/")
	}
	u.Path = strings.TrimRight(u.Path, "/")
	return strings.TrimRight(u.String(), "/")
}
