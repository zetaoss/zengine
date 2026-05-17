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

func New(cfg *config.Config) *LLMClient {
	endpoint := ""
	if cfg != nil {
		endpoint = cfg.API.LLMEndpoint
	}
	return NewClientWithEndpoint(endpoint)
}

func NewClientWithEndpoint(endpoint string) *LLMClient {
	return &LLMClient{Endpoint: normalizeEndpoint(endpoint), Client: &http.Client{Timeout: 60 * time.Second}}
}

func (s *LLMClient) ChatCompletion(ctx context.Context, model string, messages []Message) (string, string, error) {
	if strings.TrimSpace(s.Endpoint) == "" {
		return "", "", fmt.Errorf("LLM_ENDPOINT is required")
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
		return "", "", err
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
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", "", fmt.Errorf("llm request failed: status=%d", resp.StatusCode)
	}
	if len(payload.Choices) == 0 || strings.TrimSpace(payload.Choices[0].Message.Content) == "" {
		return "", "", fmt.Errorf("empty llm output")
	}
	return payload.Choices[0].Message.Content, payload.Model, nil
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
