package prompts

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// Chain is a pipeline of steps, following LangChain. Each step's output is saved under its Output name and
// can be an input of later steps. Built-in chains are YAML files in chains/, embedded and loaded at start, and
// listed read-only in the AI tool's chains tab (the aiedit prompts are stored in the database and editable).
// Metadata describes the chain; Spec is what decides its results.
type Chain struct {
	Name     string   `json:"name" yaml:"name"`
	Metadata Metadata `json:"metadata" yaml:"metadata"`
	Spec     Spec     `json:"spec" yaml:"spec"`
}

// Metadata.Version is v<major>.<minor>, recorded with each call and result of the chain. A stored result of the
// same major is current (SameMajor): a minor bump keeps earlier results, a major bump makes them all stale.
type Metadata struct {
	Version     string `json:"version" yaml:"version"`
	Description string `json:"description" yaml:"description"`
}

type Spec struct {
	Inputs []string `json:"inputs" yaml:"inputs"`
	Steps  []Step   `json:"steps" yaml:"steps"`
}

// Step is one LLM call of a chain, LangChain's prompt | model | parser: the prompt is rendered, sent to the
// model, and the response is parsed into the value named by Output.
type Step struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description" yaml:"description"`
	Prompt      `yaml:",inline"`
	Parser      Parser `json:"parser" yaml:"parser"`
	Output      string `json:"output" yaml:"output"`
}

// Major returns the major part of a version: "v2" for "v2.3", and for "v2" (written before minors).
func Major(version string) string {
	major, _, _ := strings.Cut(version, ".")
	return major
}

// SameMajor reports whether a result recorded with version a is current for version b.
func SameMajor(a, b string) bool {
	return Major(a) == Major(b)
}

// Step returns the step named name and panics when there is none, so a wrong name fails at start.
func (c Chain) Step(name string) Step {
	for _, st := range c.Spec.Steps {
		if st.Name == name {
			return st
		}
	}
	panic(fmt.Sprintf("chain %s has no step %q", c.Name, name))
}

// Prompt is a chat prompt template. Message contents may hold {name} placeholders for the variables listed in
// Inputs, which are chain inputs or outputs of earlier steps.
type Prompt struct {
	Inputs   []string  `json:"inputs" yaml:"inputs"`
	Messages []Message `json:"messages" yaml:"messages"`
}

type Message struct {
	Role    string `json:"role" yaml:"role"`
	Content string `json:"content" yaml:"content"`
}

// Parser turns a model response into a value.
type Parser string

const (
	ParserText Parser = "text"
	ParserJSON Parser = "json"
)

// Render returns the messages with each {input} placeholder replaced by vars[input]. Values are inserted as
// they are, so placeholders inside a value are not expanded.
func (p Prompt) Render(vars map[string]string) []Message {
	pairs := make([]string, 0, 2*len(p.Inputs))
	for _, in := range p.Inputs {
		pairs = append(pairs, "{"+in+"}", vars[in])
	}
	r := strings.NewReplacer(pairs...)
	out := make([]Message, len(p.Messages))
	for i, m := range p.Messages {
		out[i] = Message{Role: m.Role, Content: r.Replace(m.Content)}
	}
	return out
}

// Parse decodes a model response into out: for ParserJSON a JSON value (a surrounding code fence is stripped)
// into what out points to, for ParserText the trimmed text into a *string.
func (p Parser) Parse(raw string, out any) error {
	switch p {
	case ParserJSON:
		return json.Unmarshal([]byte(stripCodeFence(raw)), out)
	case ParserText:
		s, ok := out.(*string)
		if !ok {
			return fmt.Errorf("text output needs a *string, got %T", out)
		}
		*s = strings.TrimSpace(raw)
		return nil
	}
	return fmt.Errorf("unknown parser %q", string(p))
}

func stripCodeFence(raw string) string {
	s := strings.TrimSpace(raw)
	if !strings.HasPrefix(s, "```") {
		return s
	}
	lines := strings.Split(s, "\n")
	lines = lines[1:]
	if len(lines) > 0 && strings.HasPrefix(lines[len(lines)-1], "```") {
		lines = lines[:len(lines)-1]
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

//go:embed chains/*.yaml
var chainFiles embed.FS

// mustLoad reads chains/<name>.yaml and panics when it is missing or invalid, so a broken file fails at start.
func mustLoad(name string) Chain {
	b, err := chainFiles.ReadFile("chains/" + name + ".yaml")
	if err != nil {
		panic(err)
	}
	var c Chain
	dec := yaml.NewDecoder(bytes.NewReader(b))
	dec.KnownFields(true)
	if err := dec.Decode(&c); err != nil {
		panic(fmt.Sprintf("chains/%s.yaml: %v", name, err))
	}
	if c.Name != name {
		panic(fmt.Sprintf("chains/%s.yaml: name is %q", name, c.Name))
	}
	return c
}

var (
	DocDefine = mustLoad("doc-define")
	DocEval   = mustLoad("doc-eval")
	DocCreate = mustLoad("doc-create")
	DocUpdate = mustLoad("doc-update")
)

func BuiltinChains() []Chain {
	return []Chain{DocDefine, DocEval, DocCreate, DocUpdate}
}
