package prompts

import (
	"context"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

var versionRe = regexp.MustCompile(`^v[0-9]+\.[0-9]+$`)

func TestBuiltinChains(t *testing.T) {
	seen := map[string]bool{}
	for _, c := range BuiltinChains() {
		if c.Name == "" || seen[c.Name] {
			t.Fatalf("invalid or duplicate chain: %q", c.Name)
		}
		seen[c.Name] = true
		if !versionRe.MatchString(c.Metadata.Version) {
			t.Errorf("%s: version %q is not v<major>.<minor>", c.Name, c.Metadata.Version)
		}
		if err := c.Validate(); err != nil {
			t.Errorf("%s: %v", c.Name, err)
		}
	}
}

func testChain(messages ...Message) Chain {
	return Chain{Name: "t", Spec: Spec{
		UserInputs: []Input{{Name: "hint"}, {Name: "flag", Type: TypeBool}, {Name: "notes", Type: TypeTextList}},
		Inputs:     []Input{{Name: "title"}},
		Steps:      []Step{{Name: "s", Prompt: Prompt{Inputs: []string{"title", "hint", "notes"}, Messages: messages}, Output: "out"}},
	}}
}

func TestValidate(t *testing.T) {
	ok := testChain(
		Message{Role: "user", Content: "{title}"},
		Message{Role: "user", If: "hint", Content: "{hint}"},
		Message{Role: "user", Each: "notes", Content: "{index}: {notes}"},
	)
	if err := ok.Validate(); err != nil {
		t.Fatalf("valid chain rejected: %v", err)
	}
	for name, c := range map[string]Chain{
		"undeclared placeholder": testChain(Message{Role: "user", Content: "{title} {hint} {notes} {other}"}),
		"textList outside each":  testChain(Message{Role: "user", Content: "{title} {hint} {notes}"}),
		"each on a string":       testChain(Message{Role: "user", Each: "hint", Content: "{title} {notes}"}),
		"index outside each":     testChain(Message{Role: "user", Content: "{title} {index}"}, Message{Role: "user", If: "hint", Each: "notes", Content: "{notes}"}),
		"unknown role":           testChain(Message{Role: "assistant", Content: "{title} {hint}"}, Message{Role: "user", Each: "notes", Content: "{notes}"}),
	} {
		if err := c.Validate(); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
	unused := ok
	unused.Spec.Inputs = append(unused.Spec.Inputs, Input{Name: "wikitext"})
	if err := unused.Validate(); err == nil {
		t.Error("an unused task input must be rejected")
	}
	dup := ok
	dup.Spec.Inputs = []Input{{Name: "title"}, {Name: "hint"}}
	if err := dup.Validate(); err == nil {
		t.Error("a duplicate variable must be rejected")
	}
}

func TestRender(t *testing.T) {
	p := Prompt{
		Inputs: []string{"title", "body", "hint", "notes"},
		Messages: []Message{
			{Role: "system", Content: `fixed {"id": 1} {{틀}} {| class="wikitable"`},
			{Role: "user", Content: "{title}: {body}"},
			{Role: "user", If: "hint", Content: "hint {hint}"},
			{Role: "user", Each: "notes", Content: "{index}. {notes} ({title})"},
		},
	}
	got := p.Render(Vars{"title": "T", "body": "{title} {{틀}}", "hint": " ", "notes": []string{"a", "b {index}"}})
	want := []Message{
		{Role: "system", Content: `fixed {"id": 1} {{틀}} {| class="wikitable"`},
		{Role: "user", Content: "T: {title} {{틀}}"},
		{Role: "user", Content: "1. a (T)"},
		{Role: "user", Content: "2. b {index} (T)"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected render:\n%+v\nwant\n%+v", got, want)
	}
	if got := p.Render(Vars{"title": "T", "hint": "x"}); len(got) != 3 || got[2].Content != "hint x" {
		t.Fatalf("a given hint and no notes: %+v", got)
	}
}

func TestUserVars(t *testing.T) {
	c := testChain()
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

func TestInputYAML(t *testing.T) {
	if !strings.Contains(DocEval.Spec.Inputs[0].Name, "title") || DocCreate.Spec.UserInputs[0].Placeholder == "" {
		t.Fatalf("inputs written as a name and as a mapping must both load: %+v %+v", DocEval.Spec.Inputs, DocCreate.Spec.UserInputs)
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

func TestTemplateYAML(t *testing.T) {
	tpl := DocCreate.Template()
	text := tpl.YAML()
	if !strings.Contains(text, "userInputs:") || !strings.Contains(text, "content: |-") {
		t.Fatalf("template YAML should read like the chain files:\n%s", text)
	}
	back, err := ParseTemplate(text)
	if err != nil || !reflect.DeepEqual(back, tpl) {
		t.Fatalf("round trip: %v\n%+v", err, back)
	}
	if _, err := ParseTemplate("messages: []\nsteps: []\n"); err == nil {
		t.Error("unknown fields must be rejected")
	}
}

func TestWikiBraces(t *testing.T) {
	p := Prompt{Inputs: []string{"title"}, Messages: []Message{{Role: "user", Content: "{title} {{title}} {{Infobox}} {{{1}}} {{틀|{title}}}"}}}
	got := p.Render(Vars{"title": "Apple"})[0].Content
	if got != "Apple {{title}} {{Infobox}} {{{1}}} {{틀|Apple}}" {
		t.Fatalf("wiki braces must stay: %q", got)
	}
	c := testChain(
		Message{Role: "user", Content: "{title} {{Infobox}} {{{1}}} {{other}}"},
		Message{Role: "user", If: "hint", Content: "{hint}"},
		Message{Role: "user", Each: "notes", Content: "{notes}"},
	)
	if err := c.Validate(); err != nil {
		t.Fatalf("wiki markup is not a placeholder: %v", err)
	}
}

func TestMagicWords(t *testing.T) {
	c := testChain(
		Message{Role: "user", Content: "{title} 끝에 {{reflist}}, 연도는 {{CURRENTYEAR}}"},
		Message{Role: "user", If: "hint", Content: "{hint}"},
		Message{Role: "user", Each: "notes", Content: "{notes}"},
	)
	if err := c.Validate(); err != nil {
		t.Fatalf("parameterless templates and magic words are wiki markup: %v", err)
	}
}

func TestFill(t *testing.T) {
	long := strings.Repeat("가", 1500)
	if got := DocDefine.Fill(Vars{"wikitext": long})["lead"].(string); len([]rune(got)) != 1000 {
		t.Fatalf("lead is cut to 1000 characters, got %d", len([]rune(got)))
	}
	if got := DocEval.Fill(Vars{"wikitext": "짧은 본문"})["body"]; got != "짧은 본문" {
		t.Fatalf("a short body is kept whole, got %q", got)
	}
	bad := Chain{Name: "t", Spec: Spec{Inputs: []Input{{Name: "lead", From: "missing"}}, Steps: []Step{{Name: "s", Prompt: Prompt{Messages: []Message{{Role: "user", Content: "{lead}"}}}}}}}
	if err := bad.Validate(); err == nil {
		t.Error("from must name a variable")
	}
	cut := Chain{Name: "t", Spec: Spec{Inputs: []Input{{Name: "text"}, {Name: "lead", From: "text", MaxChars: 3}}, Steps: []Step{{Name: "s", Prompt: Prompt{Messages: []Message{{Role: "user", Content: "{lead}"}}}}}}}
	if err := cut.Validate(); err != nil {
		t.Errorf("a source used only through its cut counts as used: %v", err)
	}
}

func functionChain(fn Step) Chain {
	return Chain{Name: "t", Spec: Spec{
		UserInputs: []Input{{Name: "flag", Type: TypeBool}},
		Inputs:     []Input{{Name: "title"}},
		Steps: []Step{fn, {Name: "s", Prompt: Prompt{Messages: []Message{
			{Role: "user", Content: "{title}"},
			{Role: "user", Each: "links", Content: "{index}: {links}"},
		}}}},
	}}
}

func TestFunctionSteps(t *testing.T) {
	ok := functionChain(Step{Name: "f", Uses: "wiki/backlinkTexts", If: "flag", Output: "links"})
	if err := ok.Validate(); err != nil {
		t.Fatalf("valid chain rejected: %v", err)
	}
	if ok.PromptIndex() != 1 || ok.PromptStep().Name != "s" {
		t.Fatalf("prompt step: %d %q", ok.PromptIndex(), ok.PromptStep().Name)
	}
	for name, st := range map[string]Step{
		"unknown function": {Name: "f", Uses: "wiki/nothing", Output: "links"},
		"no output":        {Name: "f", Uses: "wiki/backlinkTexts"},
		"messages":         {Name: "f", Uses: "wiki/backlinkTexts", Output: "links", Prompt: Prompt{Messages: []Message{{Role: "user", Content: "x"}}}},
		"undeclared if":    {Name: "f", Uses: "wiki/backlinkTexts", If: "other", Output: "links"},
		"output clash":     {Name: "f", Uses: "wiki/backlinkTexts", Output: "title"},
	} {
		if err := functionChain(st).Validate(); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
	if err := (Chain{Name: "t", Spec: Spec{Inputs: []Input{{Name: "title"}}, Steps: []Step{{Name: "f", Uses: "wiki/backlinkTexts", Output: "links"}}}}).Validate(); err == nil {
		t.Error("a chain without an LLM step must be rejected")
	}
	if err := ok.ValidateMessages([]Message{{Role: "user", Each: "links", Content: "{links}"}}); err != nil {
		t.Errorf("a template may read an earlier function step's output: %v", err)
	}

	calls := 0
	funcs := map[string]Func{"wiki/backlinkTexts": func(_ context.Context, vars Vars) (any, error) {
		calls++
		return []string{"from " + vars["title"].(string)}, nil
	}}
	got, err := ok.RunFunctions(context.Background(), Vars{"title": "A", "flag": true}, funcs)
	if err != nil || calls != 1 || !reflect.DeepEqual(got["links"], []string{"from A"}) {
		t.Fatalf("run: %v %d %v", err, calls, got)
	}
	got, err = ok.RunFunctions(context.Background(), Vars{"title": "A"}, funcs)
	if err != nil || calls != 1 || !reflect.DeepEqual(got["links"], []string{}) {
		t.Fatalf("skipped run: %v %d %v", err, calls, got)
	}
	if _, err := ok.RunFunctions(context.Background(), Vars{"title": "A", "flag": true}, nil); err == nil {
		t.Error("a function without code must fail")
	}
}

func TestUserInputAttributes(t *testing.T) {
	tone := Input{Name: "tone", Type: TypeChoice, Options: []string{"존댓말", "반말"}, Default: "존댓말"}
	topic := Input{Name: "topic", Required: true}
	c := testChain(Message{Role: "user", Content: "{title} {hint} {tone} {topic}"}, Message{Role: "user", Each: "notes", Content: "{notes}"})
	c.Spec.UserInputs = append(c.Spec.UserInputs, tone, topic)
	if err := c.Validate(); err != nil {
		t.Fatalf("valid chain rejected: %v", err)
	}

	got, err := c.UserVars(map[string]any{"topic": "사과"})
	if err != nil || got["tone"] != "존댓말" || got["topic"] != "사과" {
		t.Fatalf("default and required: %v %v", err, got)
	}
	if got, err := c.UserVars(map[string]any{"topic": "사과", "tone": "반말"}); err != nil || got["tone"] != "반말" {
		t.Fatalf("a picked option: %v %v", err, got)
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

	for name, in := range map[string]Input{
		"choice without options": {Name: "x", Type: TypeChoice},
		"choice with one option": {Name: "x", Type: TypeChoice, Options: []string{"a"}},
		"duplicate options":      {Name: "x", Type: TypeChoice, Options: []string{"a", "a"}},
		"options on a string":    {Name: "x", Options: []string{"a", "b"}},
		"default not an option":  {Name: "x", Type: TypeChoice, Options: []string{"a", "b"}, Default: "c"},
		"string default on bool": {Name: "x", Type: TypeBool, Default: "true"},
		"default on a textList":  {Name: "x", Type: TypeTextList, Default: "a"},
		"required bool":          {Name: "x", Type: TypeBool, Required: true},
		"from on a user input":   {Name: "x", From: "title"},
	} {
		if _, err := c.WithUserInputs([]Input{in}); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
	if _, err := c.WithUserInputs([]Input{{Name: "x", Type: TypeBool, Default: true}, tone}); err != nil {
		t.Errorf("a bool default and a choice are valid: %v", err)
	}
	task := c
	task.Spec.Inputs = []Input{{Name: "title", Required: true}}
	if err := task.Validate(); err == nil {
		t.Error("required is for user inputs")
	}

	var parsed Template
	if parsed, err = ParseTemplate("userInputs:\n  - name: tone\n    type: choice\n    options: [존댓말, 반말]\n    default: 반말\n  - name: short\n    type: bool\n    default: true\nmessages: []\n"); err != nil {
		t.Fatal(err)
	}
	if _, err := c.WithUserInputs(parsed.UserInputs); err != nil || parsed.UserInputs[0].Default != "반말" || parsed.UserInputs[1].Default != true {
		t.Errorf("YAML attributes: %v %+v", err, parsed.UserInputs)
	}
}
