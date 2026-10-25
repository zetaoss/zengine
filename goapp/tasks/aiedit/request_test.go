package aiedit

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/zetaoss/zengine/goapp/app/config"
	"github.com/zetaoss/zengine/goapp/models"
)

// fakeWiki answers the MediaWiki API calls Render makes: the saved page, its backlinks and their wikitext.
func fakeWiki(t *testing.T) *config.Config {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		var out any
		switch {
		case q.Get("list") == "backlinks":
			out = map[string]any{"query": map[string]any{"backlinks": []map[string]any{{"pageid": 7}}}}
		case q.Get("pageids") == "7":
			out = map[string]any{"query": map[string]any{"pages": []map[string]any{{"pageid": 7, "title": "과일", "revisions": []map[string]any{{"slots": map[string]any{"main": map[string]any{"content": "* [[사과]]는 빨갛다."}}}}}}}}
		case q.Get("titles") != "":
			out = map[string]any{"query": map[string]any{"pages": []map[string]any{{"pageid": 1, "title": q.Get("titles"), "revisions": []map[string]any{{"slots": map[string]any{"main": map[string]any{"content": "저장된 본문"}}}}}}}}
		}
		_ = json.NewEncoder(w).Encode(out)
	}))
	t.Cleanup(ts.Close)
	cfg := &config.Config{}
	cfg.App.APIServer = ts.URL
	return cfg
}

func contents(r Rendered) string {
	return MessagesText(r.Messages)
}

func TestRenderCreate(t *testing.T) {
	cfg := fakeWiki(t)
	if _, err := Render(context.Background(), cfg, Request{RequestType: models.AIEditRequestCreate, Title: "사과"}); !IsInputError(err) {
		t.Fatalf("a two-letter title needs a hint, got %v", err)
	}
	r, err := Render(context.Background(), cfg, Request{RequestType: models.AIEditRequestCreate, Title: "사과", UserInputs: map[string]any{
		"hint": "apple", "readBacklinks": true, "materials": []any{"자료 하나", ""}, "request": "",
	}})
	if err != nil {
		t.Fatal(err)
	}
	got := contents(r)
	for _, want := range []string{"### 문서 제목: 사과", "### 힌트: apple", "### 자료 1:\n자료 하나", "### 역링크 1:\n문서: 과일\n* [[사과]]는 빨갛다."} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in\n%s", want, got)
		}
	}
	if strings.Contains(got, "### 요청사항") || strings.Contains(got, "### 자료 2") {
		t.Errorf("an empty request and a blank material must send nothing:\n%s", got)
	}
	if r.Messages[0].Role != "system" || len(r.Messages) != 5 {
		t.Errorf("unexpected messages %+v", r.Messages)
	}
}

func TestRenderUpdateTemplate(t *testing.T) {
	cfg := fakeWiki(t)
	prompt := &models.AIEditPrompt{ID: 3, RequestType: models.AIEditRequestUpdate, Content: "'{title}' 문서를 고쳐라.\n{wikitext}\n{{틀}}"}
	r, err := Render(context.Background(), cfg, Request{RequestType: models.AIEditRequestUpdate, Title: "자바 (섬)", Prompt: prompt, UserInputs: map[string]any{"request": "표로 정리"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Messages) != 2 || r.Messages[0].Content != "'자바 (섬)' 문서를 고쳐라.\n저장된 본문\n{{틀}}" || r.Messages[1].Content != "### 요청사항:\n표로 정리" {
		t.Fatalf("a template replaces the unconditional messages and keeps the run form's: %+v", r.Messages)
	}
	if r.Vars["backlinkTexts"] == nil {
		t.Error("backlinkTexts must be filled (empty) when readBacklinks is off")
	}
	if _, err := Render(context.Background(), cfg, Request{RequestType: models.AIEditRequestUpdate, Title: "자바", UserInputs: map[string]any{"hint": "x"}}); !IsInputError(err) {
		t.Errorf("doc-update has no hint, got %v", err)
	}
}

func TestNeedsHint(t *testing.T) {
	for title, want := range map[string]bool{"사과": true, "Go": true, "배": true, "사 과": true, "자바 (섬)": false, "사과나무": false, "C++": false} {
		if got := NeedsHint(title); got != want {
			t.Errorf("NeedsHint(%q) = %v, want %v", title, got, want)
		}
	}
}
