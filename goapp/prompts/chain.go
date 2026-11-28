package prompts

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/zetaoss/zengine/goapp/prompts/jinja"
	"gopkg.in/yaml.v3"
)

// Chain is a pipeline of steps, following LangChain. Each step is a Semantic Kernel style prompt (a jinja2
// template per message); a step's output is saved under its output variable's name and can be read by later
// steps. Built-in chains are YAML files in chains/, embedded and loaded at start, and listed read-only in the AI
// tool's chains tab.
//
// Version is v<major>.<minor>, recorded with each call and result of the chain. A stored result of the same
// major is current (SameMajor): a minor bump keeps earlier results, a major bump makes them all stale. Inputs are
// what the task running the chain fills (a page's title, its wikitext): the chain's variables.
type Chain struct {
	Name        string  `json:"name" yaml:"name"`
	Description string  `json:"description" yaml:"description"`
	Version     string  `json:"version" yaml:"version"`
	Inputs      []Input `json:"inputs" yaml:"inputs"`
	Steps       []Step  `json:"steps" yaml:"steps"`
}

// Input is a variable of a chain, a string the task running it fills, written as its name alone or with a
// description.
type Input struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

// UnmarshalYAML also reads an input written as its name alone.
func (in *Input) UnmarshalYAML(n *yaml.Node) error {
	if n.Kind == yaml.ScalarNode {
		in.Name = n.Value
		return nil
	}
	type plain Input
	return n.Decode((*plain)(in))
}

// Vars are the values of a chain's variables for one run, by name.
type Vars map[string]any

// Step is one LLM call of a chain, LangChain's prompt | model | parser: its Message, a jinja2 template per role
// (ParseChat), is rendered and sent to the model, and the response is read in its output variable's Format. The
// output's Name is given only when a later step reads it (like ADK's output_key). Inputs are the variables the
// messages read, found when the chain loads, like LangChain's input_variables.
type Step struct {
	Name           string           `json:"name" yaml:"name"`
	Description    string           `json:"description,omitempty" yaml:"description,omitempty"`
	Message        MessageTemplates `json:"message" yaml:"message"`
	OutputVariable *OutputVariable  `json:"output_variable,omitempty" yaml:"output_variable,omitempty"`
	Inputs         []string         `json:"inputs" yaml:"-"`
	chat           ChatTemplate
}

// OutputVariable is what a step returns: Format, and Name when a later step reads it.
type OutputVariable struct {
	Name   string `json:"name,omitempty" yaml:"name,omitempty"`
	Format Format `json:"format,omitempty" yaml:"format,omitempty"`
}

// Output is the name the step's value is saved under, or "".
func (st Step) Output() string {
	if st.OutputVariable == nil {
		return ""
	}
	return st.OutputVariable.Name
}

// Format is what the step asks the model to return.
func (st Step) Format() Format {
	if st.OutputVariable == nil {
		return FormatText
	}
	return st.OutputVariable.Format.OrText()
}

// Render returns the messages to send. A variable not in vars is empty. Values are inserted as they are.
func (st Step) Render(vars Vars) []Message {
	messages, _ := st.chat.Render(vars)
	return messages
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
	for _, st := range c.Steps {
		if st.Name == name {
			return st
		}
	}
	panic(fmt.Sprintf("chain %s has no step %q", c.Name, name))
}

// Message is one rendered message.
type Message struct {
	Role    string `json:"role" yaml:"role"`
	Content string `json:"content" yaml:"content"`
}

// Format is what a step asks the model to return, like an API's response_format: text (the default, "" in a
// file) or json. Parse reads a response in it.
type Format string

const (
	FormatText Format = "text"
	FormatJSON Format = "json"
)

// Parse reads a model response, with a code fence around the whole of it stripped: for FormatJSON a JSON value
// into what out points to, for FormatText the trimmed text into a *string.
func (f Format) Parse(raw string, out any) error {
	switch f {
	case FormatJSON:
		stripped := stripCodeFence(raw)
		if err := json.Unmarshal([]byte(stripped), out); err == nil {
			return nil
		}
		extracted := extractJSON(raw)
		if extracted != stripped {
			if err := json.Unmarshal([]byte(extracted), out); err == nil {
				return nil
			}
		}
		return json.Unmarshal([]byte(extracted), out)
	case FormatText, "":
		s, ok := out.(*string)
		if !ok {
			return fmt.Errorf("text output needs a *string, got %T", out)
		}
		*s = stripCodeFence(raw)
		return nil
	}
	return fmt.Errorf("unknown format %q", string(f))
}

