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
	"github.com/zetaoss/zengine/goapp/prompts"
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
	r, err := Render(context.Background(), cfg, Request{RequestType: models.AIEditRequestCreate, Title: "사과", UserInputs: map[string]any{
		"hint": "apple", "readBacklinks": true, "copiedTexts": []any{"자료 하나", ""}, "request": "",
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

func TestRenderOwnMessages(t *testing.T) {
	cfg := fakeWiki(t)
	own := `[{"role":"system","content":"규칙"},{"role":"user","content":"{title}"},{"role":"user","if":"request","content":"요청: {request}"},{"role":"user","each":"copiedTexts","content":"{index}) {copiedTexts}"}]`
	prompt := &models.AIEditPrompt{ID: 4, RequestType: models.AIEditRequestCreate, Content: "unused", Messages: own}
	r, err := Render(context.Background(), cfg, Request{RequestType: models.AIEditRequestCreate, Title: "자바 (섬)", Prompt: prompt, UserInputs: map[string]any{"request": "짧게", "copiedTexts": []any{"a", "b"}}})
	if err != nil {
		t.Fatal(err)
	}
	if got := contents(r); got != "규칙\n\n자바 (섬)\n\n요청: 짧게\n\n1) a\n\n2) b" {
		t.Fatalf("a template's own messages replace the step's: %q", got)
	}
}

func TestValidateTemplateMessages(t *testing.T) {
	chain := ChainOf(models.AIEditRequestCreate)
	if err := chain.ValidateMessages([]prompts.Message{{Role: "user", Content: "{title}만 쓴다"}}); err != nil {
		t.Errorf("a template need not use every input: %v", err)
	}
	for name, msgs := range map[string][]prompts.Message{
		"unknown placeholder": {{Role: "user", Content: "{other}"}},
		"each on a string":    {{Role: "user", Each: "request", Content: "{request}"}},
		"texts outside each":  {{Role: "user", Content: "{copiedTexts}"}},
		"no messages":         {},
	} {
		if err := chain.ValidateMessages(msgs); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}

func TestRenderOwnUserInputs(t *testing.T) {
	cfg := fakeWiki(t)
	prompt := &models.AIEditPrompt{ID: 5, RequestType: models.AIEditRequestCreate,
		UserInputs: `[{"name":"tone","label":"문체"},{"name":"table","type":"bool"}]`,
		Messages:   `[{"role":"user","content":"{title} / {tone}"},{"role":"user","if":"table","content":"표로"}]`}
	r, err := Render(context.Background(), cfg, Request{RequestType: models.AIEditRequestCreate, Title: "사과", Prompt: prompt, UserInputs: map[string]any{"tone": "존댓말", "table": true}})
	if err != nil {
		t.Fatalf("a template's own inputs reach the messages: %v", err)
	}
	if got := contents(r); got != "사과 / 존댓말\n\n표로" {
		t.Fatalf("unexpected %q", got)
	}
	if _, err := Render(context.Background(), cfg, Request{RequestType: models.AIEditRequestCreate, Title: "사과", Prompt: prompt, UserInputs: map[string]any{"hint": "x"}}); !IsInputError(err) {
		t.Errorf("an input the template does not declare is rejected, got %v", err)
	}
	bad := *prompt
	bad.UserInputs = `[{"name":"title"}]`
	if _, err := TemplateChain(bad); err == nil {
		t.Error("a user input named like a task input must be rejected")
	}
}

func TestBuiltinTemplatesValidate(t *testing.T) {
	for _, rt := range []string{models.AIEditRequestCreate, models.AIEditRequestUpdate} {
		chain := ChainOf(rt)
		tpl, err := prompts.ParseTemplate(chain.Template().YAML())
		if err != nil {
			t.Fatal(err)
		}
		c, err := chain.WithUserInputs(tpl.UserInputs)
		if err == nil {
			err = c.ValidateMessages(tpl.Messages)
		}
		if err != nil {
			t.Errorf("%s: a clone of the built-in chain must save as it is: %v", chain.Name, err)
		}
	}
}

func TestRenderEmptiedEditor(t *testing.T) {
	cfg := fakeWiki(t)
	empty := ""
	r, err := Render(context.Background(), cfg, Request{RequestType: models.AIEditRequestUpdate, Title: "자바", Wikitext: &empty})
	if err != nil {
		t.Fatal(err)
	}
	if r.Vars["wikitext"] != "" || strings.Contains(contents(r), "저장된 본문") {
		t.Fatalf("an emptied editor must not bring the saved page back: %q", r.Vars["wikitext"])
	}
	r, err = Render(context.Background(), cfg, Request{RequestType: models.AIEditRequestUpdate, Title: "자바"})
	if err != nil || r.Vars["wikitext"] != "저장된 본문" {
		t.Fatalf("without the editor's text the saved page is read: %v %q", err, r.Vars["wikitext"])
	}
}

func TestTemplateWithoutUserInputs(t *testing.T) {
	tpl, err := prompts.ParseTemplate("messages:\n  - role: user\n    content: \"{title}\"\n")
	if err != nil || tpl.UserInputs == nil || len(tpl.UserInputs) != 0 {
		t.Fatalf("omitted userInputs read as an empty list: %v %#v", err, tpl.UserInputs)
	}
	p := models.AIEditPrompt{RequestType: models.AIEditRequestCreate, UserInputs: "null"}
	if got := TemplateUserInputs(p); got == nil || len(got) != 0 {
		t.Fatalf("a stored null is no inputs, not the chain's: %#v", got)
	}
}
