package prompts

import (
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
		UserInputs: []Input{{Name: "hint"}, {Name: "flag", Type: TypeBool}, {Name: "notes", Type: TypeTexts}},
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
		"texts outside each":     testChain(Message{Role: "user", Content: "{title} {hint} {notes}"}),
		"each on a string":       testChain(Message{Role: "user", Each: "hint", Content: "{title} {notes}"}),
		"index outside each":     testChain(Message{Role: "user", Content: "{title} {index}"}, Message{Role: "user", If: "hint", Each: "notes", Content: "{notes}"}),
		"unused step input":      testChain(Message{Role: "user", Content: "{title}"}),
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
	if !strings.Contains(DocEval.Spec.Inputs[0].Name, "title") || DocCreate.Spec.UserInputs[0].Label == "" {
		t.Fatalf("inputs written as a name and as a mapping must both load: %+v %+v", DocEval.Spec.Inputs, DocCreate.Spec.UserInputs)
	}
}