// OrText is the format, text when none is set.
func (f Format) OrText() Format {
	if f == "" {
		return FormatText
	}
	return f
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

// stripOuterReasoning removes reasoning/thinking blocks only when they occur outside of JSON string literals.
func stripOuterReasoning(raw string) string {
	var sb strings.Builder
	sb.Grow(len(raw))
	inQuote := false
	escape := false

	for i := 0; i < len(raw); {
		c := raw[i]
		if escape {
			sb.WriteByte(c)
			escape = false
			i++
			continue
		}
		if c == '\\' {
			sb.WriteByte(c)
			escape = true
			i++
			continue
		}
		if c == '"' {
			inQuote = !inQuote
			sb.WriteByte(c)
			i++
			continue
		}
		if !inQuote {
			rem := raw[i:]
			if strings.HasPrefix(rem, "<think>") {
				if endIdx := strings.Index(rem, "</think>"); endIdx != -1 {
					i += endIdx + len("</think>")
					continue
				}
			}
			if strings.HasPrefix(rem, "<|thought|>") {
				if endIdx := strings.Index(rem, "<|end_of_thought|>"); endIdx != -1 {
					i += endIdx + len("<|end_of_thought|>")
					continue
				}
			}
			if strings.HasPrefix(rem, "<|channel|>thought") {
				afterTag := rem[len("<|channel|>thought"):]
				if endIdx := strings.Index(afterTag, "<|channel|>"); endIdx != -1 {
					i += len("<|channel|>thought") + endIdx + len("<|channel|>")
					continue
				}
			}
		}
		sb.WriteByte(c)
		i++
	}
	return sb.String()
}

// extractJSON extracts a JSON object or array from raw text, stripping commentary, markdown fences,
// reasoning blocks, normalizing bare backticks around values, and escaping unescaped newlines in strings.
func extractJSON(raw string) string {
	raw = stripOuterReasoning(raw)

	start := -1
	var openChar, closeChar rune

	// Find the first top-level '{' or '['
	for i, r := range raw {
		if r == '{' || r == '[' {
			start = i
			openChar = r
			if r == '{' {
				closeChar = '}'
			} else {
				closeChar = ']'
			}
			break
		}
	}
	if start == -1 {
		return raw
	}

	var sb strings.Builder
	depth := 0
	inQuote := false
	inBacktick := false
	escape := false
	runes := []rune(raw[start:])
	matched := false

	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if escape {
			sb.WriteRune(r)
			escape = false
			continue
		}
		if r == '\\' {
			sb.WriteRune(r)
			escape = true
			continue
		}
		if !inBacktick && r == '"' {
			inQuote = !inQuote
			sb.WriteRune(r)
			continue
		}
		if inQuote {
			switch r {
			case '\n':
				sb.WriteString("\\n")
			case '\r':
				// omit bare carriage return
			case '\t':
				sb.WriteString("\\t")
			default:
				sb.WriteRune(r)
			}
			continue
		}

		// Outside of regular double quotes
		if r == '`' {
			// Check if it is a triple backtick (code fence inside or closing)
			if i+2 < len(runes) && runes[i+1] == '`' && runes[i+2] == '`' {
				i += 2
				continue
			}
			// Single backtick outside quotes: normalize as double quote
			inBacktick = !inBacktick
			sb.WriteRune('"')
			continue
		}

		if !inBacktick {
			if r == openChar {
				depth++
			} else if r == closeChar {
				depth--
				if depth == 0 {
					sb.WriteRune(r)
					matched = true
					break
				}
			}
		}
		sb.WriteRune(r)
	}

	if matched {
		return sb.String()
	}
	return raw
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
	if err := c.Compile(); err != nil {
		panic(fmt.Sprintf("chains/%s.yaml: %v", name, err))
	}
	return c
}

var (
	nameRe    = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]*$`)
	versionRe = regexp.MustCompile(`^v[0-9]+\.[0-9]+$`)
)

// Compile checks the chain and parses its steps' messages. Variable names are distinct and start with a
// letter; each step's messages read only the chain's variables and earlier outputs, which are all strings, so
// none is looped over; every input is read by some step.
func (c *Chain) Compile() error {
	if !versionRe.MatchString(c.Version) {
		return fmt.Errorf("version %q is not v<major>.<minor>", c.Version)
	}
	var names []string
	for _, in := range c.Inputs {
		if !nameRe.MatchString(in.Name) || in.Name == "loop" {
			return fmt.Errorf("invalid variable name %q", in.Name)
		}
		if slices.Contains(names, in.Name) {
			return fmt.Errorf("duplicate variable %q", in.Name)
		}
		names = append(names, in.Name)
	}
	if len(c.Steps) == 0 {
		return fmt.Errorf("no steps")
	}
	c.Steps = slices.Clone(c.Steps)
	used := map[string]bool{}
	for i := range c.Steps {
		st := &c.Steps[i]
		if err := st.compile(names, used); err != nil {
			return fmt.Errorf("step %s: %w", st.Name, err)
		}
		if out := st.Output(); out != "" {
			if slices.Contains(names, out) || !nameRe.MatchString(out) || out == "loop" {
				return fmt.Errorf("step %s: invalid or duplicate output %q", st.Name, out)
			}
			names = append(names, out)
		}
	}
	for _, in := range c.Inputs {
		if !used[in.Name] {
			return fmt.Errorf("input %q is not used", in.Name)
		}
	}
	return nil
}

func (st *Step) compile(names []string, used map[string]bool) error {
	if st.Name == "" {
		return fmt.Errorf("a step needs a name")
	}
	if st.OutputVariable != nil {
		if f := st.OutputVariable.Format.OrText(); f != FormatJSON && f != FormatText {
			return fmt.Errorf("unknown format %q", st.OutputVariable.Format)
		}
	}
	chat, err := ParseChat(st.Message, names)
	if err != nil {
		return err
	}
	for _, u := range chat.Uses() {
		if u.Kind == jinja.UseLoop {
			return fmt.Errorf("for loops over %s, which is not an array", u.Name)
		}
		used[u.Name] = true
	}
	st.chat = chat
	st.Inputs = []string{}
	for _, u := range chat.Uses() {
		if !slices.Contains(st.Inputs, u.Name) {
			st.Inputs = append(st.Inputs, u.Name)
		}
	}
	return nil
}

var (
	DocDefine = mustLoad("doc-define")
	DocEval   = mustLoad("doc-eval")
)

func BuiltinChains() []Chain {
	return []Chain{DocDefine, DocEval}
}
