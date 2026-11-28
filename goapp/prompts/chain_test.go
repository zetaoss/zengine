package prompts

import (
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestBuiltinChains(t *testing.T) {
	seen := map[string]bool{}
	for _, c := range BuiltinChains() {
		if c.Name == "" || seen[c.Name] {
			t.Fatalf("invalid or duplicate chain: %q", c.Name)
		}
		seen[c.Name] = true
		if err := c.Compile(); err != nil {
			t.Errorf("%s: %v", c.Name, err)
		}
	}
}

var testMessages = MessageTemplates{
	System: `고정 지침 {"id": 1} {{틀}} {| class="wikitable"`,
	User: `<title>
{{ title }}
</title>
{% if hint %}
<hint>
{{ hint }}
</hint>
{% endif %}`,
}

func testChain(m MessageTemplates) Chain {
	return Chain{
		Name: "t", Version: "v1.0",
		Inputs: []Input{{Name: "title"}, {Name: "hint"}},
		Steps:  []Step{{Name: "s", Message: m, OutputVariable: &OutputVariable{Name: "out"}}},
	}
}

func compiled(t *testing.T, c Chain) Chain {
	t.Helper()
	if err := c.Compile(); err != nil {
		t.Fatalf("valid chain rejected: %v", err)
	}
	return c
}

func TestCompile(t *testing.T) {
	c := compiled(t, testChain(testMessages))
	if got := c.Steps[0].Inputs; !reflect.DeepEqual(got, []string{"title", "hint"}) {
		t.Errorf("Inputs = %v", got)
	}
	for name, m := range map[string]MessageTemplates{
		"undeclared variable": {User: "{{ title }} {{ other }}"},
		"for over a string":   {User: "{{ title }}{% for h in hint %}{% endfor %}"},
		"title unused":        {User: "{{ hint }}"},
		"hint unused":         {User: "{{ title }}"},
		"no message":          {System: "\n", User: " "},
		"error in system":     {System: "{{ titel }}", User: "{{ title }}"},
	} {
		if c := testChain(m); c.Compile() == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
	dup := testChain(testMessages)
	dup.Inputs = []Input{{Name: "title"}, {Name: "hint"}, {Name: "title"}}
	if err := dup.Compile(); err == nil {
		t.Error("a duplicate variable must be rejected")
	}
	noVersion := testChain(testMessages)
	noVersion.Version = "v2"
	if err := noVersion.Compile(); err == nil {
		t.Error("a version without a minor must be rejected")
	}
}

func TestRender(t *testing.T) {
	c := compiled(t, testChain(testMessages))
	got := c.Steps[0].Render(Vars{"title": "T {{ hint }}", "hint": " "})
	want := []Message{
		{Role: "system", Content: `고정 지침 {"id": 1} {{틀}} {| class="wikitable"`},
		{Role: "user", Content: "<title>\nT {{ hint }}\n</title>"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected render:\n%+v\nwant\n%+v", got, want)
	}
	if got := c.Steps[0].Render(Vars{"title": "T", "hint": "x"}); len(got) != 2 || got[1].Content != "<title>\nT\n</title>\n<hint>\nx\n</hint>" {
		t.Fatalf("a given hint: %+v", got)
	}
}

func TestInputYAML(t *testing.T) {
	if DocEval.Inputs[0].Name != "title" {
		t.Fatalf("an input written as its name must load: %+v", DocEval.Inputs)
	}
	var in []Input
	if err := yaml.Unmarshal([]byte("- title\n- name: wikitext\n  description: 본문\n"), &in); err != nil {
		t.Fatal(err)
	}
	if want := []Input{{Name: "title"}, {Name: "wikitext", Description: "본문"}}; !reflect.DeepEqual(in, want) {
		t.Fatalf("inputs as a name and as a mapping: %+v", in)
	}
}

func TestParse(t *testing.T) {
	var v struct {
		A int `json:"a"`
	}
	if err := FormatJSON.Parse("```json\n{\"a\": 1}\n```", &v); err != nil || v.A != 1 {
		t.Fatalf("json in a code fence: %v %+v", err, v)
	}
	if err := FormatJSON.Parse(`{"a": 2}`, &v); err != nil || v.A != 2 {
		t.Fatalf("plain json: %v %+v", err, v)
	}
	if err := FormatJSON.Parse("not json", &v); err == nil {
		t.Fatal("invalid json parsed")
	}
	var s string
	if err := FormatText.Parse("  == 개요 ==\n본문\n", &s); err != nil || s != "== 개요 ==\n본문" {
		t.Fatalf("text: %v %q", err, s)
	}
	if err := FormatText.Parse("x", &v); err == nil {
		t.Fatal("text into a non-string accepted")
	}
	if err := Format("").Parse("```\n== 개요 ==\n```", &s); err != nil || s != "== 개요 ==" {
		t.Fatalf("text is the default and its code fence is stripped: %v %q", err, s)
	}
	if err := Format("xml").Parse("x", &s); err == nil {
		t.Fatal("unknown format accepted")
	}
}

func TestMajor(t *testing.T) {
	for v, want := range map[string]string{"v2.3": "v2", "v2": "v2", "v10.0": "v10"} {
		if got := Major(v); got != want {
			t.Errorf("Major(%q) = %q, want %q", v, got, want)
		}
	}
	if !SameMajor("v2", "v2.1") || !SameMajor("v2.0", "v2.9") || SameMajor("v2.1", "v3.0") || SameMajor("v1.0", "v10.0") {
		t.Error("SameMajor must compare majors only")
	}
}

func TestFormatParseJSON(t *testing.T) {
	type evalResult struct {
		GeneralQuality []struct {
			ID      string `json:"id"`
			Reason  string `json:"reason"`
			Verdict string `json:"verdict"`
		} `json:"general_quality"`
		Summary string `json:"summary"`
	}

	cases := []struct {
		name        string
		raw         string
		wantVerdict string
		wantReason  string
		wantSummary string
	}{
		{
			name:        "plain json",
			raw:         `{"general_quality":[{"id":"g1","reason":"정상적인 설명","verdict":"pass"}],"summary":"요약"}`,
			wantVerdict: "pass",
			wantReason:  "정상적인 설명",
			wantSummary: "요약",
		},
		{
			name: "markdown fence with outer commentary",
			raw: "Here is the evaluation result:\n```json\n" +
				`{"general_quality":[{"id":"g1","reason":"코드블록 안의 설명","verdict":"pass"}],"summary":"요약"}` +
				"\n```\nHope this helps!",
			wantVerdict: "pass",
			wantReason:  "코드블록 안의 설명",
			wantSummary: "요약",
		},
		{
			name: "backticks around value and inner backticks in string",
			raw: "{\n" +
				`  "general_quality": [` + "\n" +
				"    {\n" +
				`      "id": "g1",` + "\n" +
				`      "reason": "문서에 딕셔너리 병합 설명이 없으며, ` + "`update()`" + `와 차이가 없다.",` + "\n" +
				`      "verdict": ` + "`fail`" + "\n" +
				"    }\n" +
				"  ],\n" +
				`  "summary": "품질 부족"` + "\n" +
				"}\n```",
			wantVerdict: "fail",
			wantReason:  "문서에 딕셔너리 병합 설명이 없으며, `update()`와 차이가 없다.",
			wantSummary: "품질 부족",
		},
		{
			name: "raw newlines in strings",
			raw: "{\n" +
				`  "general_quality": [{"id":"g1","reason":"첫째 줄\n둘째 줄","verdict":` + "`pass`" + `}],` + "\n" +
				`  "summary": "개행 포함 요약"` + "\n" +
				"}",
			wantVerdict: "pass",
			wantReason:  "첫째 줄\n둘째 줄",
			wantSummary: "개행 포함 요약",
		},
		{
			name: "reasoning tags before json",
			raw: "<think>analyzing the dictionary merge page</think>\n" +
				`{"general_quality":[{"id":"g1","reason":"사유","verdict":"fail"}],"summary":"완료"}`,
			wantVerdict: "fail",
			wantReason:  "사유",
			wantSummary: "완료",
		},
		{
			name: "reasoning tags with draft json strips reasoning and selects final json",
			raw: `<think>Consider draft {"general_quality":[{"id":"g1","reason":"draft","verdict":"fail"}],"summary":"draft"}; revise it.</think>` + "\n" +
				`{"general_quality":[{"id":"g1","reason":"final","verdict":"pass"}],"summary":"final"}`,
			wantVerdict: "pass",
			wantReason:  "final",
			wantSummary: "final",
		},
		{
			name: "preserves reasoning-tag literals inside JSON strings",
			raw: "Here is the result:\n" +
				`{"general_quality":[{"id":"g1","reason":"The <think>example</think> tag is present.","verdict":"pass"}],"summary":"The <think>tag</think> remains."}`,
			wantVerdict: "pass",
			wantReason:  "The <think>example</think> tag is present.",
			wantSummary: "The <think>tag</think> remains.",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var out evalResult
			if err := FormatJSON.Parse(c.raw, &out); err != nil {
				t.Fatalf("Parse failed: %v", err)
			}
			if len(out.GeneralQuality) != 1 {
				t.Fatalf("expected 1 general_quality item, got %d", len(out.GeneralQuality))
			}
			got := out.GeneralQuality[0]
			if got.Verdict != c.wantVerdict {
				t.Errorf("verdict: got %q, want %q", got.Verdict, c.wantVerdict)
			}
			if got.Reason != c.wantReason {
				t.Errorf("reason: got %q, want %q", got.Reason, c.wantReason)
			}
			if out.Summary != c.wantSummary {
				t.Errorf("summary: got %q, want %q", out.Summary, c.wantSummary)
			}
		})
	}
}
