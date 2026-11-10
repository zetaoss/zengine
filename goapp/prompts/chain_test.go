package prompts

import (
	"reflect"
	"slices"
	"strings"
	"testing"
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

const testTemplate = `<message role="system">
고정 지침 {"id": 1} {{틀}} {| class="wikitable"
</message>
<message role="user">
<title>
{{ title }}
</title>
{% if hint %}
<hint>
{{ hint }}
</hint>
{% endif %}
{% for n in notes %}
<notes>
{{ loop.index }}. {{ n }}
</notes>
{% endfor %}
</message>`

func testChain(template string) Chain {
	return Chain{
		Name: "t", Version: "v1.0",
		UserInputs: []Input{
			{Name: "hint"},
			{Name: "flag", JSONSchema: &JSONSchema{Type: "boolean"}},
			{Name: "notes", JSONSchema: &JSONSchema{Type: "array", Items: &JSONSchema{Type: "string"}}},
		},
		Inputs: []Input{{Name: "title"}},
		Steps:  []Step{{Name: "s", TemplateFormat: "jinja2", Template: template, OutputVariable: &OutputVariable{Name: "out"}}},
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
	c := compiled(t, testChain(testTemplate))
	if got := c.Steps[0].Inputs; !reflect.DeepEqual(got, []string{"title", "hint", "notes"}) {
		t.Errorf("Inputs = %v", got)
	}
	for name, tpl := range map[string]string{
		"undeclared variable": "<message role=\"user\">\n{{ title }} {{ other }}\n</message>",
		"for over a string":   "<message role=\"user\">\n{{ title }}{% for h in hint %}{% endfor %}\n</message>",
		"assistant message":   "<message role=\"assistant\">\n{{ title }}\n</message>",
		"title unused":        "<message role=\"user\">\n{{ hint }}\n</message>",
		"text outside":        "x\n<message role=\"user\">\n{{ title }}\n</message>",
	} {
		if c := testChain(tpl); c.Compile() == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
	other := testChain(testTemplate)
	other.Steps[0].TemplateFormat = "handlebars"
	if err := other.Compile(); err == nil {
		t.Error("a template format other than jinja2 must be rejected")
	}
	dup := testChain(testTemplate)
	dup.Inputs = []Input{{Name: "title"}, {Name: "hint"}}
	if err := dup.Compile(); err == nil {
		t.Error("a duplicate variable must be rejected")
	}
	noVersion := testChain(testTemplate)
	noVersion.Version = "v2"
	if err := noVersion.Compile(); err == nil {
		t.Error("a version without a minor must be rejected")
	}
}

func TestRender(t *testing.T) {
	c := compiled(t, testChain(testTemplate))
	got := c.Steps[0].Render(Vars{"title": "T", "hint": " ", "notes": []string{"a", "b {{ title }}"}})
	want := []Message{
		{Role: "system", Content: `고정 지침 {"id": 1} {{틀}} {| class="wikitable"`},
		{Role: "user", Content: "<title>\nT\n</title>\n<notes>\n1. a\n</notes>\n<notes>\n2. b {{ title }}\n</notes>"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected render:\n%+v\nwant\n%+v", got, want)
	}
	if got := c.Steps[0].Render(Vars{"title": "T", "hint": "x"}); len(got) != 2 || got[1].Content != "<title>\nT\n</title>\n<hint>\nx\n</hint>" {
		t.Fatalf("a given hint and no notes: %+v", got)
	}
}

func TestUserVars(t *testing.T) {
	c := compiled(t, testChain(testTemplate))
	vars, err := c.UserVars(map[string]any{"hint": "apple", "flag": true, "notes": []any{"a", "  ", "b"}})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(vars, Vars{"hint": "apple", "flag": true, "notes": []string{"a", "b"}}) {
		t.Fatalf("unexpected vars %+v", vars)
	}
	for _, raw := range []map[string]any{{"other": "x"}, {"hint": 1}, {"flag": "yes"}, {"notes": "a"}, {"notes": []any{1}}} {
		if _, err := c.UserVars(raw); err == nil {
			t.Errorf("expected an error for %v", raw)
		}
	}
}

func TestInputTypes(t *testing.T) {
	str := &JSONSchema{Type: "string"}
	for want, in := range map[InputType]Input{
		TypeString:   {Name: "a"},
		TypeText:     {Name: "a", JSONSchema: str, UI: &UI{Variant: "longText"}},
		TypeChoice:   {Name: "a", JSONSchema: &JSONSchema{Type: "string", Enum: []string{"x", "y"}}},
		TypeBool:     {Name: "a", JSONSchema: &JSONSchema{Type: "boolean"}},
		TypeTextList: {Name: "a", JSONSchema: &JSONSchema{Type: "array", Items: str}, UI: &UI{Variant: "longText"}},
	} {
		if err := in.validateUser(); err != nil || in.Type() != want {
			t.Errorf("%+v: type %s, err %v; want %s", in, in.Type(), err, want)
		}
	}
	for name, in := range map[string]Input{
		"number":                 {Name: "x", JSONSchema: &JSONSchema{Type: "number"}},
		"array of booleans":      {Name: "x", JSONSchema: &JSONSchema{Type: "array", Items: &JSONSchema{Type: "boolean"}}},
		"array without items":    {Name: "x", JSONSchema: &JSONSchema{Type: "array"}},
		"choice with one value":  {Name: "x", JSONSchema: &JSONSchema{Type: "string", Enum: []string{"a"}}},
		"duplicate enum":         {Name: "x", JSONSchema: &JSONSchema{Type: "string", Enum: []string{"a", "a"}}},
		"default not in enum":    {Name: "x", JSONSchema: &JSONSchema{Type: "string", Enum: []string{"a", "b"}}, Default: "c"},
		"string default on bool": {Name: "x", JSONSchema: &JSONSchema{Type: "boolean"}, Default: "true"},
		"default on an array":    {Name: "x", JSONSchema: &JSONSchema{Type: "array", Items: str}, Default: "a"},
		"required bool":          {Name: "x", JSONSchema: &JSONSchema{Type: "boolean"}, IsRequired: true},
		"unknown variant":        {Name: "x", UI: &UI{Variant: "obscured"}},
		"ui on a bool":           {Name: "x", JSONSchema: &JSONSchema{Type: "boolean"}, UI: &UI{Placeholder: "p"}},
		"reserved name":          {Name: "loop"},
	} {
		if err := in.validateUser(); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
	tone := Input{Name: "tone", JSONSchema: &JSONSchema{Type: "string", Enum: []string{"존댓말", "반말"}}, Default: "존댓말"}
	topic := Input{Name: "topic", IsRequired: true}
	c := testChain(testTemplate)
	c.UserInputs = append(c.UserInputs, tone, topic)
	c = compiled(t, c)
	if got, err := c.UserVars(map[string]any{"topic": "사과"}); err != nil || got["tone"] != "존댓말" || got["topic"] != "사과" {
		t.Fatalf("default and required: %v %v", err, got)
	}
	for name, raw := range map[string]map[string]any{
		"required missing": {},
		"required blank":   {"topic": "  "},
		"not an option":    {"topic": "사과", "tone": "평어"},
	} {
		if _, err := c.UserVars(raw); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
	task := testChain(testTemplate)
	task.Inputs = []Input{{Name: "title", IsRequired: true}}
	if err := task.Compile(); err == nil {
		t.Error("is_required is for user inputs")
	}
}

func TestBuiltinInputs(t *testing.T) {
	if DocEval.Inputs[0].Name != "title" || DocCreate.UserInputs[0].UI.Placeholder == "" {
		t.Fatalf("inputs written as a name and as a mapping must both load: %+v %+v", DocEval.Inputs, DocCreate.UserInputs)
	}
	for _, c := range []Chain{DocCreate, DocUpdate} {
		for _, name := range []string{"texts", "backlinks"} {
			i := slices.IndexFunc(c.UserInputs, func(in Input) bool { return in.Name == name })
			if i < 0 || c.UserInputs[i].Type() != TypeTextList {
				t.Errorf("%s: %s must be an array of strings", c.Name, name)
			}
		}
	}
}

func TestPromptTemplate(t *testing.T) {
	pt := DocCreate.PromptTemplate()
	if pt.Name != "doc-create" || pt.TemplateFormat != "jinja2" || pt.Template != DocCreate.Steps[0].Template {
		t.Fatalf("clone of doc-create: %+v", pt)
	}
	var names []string
	for _, v := range pt.InputVariables {
		names = append(names, v.Name)
	}
	if strings.Join(names, ",") != "title,hint,texts,backlinks" {
		t.Errorf("input_variables = %v", names)
	}
	back, err := ParsePromptTemplate(pt.YAML())
	if err != nil || !reflect.DeepEqual(back, pt) {
		t.Fatalf("YAML round trip: %v\n%s", err, pt.YAML())
	}
	if !strings.Contains(pt.YAML(), "template: |-\n") {
		t.Errorf("the template must be a block: %s", pt.YAML())
	}
	if _, err := ParsePromptTemplate("name: x\nunknown: 1\n"); err == nil {
		t.Error("an unknown field must be rejected")
	}
}

func TestWithTemplate(t *testing.T) {
	tpl := PromptTemplate{
		Name: "mine", TemplateFormat: "jinja2",
		Template:       "<message role=\"system\">\n지침 {{풀기:사전링크}}\n</message>\n<message role=\"user\">\n{{ title }}\n{% if hint %}{{ hint }}{% endif %}\n</message>",
		InputVariables: []InputVariable{{Name: "hint"}, {Name: "title"}},
	}
	c, err := DocCreate.WithTemplate(tpl)
	if err != nil {
		t.Fatal(err)
	}
	if len(c.UserInputs) != 1 || c.UserInputs[0].Name != "hint" {
		t.Errorf("picked user inputs: %+v", c.UserInputs)
	}
	got := c.PromptStep().Render(Vars{"title": "눈", "hint": "snow"})
	if len(got) != 2 || got[0].Content != "지침 {{풀기:사전링크}}" || got[1].Content != "눈\nsnow" {
		t.Errorf("render: %+v", got)
	}
	if DocCreate.PromptStep().Template == tpl.Template {
		t.Error("WithTemplate must not change the built-in chain")
	}
	for name, mutate := range map[string]func(*PromptTemplate){
		"unknown variable":   func(p *PromptTemplate) { p.InputVariables = append(p.InputVariables, InputVariable{Name: "body"}) },
		"repeated variable":  func(p *PromptTemplate) { p.InputVariables = append(p.InputVariables, InputVariable{Name: "hint"}) },
		"undeclared use":     func(p *PromptTemplate) { p.InputVariables = []InputVariable{{Name: "title"}} },
		"title not read":     func(p *PromptTemplate) { p.Template = "<message role=\"user\">\nx\n</message>" },
		"other format":       func(p *PromptTemplate) { p.TemplateFormat = "liquid" },
		"wiki name as a var": func(p *PromptTemplate) { p.Template = "<message role=\"user\">\n{{ title }} {{PAGENAME}}\n</message>" },
	} {
		p := tpl
		p.InputVariables = slices.Clone(tpl.InputVariables)
		mutate(&p)
		if _, err := DocCreate.WithTemplate(p); err == nil {
			t.Errorf("%s: expected an error", name)
		}
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

func TestNextVersion(t *testing.T) {
	for _, c := range []struct {
		v     string
		major bool
		want  string
	}{{"v1.0", false, "v1.1"}, {"v1.9", false, "v1.10"}, {"v1.3", true, "v2.0"}, {"v2", false, "v2.1"}, {"", false, "v0.1"}} {
		if got := NextVersion(c.v, c.major); got != c.want {
			t.Errorf("NextVersion(%q, %v) = %q, want %q", c.v, c.major, got, c.want)
		}
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
