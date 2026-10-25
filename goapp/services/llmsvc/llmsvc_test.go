package llmsvc

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/zetaoss/zengine/goapp/prompts"
)

// newServer answers the n-th chat completion with contents[n] (the last one repeats).
func newServer(t *testing.T, contents ...string) (*httptest.Server, *atomic.Int32) {
	t.Helper()
	var calls atomic.Int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := int(calls.Add(1)) - 1
		content := contents[min(n, len(contents)-1)]
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"model":   "test/model",
			"choices": []map[string]any{{"message": map[string]any{"content": content}}},
			"usage":   map[string]any{"prompt_tokens": 10, "completion_tokens": 5, "total_tokens": 15},
		})
	}))
	t.Cleanup(ts.Close)
	return ts, &calls
}

var testChain = prompts.Chain{Name: "test-chain", Metadata: prompts.Metadata{Version: "v9"}}

var testStep = prompts.Step{
	Name:   "test",
	Prompt: prompts.Prompt{Inputs: []string{"title"}, Messages: []prompts.Message{{Role: "user", Content: "{title}"}}},
	Parser: prompts.ParserJSON,
	Output: "result",
}

type testOutput struct {
	Violated bool   `json:"violated"`
	Count    int    `json:"count"`
	Summary  string `json:"summary"`
}

func TestRunStepRetryDropsFailedAttempt(t *testing.T) {
	// The first response has a type error after violated, which encoding/json still decodes; the retry
	// omits violated, so it must stay false rather than carry over from the failed attempt.
	ts, calls := newServer(t, `{"violated": true, "count": "two"}`, `{"count": 2, "summary": "ok"}`)
	var out testOutput
	res, err := NewWithEndpoint(ts.URL).RunStep(context.Background(), "test/model", testChain, testStep, map[string]string{"title": "T"}, &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls.Load() != 2 || res.Usage.PromptTokens != 20 || res.Usage.CompletionTokens != 10 {
		t.Fatalf("expected 2 calls with summed usage, got %d calls, %+v", calls.Load(), res.Usage)
	}
	if out != (testOutput{Count: 2, Summary: "ok"}) {
		t.Fatalf("failed attempt leaked into output: %+v", out)
	}
}

func TestRunStepFailureTruncatesRaw(t *testing.T) {
	ts, calls := newServer(t, "not json "+strings.Repeat("가", 2000))
	out := testOutput{Summary: "before"}
	_, err := NewWithEndpoint(ts.URL).RunStep(context.Background(), "test/model", testChain, testStep, nil, &out)
	if err == nil || calls.Load() != 2 {
		t.Fatalf("expected an error after 2 calls, got %v after %d", err, calls.Load())
	}
	if n := len([]rune(err.Error())); n > maxRawInError+200 {
		t.Fatalf("error not truncated: %d runes", n)
	}
	if out.Summary != "before" {
		t.Fatalf("output changed on failure: %+v", out)
	}
}

func TestRunStepRejectsNonPointer(t *testing.T) {
	var out testOutput
	if _, err := NewWithEndpoint("http://127.0.0.1:0").RunStep(context.Background(), "m", testChain, testStep, nil, out); err == nil {
		t.Fatal("non-pointer out accepted")
	}
}

func TestRunStepKeepsParseErrorWhenRetryFails(t *testing.T) {
	var calls atomic.Int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			_ = json.NewEncoder(w).Encode(map[string]any{"choices": []map[string]any{{"message": map[string]any{"content": "not json"}}}})
			return
		}
		w.Header().Set("Retry-After", "3600")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":"rate limited"}`))
	}))
	t.Cleanup(ts.Close)

	var out testOutput
	_, err := NewWithEndpoint(ts.URL).RunStep(context.Background(), "", testChain, testStep, map[string]string{"title": "x"}, &out)
	if err == nil || !strings.Contains(err.Error(), "status=429") || !strings.Contains(err.Error(), "previous attempt: invalid") || !strings.Contains(err.Error(), "not json") {
		t.Fatalf("expected the 429 with the earlier parse error, got %v", err)
	}
}

func TestRunStepRecordsEachCall(t *testing.T) {
	ts, _ := newServer(t, `not json`, `{"count": 1}`)
	var got []Call
	svc := NewWithEndpoint(ts.URL).WithRecorder(func(_ context.Context, c Call) { got = append(got, c) })

	var out testOutput
	if _, err := svc.RunStep(context.Background(), "test/model", testChain, testStep, map[string]string{"title": "T"}, &out); err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("expected both calls recorded, got %+v", got)
	}
	first, second := got[0], got[1]
	if first.OK || first.Err == nil || !strings.Contains(first.Err.Error(), "invalid json output") {
		t.Fatalf("expected the unparsable answer recorded as a failure, got %+v", first)
	}
	if !second.OK || second.Err != nil || second.Chain != "test-chain" || second.ChainVersion != "v9" || second.Step != "test" ||
		second.Model != "test/model" || second.ResolvedModel != "test/model" || second.Usage.PromptTokens != 10 {
		t.Fatalf("unexpected recorded success: %+v", second)
	}
}
