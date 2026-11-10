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
		"hint": "apple", "backlinks": []any{"과일", "과수"}, "texts": []any{"자료 하나", ""},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if r.Messages[0].Role != "system" || len(r.Messages) != 2 {
		t.Fatalf("unexpected messages %+v", r.Messages)
	}
	want := "# 변수\n\n<title>\n사과\n</title>\n<hint>\napple\n</hint>\n<texts>\n자료 하나\n</texts>\n<backlinks>\n과일\n과수\n</backlinks>"
	if got := r.Messages[1].Content; got != want {
		t.Errorf("user message:\n%s\nwant\n%s", got, want)
	}
}

func TestRenderUpdateVariables(t *testing.T) {
	cfg := fakeWiki(t)
	r, err := Render(context.Background(), cfg, Request{RequestType: models.AIEditRequestUpdate, Title: "사과"})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Messages) != 2 || r.Messages[1].Content != "# 변수\n\n<title>\n사과\n</title>\n<body>\n저장된 본문\n</body>" {
		t.Fatalf("unexpected update messages: %+v", r.Messages)
	}
	if _, err := Render(context.Background(), cfg, Request{RequestType: models.AIEditRequestUpdate, Title: "자바", UserInputs: map[string]any{"hint": "x"}}); !IsInputError(err) {
		t.Errorf("doc-update has no hint, got %v", err)
	}
}

func TestRenderLimits(t *testing.T) {
	cfg := fakeWiki(t)
	many := make([]any, MaxListItems+1)
	for i := range many {
		many[i] = "x"
	}
	for name, inputs := range map[string]map[string]any{
		"too many backlinks": {"backlinks": many},
		"too long a hint":    {"hint": strings.Repeat("가", MaxValueChars+1)},
	} {
		if _, err := Render(context.Background(), cfg, Request{RequestType: models.AIEditRequestCreate, Title: "사과", UserInputs: inputs}); !IsInputError(err) {
			t.Errorf("%s: got %v", name, err)
		}
	}
}

const ownSpec = `name: 짧은 문서
template_format: jinja2
template: |-
  <message role="system">
  규칙 {{풀기:사전링크}}
  </message>
  <message role="user">
  {{ title }}
  {% for t in texts %}
  {{ loop.index }}) {{ t }}
  {% endfor %}
  </message>
input_variables:
  - name: title
  - name: texts
`

func TestRenderTemplate(t *testing.T) {
	cfg := fakeWiki(t)
	prompt := &models.AIEditPrompt{ID: 4, RequestType: models.AIEditRequestCreate, Spec: ownSpec}
	r, err := Render(context.Background(), cfg, Request{RequestType: models.AIEditRequestCreate, Title: "자바 (섬)", Prompt: prompt, UserInputs: map[string]any{"texts": []any{"a", "b"}}})
	if err != nil {
		t.Fatal(err)
	}
	if got := contents(r); got != "규칙 {{풀기:사전링크}}\n\n자바 (섬)\n1) a\n2) b" {
		t.Fatalf("a template runs its own messages: %q", got)
	}
	if _, err := Render(context.Background(), cfg, Request{RequestType: models.AIEditRequestCreate, Title: "사과", Prompt: prompt, UserInputs: map[string]any{"hint": "x"}}); !IsInputError(err) {
		t.Errorf("an input the template does not pick is rejected, got %v", err)
	}
	if _, err := Render(context.Background(), cfg, Request{RequestType: models.AIEditRequestCreate, Title: "사과", Prompt: &models.AIEditPrompt{ID: 5}}); err == nil {
		t.Error("a template without a spec cannot run")
	}
}

func ptr(s string) *string { return &s }

