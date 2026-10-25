package prompts

import (
	"regexp"
	"slices"
	"testing"
)

var placeholderRe = regexp.MustCompile(`\{([a-z_]+)\}`)

var versionRe = regexp.MustCompile(`^v[0-9]+\.[0-9]+$`)

func TestBuiltinChains(t *testing.T) {
	seen := map[string]bool{}
	for _, c := range BuiltinChains() {
		if c.Name == "" || c.Metadata.Version == "" || len(c.Spec.Steps) == 0 || seen[c.Name] {
			t.Fatalf("invalid or duplicate chain (name, version and steps are required): %q", c.Name)
		}
		seen[c.Name] = true
		if !versionRe.MatchString(c.Metadata.Version) {
			t.Errorf("%s: version %q is not v<major>.<minor>", c.Name, c.Metadata.Version)
		}
		available := slices.Clone(c.Spec.Inputs)
		for _, st := range c.Spec.Steps {
			if st.Name == "" || st.Output == "" || len(st.Messages) == 0 {
				t.Fatalf("%s/%s: name, output and messages are required", c.Name, st.Name)
			}
			if st.Parser != ParserJSON && st.Parser != ParserText {
				t.Errorf("%s/%s: unknown parser %q", c.Name, st.Name, st.Parser)
			}
			used := map[string]bool{}
			for _, m := range st.Messages {
				if m.Content == "" || (m.Role != "system" && m.Role != "user") {
					t.Fatalf("%s/%s: invalid %q message", c.Name, st.Name, m.Role)
				}
				for _, sub := range placeholderRe.FindAllStringSubmatch(m.Content, -1) {
					if !slices.Contains(st.Inputs, sub[1]) {
						t.Errorf("%s/%s: placeholder {%s} is not an input", c.Name, st.Name, sub[1])
					}
					used[sub[1]] = true
				}
			}
			for _, in := range st.Inputs {
				if !slices.Contains(available, in) {
					t.Errorf("%s/%s: input %q is neither a chain input nor an earlier output", c.Name, st.Name, in)
				}
				if !used[in] {
					t.Errorf("%s/%s: input %q has no placeholder", c.Name, st.Name, in)
				}
			}
			available = append(available, st.Output)
		}
	}
}

func TestRender(t *testing.T) {
	p := Prompt{
		Inputs: []string{"title", "body"},
		Messages: []Message{
			{Role: "system", Content: `fixed {"id": 1}`},
			{Role: "user", Content: "{title}: {body}"},
		},
	}
	got := p.Render(map[string]string{"title": "T", "body": "{title} {{틀}}"})
	if got[0].Content != `fixed {"id": 1}` || got[1].Content != "T: {title} {{틀}}" {
		t.Fatalf("unexpected render: %+v", got)
	}
}

func TestParse(t *testing.T) {
	var v struct {
		A int `json:"a"`
	}
	if err := ParserJSON.Parse("```json\n{\"a\": 1}\n```", &v); err != nil || v.A != 1 {
		t.Fatalf("json in a code fence: %v %+v", err, v)
	}
	if err := ParserJSON.Parse(`{"a": 2}`, &v); err != nil || v.A != 2 {
		t.Fatalf("plain json: %v %+v", err, v)
	}
	if err := ParserJSON.Parse("not json", &v); err == nil {
		t.Fatal("invalid json parsed")
	}
	var s string
	if err := ParserText.Parse("  == 개요 ==\n본문\n", &s); err != nil || s != "== 개요 ==\n본문" {
		t.Fatalf("text: %v %q", err, s)
	}
	if err := ParserText.Parse("x", &v); err == nil {
		t.Fatal("text into a non-string accepted")
	}
	if err := Parser("").Parse("x", &s); err == nil {
		t.Fatal("missing parser accepted")
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
