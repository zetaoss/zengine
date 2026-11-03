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
			out = map[string]any{"query": map[string]any{"backlinks": []map[string]any{{"pageid": 7, "title": "과일"}, {"pageid": 8, "title": "과수"}}}}
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
		"hint": "apple", "useBacklinks": true, "texts": []any{"자료 하나", ""}, "request": "",
	}})
	if err != nil {
		t.Fatal(err)
	}
	got := contents(r)
	for _, want := range []string{"-----title-----\n사과", "-----hint-----\napple", "-----texts 1-----\n자료 하나", "-----backlinks-----\n과일\n과수"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in\n%s", want, got)
		}
	}
	if strings.Contains(got, "-----request-----") || strings.Contains(got, "참고 텍스트 2") {
		t.Errorf("an empty request and a blank material must send nothing:\n%s", got)
	}
	if r.Messages[0].Role != "system" || len(r.Messages) != 2 {
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
	if len(r.Messages) != 1 || r.Messages[0].Content != "'자바 (섬)' 문서를 고쳐라.\n저장된 본문\n{{틀}}" {
		t.Fatalf("a template replaces the unconditional messages, and a request from an older run form is ignored: %+v", r.Messages)
	}
	if r.Vars["backlinks"] == nil {
		t.Error("backlinks must be filled (empty) when useBacklinks is off")
	}
	if _, err := Render(context.Background(), cfg, Request{RequestType: models.AIEditRequestUpdate, Title: "자바", UserInputs: map[string]any{"hint": "x"}}); !IsInputError(err) {
		t.Errorf("doc-update has no hint, got %v", err)
	}
}

func TestRenderOwnMessages(t *testing.T) {
	cfg := fakeWiki(t)
	own := `[{"role":"system","content":"규칙"},{"role":"user","content":"{title}"},{"role":"user","if":"request","content":"요청: {request}"},{"role":"user","each":"texts","content":"{index}) {texts}"}]`
	inputs := `[{"name":"request","type":"text"},{"name":"texts","type":"textList"}]`
	prompt := &models.AIEditPrompt{ID: 4, RequestType: models.AIEditRequestCreate, Content: "unused", Messages: own, UserInputs: inputs}
	r, err := Render(context.Background(), cfg, Request{RequestType: models.AIEditRequestCreate, Title: "자바 (섬)", Prompt: prompt, UserInputs: map[string]any{"request": "짧게", "texts": []any{"a", "b"}}})
	if err != nil {
		t.Fatal(err)
	}
	if got := contents(r); got != "규칙\n\n자바 (섬)\n\n1) a\n\n2) b" {
		t.Fatalf("a template's own messages replace the step's; a stored request input is no longer one, and its value is ignored: %q", got)
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
		"no messages":         {},
	} {
		if err := chain.ValidateMessages(msgs); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}

func TestRenderPickedUserInputs(t *testing.T) {
	cfg := fakeWiki(t)
	prompt := &models.AIEditPrompt{ID: 5, RequestType: models.AIEditRequestCreate,
		UserInputs: `[{"name":"extraTexts"},{"name":"hint","placeholder":"old"},{"name":"tone"}]`,
		Messages:   `[{"role":"user","content":"{title} / {hint}"},{"role":"user","each":"extraTexts","content":"{extraTexts}"}]`}
	if got := TemplateUserInputs(*prompt); len(got) != 2 || got[0].Name != "hint" || got[0].Placeholder == "old" || got[1].Name != "texts" || got[1].Type != prompts.TypeTextList {
		t.Fatalf("a template picks its chain's inputs by name, as the chain defines them and in its order: %+v", got)
	}
	r, err := Render(context.Background(), cfg, Request{RequestType: models.AIEditRequestCreate, Title: "사과", Prompt: prompt, UserInputs: map[string]any{"hint": "과일", "texts": []any{"자료"}}})
	if err != nil {
		t.Fatal(err)
	}
	if got := contents(r); got != "사과 / 과일\n\n자료" {
		t.Fatalf("unexpected %q", got)
	}
	if _, err := Render(context.Background(), cfg, Request{RequestType: models.AIEditRequestCreate, Title: "사과", Prompt: prompt, UserInputs: map[string]any{"tone": "x"}}); !IsInputError(err) {
		t.Errorf("an input the template does not pick is rejected, got %v", err)
	}
}

func TestPickUserInputs(t *testing.T) {
	got, err := PickUserInputs(models.AIEditRequestCreate, []prompts.Input{{Name: "texts"}, {Name: "readBacklinks"}})
	if err != nil || len(got) != 2 || got[0].Name != "useBacklinks" || got[0].Type != prompts.TypeBool || got[1].Type != prompts.TypeTextList {
		t.Fatalf("names take the chain's definitions and order, older names their names now: %+v %v", got, err)
	}
	for name, given := range map[string][]prompts.Input{
		"unknown":    {{Name: "tone"}},
		"request":    {{Name: "request"}},
		"repeated":   {{Name: "hint"}, {Name: "hint"}},
		"attributes": {{Name: "hint", Placeholder: "x"}},
		"not update": nil,
	} {
		requestType := models.AIEditRequestCreate
		if name == "not update" {
			requestType, given = models.AIEditRequestUpdate, []prompts.Input{{Name: "hint"}}
		}
		if _, err := PickUserInputs(requestType, given); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}

func TestRenderUnsetInputIsEmpty(t *testing.T) {
	cfg := fakeWiki(t)
	prompt := &models.AIEditPrompt{ID: 7, RequestType: models.AIEditRequestCreate,
		UserInputs: `[{"name":"hint"}]`,
		Messages:   `[{"role":"user","content":"{title} / {hint} / {example}"}]`}
	r, err := Render(context.Background(), cfg, Request{RequestType: models.AIEditRequestCreate, Title: "Apple", Prompt: prompt})
	if err != nil {
		t.Fatal(err)
	}
	if got := contents(r); got != "Apple /  / {example}" {
		t.Fatalf("a declared input left unset renders as nothing, an undeclared name stays: %q", got)
	}
}

func TestRenderLegacyBacklinkTemplate(t *testing.T) {
	cfg := fakeWiki(t)
	prompt := &models.AIEditPrompt{RequestType: models.AIEditRequestCreate, Messages: `[{"role":"user","each":"backlinkTexts","content":"{backlinkTexts}"}]`}
	rendered, err := Render(context.Background(), cfg, Request{RequestType: models.AIEditRequestCreate, Title: "사과", Prompt: prompt, UserInputs: map[string]any{"readBacklinks": true}})
	if err != nil {
		t.Fatal(err)
	}
	if got := contents(rendered); got != "과일\n\n과수" {
		t.Fatalf("legacy template: got %q", got)
	}
	if err := rendered.Chain.ValidateMessages(rendered.Chain.PromptStep().Messages); err != nil {
		t.Fatalf("legacy template must remain valid: %v", err)
	}
}