func TestLegacySpecContent(t *testing.T) {
	content := "# 목표\n'{title}' 문서를 고쳐라.\n{wikitext}\n\n{추가 정보}\n{{풀기:사전링크}} {{PAGENAME}} {{#if:x}} {% x\n{# 추가 정보:code}"
	spec, err := LegacySpec(legacyPrompt{ID: 2, Title: "개선", RequestType: models.AIEditRequestUpdate, Content: ptr(content)})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(spec, "{% raw %}{# 추가 정보:code}{% endraw %}") {
		t.Errorf("wiki comment-like markup should be protected as a whole raw block:\n%s", spec)
	}
	if !strings.Contains(spec, "{{#if:x}}") {
		t.Errorf("a wiki parser function should stay unmodified:\n%s", spec)
	}
	p := models.AIEditPrompt{ID: 2, RequestType: models.AIEditRequestUpdate, Spec: spec}
	r, err := Render(context.Background(), fakeWiki(t), Request{RequestType: models.AIEditRequestUpdate, Title: "자바", Prompt: &p, UserInputs: map[string]any{"texts": []any{"자료"}}})
	if err != nil {
		t.Fatalf("%v\n%s", err, spec)
	}
	if want := "# 목표\n'자바' 문서를 고쳐라.\n저장된 본문\n\n{추가 정보}\n{{풀기:사전링크}} {{PAGENAME}} {{#if:x}} {% x\n{# 추가 정보:code}"; r.Messages[0].Content != want {
		t.Errorf("system:\n%s\nwant\n%s", r.Messages[0].Content, want)
	}
	if want := "# 변수\n\n<texts>\n자료\n</texts>\n# 대상 문서\n<title>\n자바\n</title>"; r.Messages[1].Content != want {
		t.Errorf("the user message lists what the instructions do not read:\n%s\nwant\n%s", r.Messages[1].Content, want)
	}
}

func TestLegacySpecMessages(t *testing.T) {
	messages := `[{"role":"system","content":"지침 틀(` + "`{{…}}`" + `)"},{"role":"user","content":"-----title-----\n{title}"},{"role":"user","if":"backlinkTitles","content":"-----backlinkTitles-----\n{backlinkTitles}"},{"role":"user","each":"extraTexts","content":"-----extraTexts {index}-----\n{extraTexts}"},{"role":"user","if":"request","content":"-----request-----\n{request}"}]`
	inputs := `[{"name":"hint"},{"name":"readBacklinks","type":"bool"},{"name":"extraTexts","type":"textList"},{"name":"request","type":"text"}]`
	spec, err := LegacySpec(legacyPrompt{ID: 8, Title: "내 템플릿", Description: "설명", RequestType: models.AIEditRequestCreate, Messages: ptr(messages), UserInputs: ptr(inputs), Model: "groq/x"})
	if err != nil {
		t.Fatal(err)
	}
	pt, err := prompts.ParsePromptTemplate(spec)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, v := range pt.InputVariables {
		names = append(names, v.Name)
	}
	if strings.Join(names, ",") != "title,hint,texts,backlinks" || pt.Name != "내 템플릿" || pt.Description != "설명" {
		t.Errorf("older names take their names now and request is dropped: %v %+v", names, pt)
	}
	p := models.AIEditPrompt{RequestType: models.AIEditRequestCreate, Spec: spec}
	r, err := Render(context.Background(), fakeWiki(t), Request{RequestType: models.AIEditRequestCreate, Title: "사과", Prompt: &p, UserInputs: map[string]any{"hint": "과일", "backlinks": []any{"과수"}, "texts": []any{"가", "나"}}})
	if err != nil {
		t.Fatal(err)
	}
	want := "-----title-----\n사과\n\n-----backlinks-----\n과수\n\n-----texts 1-----\n가\n\n-----texts 2-----\n나\n\n# 변수\n\n<hint>\n과일\n</hint>"
	if len(r.Messages) != 2 || r.Messages[0].Content != "지침 틀(`{{…}}`)" || r.Messages[1].Content != want {
		t.Errorf("messages are kept as they ran, and what they do not read is listed after:\n%+v\nwant user\n%s", r.Messages, want)
	}
}

func TestLegacySpecUserInstructions(t *testing.T) {
	messages := `[{"role":"user","content":"'{title}' 문서를 표로 정리한다.\n{wikitext}"},{"role":"user","each":"copiedTexts","content":"자료 {index}: {copiedTexts}"}]`
	spec, err := LegacySpec(legacyPrompt{ID: 9, Title: "복제", RequestType: models.AIEditRequestUpdate, Messages: ptr(messages)})
	if err != nil {
		t.Fatal(err)
	}
	p := models.AIEditPrompt{RequestType: models.AIEditRequestUpdate, Spec: spec}
	r, err := Render(context.Background(), fakeWiki(t), Request{RequestType: models.AIEditRequestUpdate, Title: "자바", Prompt: &p, UserInputs: map[string]any{"texts": []any{"가"}}})
	if err != nil {
		t.Fatalf("%v\n%s", err, spec)
	}
	if len(r.Messages) != 1 || r.Messages[0].Role != "user" || r.Messages[0].Content != "'자바' 문서를 표로 정리한다.\n저장된 본문\n\n자료 1: 가" {
		t.Errorf("instructions in a user message must survive: %+v\n%s", r.Messages, spec)
	}
}

