package aiprompts

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/zetaoss/zengine/goapp/prompts"
	"github.com/zetaoss/zengine/goapp/server/serverctx"
	"github.com/zetaoss/zengine/goapp/tasks/aiedit"
)

func TestCheckRendersPreview(t *testing.T) {
	body := `{"request_type":"update","user_inputs_yaml":"- texts\n- useBacklinks\n","messages":[{"role":"system","content":"규칙"},{"role":"user","content":"{title}: {body}"},{"role":"user","each":"texts","content":"{index}: {texts}"},{"role":"user","if":"backlinks","content":"{backlinks}"}],"preview_values":{"texts":["하나","","둘"],"useBacklinks":true},"preview_title":"새 제목","preview_body":"","preview_backlinks":[" 첫째 ","", "둘째"]}`
	w := httptest.NewRecorder()
	Check(&serverctx.Context{W: w, R: httptest.NewRequest("POST", "/", strings.NewReader(body))})
	var got struct {
		OK           bool              `json:"ok"`
		Error        string            `json:"error"`
		PreviewError string            `json:"preview_error"`
		Messages     []prompts.Message `json:"rendered_messages"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if !got.OK || got.PreviewError != "" {
		t.Fatalf("preview failed: %s %s", got.Error, got.PreviewError)
	}
	want := []string{"규칙", "새 제목: \n\n1: 하나\n\n2: 둘\n\n첫째\n둘째"}
	if len(got.Messages) != len(want) {
		t.Fatalf("messages: %+v", got.Messages)
	}
	for i, msg := range got.Messages {
		if msg.Content != want[i] {
			t.Errorf("message %d = %q, want %q", i, msg.Content, want[i])
		}
	}
}

func TestValidateEditorRoles(t *testing.T) {
	for _, tc := range []struct {
		messages []prompts.Message
		valid    bool
	}{
		{nil, false},
		{[]prompts.Message{{Role: "user"}}, false},
		{[]prompts.Message{{Role: "system"}, {Role: "system"}}, false},
		{[]prompts.Message{{Role: "system", If: "flag"}, {Role: "user"}}, false},
		{[]prompts.Message{{Role: "system"}, {Role: "user"}, {Role: "user"}}, true},
	} {
		if (validateEditorRoles(tc.messages) == nil) != tc.valid {
			t.Errorf("roles %+v, valid %v", tc.messages, tc.valid)
		}
	}
}

func TestEditorMessages(t *testing.T) {
	inputs := []prompts.Input{{Name: "hint"}, {Name: "useBacklinks", Type: prompts.TypeBool}, {Name: "copiedTexts", Type: prompts.TypeTextList}, {Name: "request", Type: prompts.TypeText}}
	messages := editorMessages("update", inputs, "규칙")
	chain, err := aiedit.ChainOf("update").WithUserInputs(inputs)
	if err != nil {
		t.Fatal(err)
	}
	if err := chain.ValidateMessages(messages); err != nil {
		t.Fatal(err)
	}
	vars := prompts.Vars{"title": "제목", "body": "본문", "hint": "의미", "copiedTexts": []string{"자료1", "자료2"}, "backlinks": []string{"관련1", "관련2"}, "request": ""}
	got := prompts.MergeUserMessages((prompts.Prompt{Messages: messages}).Render(vars))
	if len(got) != 2 {
		t.Fatalf("messages: %+v", got)
	}
	for _, want := range []string{"-----title-----\n제목", "-----body-----\n본문", "-----hint-----\n의미", "-----copiedTexts 1-----\n자료1", "-----copiedTexts 2-----\n자료2", "-----backlinks-----\n관련1\n관련2"} {
		if !strings.Contains(got[1].Content, want) {
			t.Errorf("missing %q in %q", want, got[1].Content)
		}
	}
	if strings.Contains(got[1].Content, "요청사항") {
		t.Errorf("empty request should be omitted: %s", got[1].Content)
	}
}

func TestCheckAutoMessages(t *testing.T) {
	body := `{"request_type":"create","auto_messages":true,"system_prompt":"규칙","user_inputs_yaml":"- name: hint\n","preview_values":{"hint":"예시"}}`
	w := httptest.NewRecorder()
	Check(&serverctx.Context{W: w, R: httptest.NewRequest("POST", "/", strings.NewReader(body))})
	var got struct {
		OK       bool              `json:"ok"`
		Messages []prompts.Message `json:"rendered_messages"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if !got.OK || len(got.Messages) != 2 || got.Messages[0].Content != "규칙" || !strings.Contains(got.Messages[1].Content, "-----hint-----\n예시") {
		t.Fatalf("unexpected response: %s", w.Body.String())
	}
}

func TestPreviewBacklinksEmptyOrDisabled(t *testing.T) {
	for _, tc := range []struct {
		values string
		titles string
	}{
		{`{"useBacklinks":true}`, `[]`},
		{`{"useBacklinks":false}`, `["관련"]`},
	} {
		body := `{"request_type":"create","auto_messages":true,"system_prompt":"규칙","user_inputs_yaml":"- useBacklinks\n","preview_values":` + tc.values + `,"preview_backlinks":` + tc.titles + `}`
		w := httptest.NewRecorder()
		Check(&serverctx.Context{W: w, R: httptest.NewRequest("POST", "/", strings.NewReader(body))})
		var got struct {
			OK       bool              `json:"ok"`
			Messages []prompts.Message `json:"rendered_messages"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
			t.Fatal(err)
		}
		if !got.OK || len(got.Messages) != 2 || strings.Contains(got.Messages[1].Content, "backlinks") {
			t.Fatalf("unexpected response: %s", w.Body.String())
		}
	}
}
