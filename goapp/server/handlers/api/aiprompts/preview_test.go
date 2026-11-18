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
message:
  system: |-
    규칙 {{풀기:사전링크}} {{ title }}
  user: |-
    {{ title }}: {{ body }}
    {% for t in texts %}
    {{ loop.index }}: {{ t }}
    {% endfor %}
    {% if backlinks %}
    {{ backlinks }}
    {% endif %}
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
	want := []string{"규칙 {{풀기:사전링크}} 새 제목", "새 제목: \n1: 하나\n2: 둘\n첫째\n둘째"}
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

func TestCheckSamples(t *testing.T) {
	got := check(t, map[string]any{"request_type": "update", "spec": updateSpec})
	if !got.OK || len(got.Messages) != 2 || !strings.HasPrefix(got.Messages[1].Content, "Hello World: 문서 시작했다") {
		t.Fatalf("title and body take samples: %+v", got)
	}
}

func TestCheckErrors(t *testing.T) {
	for name, spec := range map[string]string{
		"undeclared variable": strings.Replace(updateSpec, "  - name: backlinks\n", "", 1),
		"unknown variable":    updateSpec + "  - name: hint\n",
		"no name":             strings.Replace(updateSpec, "name: 고치기\n", "", 1),
		"unknown field":       updateSpec + "extra: 1\n",
		"variable typo":       strings.Replace(updateSpec, "규칙", "규칙 {{titel}}", 1),
	} {
		if got := check(t, map[string]any{"request_type": "update", "spec": spec}); got.OK || got.Error == "" {
			t.Errorf("%s: expected an error, got %+v", name, got)
		}
	}
}

func TestPromptViewFormatsQuotedTemplateAsBlock(t *testing.T) {
	quoted := `name: 고치기
message:
  system: "규칙 {{풀기:사전링크}}\n둘째 줄"
  user: "{{ title }}"
input_variables:
  - name: title
`
	view := newPromptView(models.AIEditPrompt{RequestType: "update", Spec: quoted})
	if !strings.Contains(view.Spec, "system: |-\n") {
		t.Fatalf("a multi-line message should be displayed as a YAML block scalar:\n%s", view.Spec)
	}
	if strings.Contains(view.Spec, `\n`) || !strings.Contains(view.Spec, "    규칙 {{풀기:사전링크}}\n    둘째 줄") {
		t.Errorf("message lines should be kept as they are: %q", view.Spec)
	}
}