func TestLegacySpecTripleBraceWikiParameter(t *testing.T) {
	content := "Literal {{{title}}} and {title}"
	spec, err := LegacySpec(legacyPrompt{ID: 10, Title: "이름", RequestType: models.AIEditRequestUpdate, Content: ptr(content)})
	if err != nil {
		t.Fatal(err)
	}
	p := models.AIEditPrompt{RequestType: models.AIEditRequestUpdate, Spec: spec}
	r, err := Render(context.Background(), fakeWiki(t), Request{RequestType: models.AIEditRequestUpdate, Title: "사과", Prompt: &p})
	if err != nil {
		t.Fatalf("%v\n%s", err, spec)
	}
	if got, want := r.Messages[0].Content, "Literal {{{title}}} and 사과"; got != want {
		t.Errorf("triple-brace wiki parameters stay literal while normal placeholders render: %q, want %q", got, want)
	}
}

func TestLegacySpecWikiTemplateKeepsFollowingNewline(t *testing.T) {
	content := "Literal {{PAGENAME}}\nNext line {title}"
	spec, err := LegacySpec(legacyPrompt{ID: 13, Title: "줄바꿈", RequestType: models.AIEditRequestUpdate, Content: ptr(content)})
	if err != nil {
		t.Fatal(err)
	}
	p := models.AIEditPrompt{RequestType: models.AIEditRequestUpdate, Spec: spec}
	r, err := Render(context.Background(), fakeWiki(t), Request{RequestType: models.AIEditRequestUpdate, Title: "사과", Prompt: &p})
	if err != nil {
		t.Fatalf("%v\n%s", err, spec)
	}
	if got, want := r.Messages[0].Content, "Literal {{PAGENAME}}\nNext line 사과"; got != want {
		t.Errorf("a raw-wrapped wiki template keeps its following newline: %q, want %q", got, want)
	}
}

func TestLegacySpecIfAndEach(t *testing.T) {
	messages := `[{"role":"user","if":"hint","each":"extraTexts","content":"자료 {index}: {extraTexts}"}]`
	spec, err := LegacySpec(legacyPrompt{ID: 11, Title: "조건 반복", RequestType: models.AIEditRequestCreate, Messages: ptr(messages)})
	if err != nil {
		t.Fatal(err)
	}
	p := models.AIEditPrompt{RequestType: models.AIEditRequestCreate, Spec: spec}
	render := func(hint string) string {
		t.Helper()
		r, err := Render(context.Background(), fakeWiki(t), Request{RequestType: models.AIEditRequestCreate, Title: "사과", Prompt: &p, UserInputs: map[string]any{"hint": hint, "texts": []any{"A", "B"}}})
		if err != nil {
			t.Fatalf("%v\n%s", err, spec)
		}
		return r.Messages[0].Content
	}
	if got := render(""); strings.Contains(got, "자료 1: A") || strings.Contains(got, "자료 2: B") {
		t.Errorf("each block must remain conditional on hint when it is absent: %q", got)
	}
	if got := render("yes"); !strings.Contains(got, "자료 1: A") || !strings.Contains(got, "자료 2: B") {
		t.Errorf("each block must render when hint is present: %q", got)
	}
}

func TestLegacySpecContentKeepsUserMessageWhenOptionalInputsAreEmpty(t *testing.T) {
	content := "'{title}' 문서를 고쳐라.\n{wikitext}"
	emptyInputs := `[]`
	spec, err := LegacySpec(legacyPrompt{ID: 12, Title: "최소 요청", RequestType: models.AIEditRequestUpdate, Content: ptr(content), UserInputs: &emptyInputs})
	if err != nil {
		t.Fatal(err)
	}
	p := models.AIEditPrompt{RequestType: models.AIEditRequestUpdate, Spec: spec}
	r, err := Render(context.Background(), fakeWiki(t), Request{RequestType: models.AIEditRequestUpdate, Title: "자바", Prompt: &p})
	if err != nil {
		t.Fatalf("a converted content-only template keeps a runnable user message:\n%v\n%s", err, spec)
	}
	if len(r.Messages) != 2 || r.Messages[1].Role != "user" || !strings.Contains(r.Messages[1].Content, "<title>\n자바\n</title>") {
		t.Errorf("converted template has a nonempty user message with the target title: %+v", r.Messages)
	}
}
