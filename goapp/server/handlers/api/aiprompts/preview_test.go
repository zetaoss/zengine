package aiprompts

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/zetaoss/zengine/goapp/models"
	"github.com/zetaoss/zengine/goapp/prompts"
	"github.com/zetaoss/zengine/goapp/server/serverctx"
)

type checkResponse struct {
	OK           bool              `json:"ok"`
	Error        string            `json:"error"`
	PreviewError string            `json:"preview_error"`
	UserInputs   []prompts.Input   `json:"user_inputs"`
	Spans        []span            `json:"spans"`
	Messages     []prompts.Message `json:"rendered_messages"`
}

func check(t *testing.T, body map[string]any) checkResponse {
	t.Helper()
	b, _ := json.Marshal(body)
	w := httptest.NewRecorder()
	Check(&serverctx.Context{W: w, R: httptest.NewRequest("POST", "/", strings.NewReader(string(b)))})
	var got checkResponse
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("%v: %s", err, w.Body.String())
	}
	return got
}

const updateSpec = `name: 고치기
template_format: jinja2
template: |-
  <message role="system">
  규칙 {{풀기:사전링크}}
  </message>
  <message role="user">
  {{ title }}: {{ body }}
  {% for t in texts %}
  {{ loop.index }}: {{ t }}
  {% endfor %}
  {% if backlinks %}
  {{ backlinks }}
  {% endif %}
  </message>
input_variables:
  - name: title
  - name: body
  - name: texts
  - name: backlinks
`

func TestCheckRendersPreview(t *testing.T) {
	got := check(t, map[string]any{
		"request_type": "update", "spec": updateSpec,
		"preview_values": map[string]any{"texts": []string{"하나", "", "둘"}, "backlinks": []string{"첫째", "둘째"}, "title": "새 제목", "body": ""},
	})
	if !got.OK || got.PreviewError != "" {
		t.Fatalf("preview failed: %s %s", got.Error, got.PreviewError)
	}
	want := []string{"규칙 {{풀기:사전링크}}", "새 제목: \n1: 하나\n2: 둘\n첫째\n둘째"}
	if len(got.Messages) != len(want) {
		t.Fatalf("messages: %+v", got.Messages)
	}
	for i, msg := range got.Messages {
		if msg.Content != want[i] {
			t.Errorf("message %d = %q, want %q", i, msg.Content, want[i])
		}
	}
	if len(got.UserInputs) != 2 || got.UserInputs[0].Name != "texts" || got.UserInputs[1].Name != "backlinks" {
		t.Errorf("run form: %+v", got.UserInputs)
	}
}

func TestCheckSamplesAndSpans(t *testing.T) {
	got := check(t, map[string]any{"request_type": "update", "spec": updateSpec})
	if !got.OK || len(got.Messages) != 2 || !strings.HasPrefix(got.Messages[1].Content, "Hello World: 문서 시작했다") {
		t.Fatalf("title and body take samples: %+v", got)
	}
	lines := strings.Split(updateSpec, "\n")
	var marked []string
	for _, s := range got.Spans {
		r := []rune(lines[s.Line-1])
		marked = append(marked, string(r[s.Start-1:s.End-1]))
	}
	want := "{{ title }}|{{ body }}|{% for t in texts %}|{{ loop.index }}|{{ t }}|{% endfor %}|{% if backlinks %}|{{ backlinks }}|{% endif %}"
	if strings.Join(marked, "|") != want {
		t.Errorf("spans mark %v", marked)
	}
}

func TestCheckErrors(t *testing.T) {
	for name, spec := range map[string]string{
		"undeclared variable": strings.Replace(updateSpec, "  - name: backlinks\n", "", 1),
		"unknown variable":    updateSpec + "  - name: hint\n",
		"no name":             strings.Replace(updateSpec, "name: 고치기\n", "", 1),
		"unknown field":       updateSpec + "extra: 1\n",
		"wiki name":           strings.Replace(updateSpec, "규칙", "규칙 {{PAGENAME}}", 1),
	} {
		if got := check(t, map[string]any{"request_type": "update", "spec": spec}); got.OK || got.Error == "" {
			t.Errorf("%s: expected an error, got %+v", name, got)
		}
	}
}

func TestPromptViewFormatsQuotedTemplateAsBlock(t *testing.T) {
	quoted := `name: 고치기
template_format: jinja2
template: "<message role=\"system\">\n규칙 {{풀기:사전링크}}\n</message>\n<message role=\"user\">\n{{ title }}\n</message>"
input_variables:
  - name: title
`
	view := newPromptView(models.AIEditPrompt{RequestType: "update", Spec: quoted})
	if !strings.Contains(view.Spec, "template: |-\n") {
		t.Fatalf("template should be displayed as a YAML block scalar:\n%s", view.Spec)
	}
	if strings.Contains(view.Spec, `\n`) || !strings.Contains(view.Spec, "<message role=\"system\">\n  규칙") {
		t.Errorf("template contents should be preserved as normal lines: %q", view.Spec)
	}
}
