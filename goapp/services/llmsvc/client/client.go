package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/app/config"
)

type LLMClient struct {
	Endpoint string
	Client   *http.Client
	// wait pauses before retrying a rate-limited request; tests replace it.
	wait func(context.Context, time.Duration) error
}

// A rate-limited request (429) is retried once after the Retry-After delay, or defaultRateLimitWait
// when none is given. A longer delay, such as a daily quota, fails at once instead of holding the task.
const (
	defaultRateLimitWait = 5 * time.Second
	maxRateLimitWait     = 30 * time.Second
)

// StatusError is a non-2xx answer from aigate. RetryAfter is its Retry-After delay, zero when absent.
type StatusError struct {
	StatusCode int
	RetryAfter time.Duration
	Message    string
}

func (e *StatusError) Error() string {
	if e.Message == "" {
		return fmt.Sprintf("llm request failed: status=%d", e.StatusCode)
	}
	return fmt.Sprintf("llm request failed: status=%d: %s", e.StatusCode, e.Message)
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
	resp, err := s.chatCompletionOnce(ctx, model, messages)
	var statusErr *StatusError
	if !errors.As(err, &statusErr) || statusErr.StatusCode != http.StatusTooManyRequests {
		return resp, err
	}
	delay := statusErr.RetryAfter
	if delay <= 0 {
		delay = defaultRateLimitWait
	}
	if delay > maxRateLimitWait {
		return resp, err
	}
	wait := s.wait
	if wait == nil {
		wait = sleepContext
	}
	if waitErr := wait(ctx, delay); waitErr != nil {
		return resp, err
	}
	return s.chatCompletionOnce(ctx, model, messages)
}

func (s *LLMClient) chatCompletionOnce(ctx context.Context, model string, messages []Message) (ChatResponse, error) {
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
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return ChatResponse{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var e struct {
			Error string `json:"error"`
		}
		_ = json.Unmarshal(raw, &e)
		secs, _ := strconv.Atoi(resp.Header.Get("Retry-After"))
		return ChatResponse{}, &StatusError{StatusCode: resp.StatusCode, RetryAfter: time.Duration(secs) * time.Second, Message: truncate(e.Error, maxErrorMessage)}
	}
	var payload struct {
		Model   string `json:"model"`
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage Usage `json:"usage"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return ChatResponse{}, err
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

// maxErrorMessage caps aigate's error text, which ends up in run records.
const maxErrorMessage = 300

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}

func sleepContext(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
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
