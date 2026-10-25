package prompts

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"strconv"
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

// Spec.UserInputs are what the requester of a run gives (the run form); they reach the messages as they are.
// Spec.Inputs are what the task running the chain fills (a page's title, its wikitext). Together they are the
// chain's variables.
type Spec struct {
	UserInputs []Input `json:"userInputs,omitempty" yaml:"userInputs,omitempty"`
	Inputs     []Input `json:"inputs" yaml:"inputs"`
	Steps      []Step  `json:"steps" yaml:"steps"`
}

// Input is a variable of a chain. Label names a user input in the run form, and Description guides it there.
type Input struct {
	Name        string    `json:"name" yaml:"name"`
	Label       string    `json:"label,omitempty" yaml:"label,omitempty"`
	Description string    `json:"description,omitempty" yaml:"description,omitempty"`
	Type        InputType `json:"type,omitempty" yaml:"type,omitempty"`
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

// InputType is the type of a variable's value: string (the default, "" in a file), bool, or texts ([]string).
type InputType string

const (
	TypeString InputType = "string"
	TypeBool   InputType = "bool"
	TypeTexts  InputType = "texts"
)

func (t InputType) orString() InputType {
	if t == "" {
		return TypeString
	}
	return t
}

// Vars are the values of a chain's variables for one run: a string, a bool, or a []string per name.
type Vars map[string]any

// present reports whether v counts as given for a message's If: a non-blank string, true, or a non-empty list.
func present(v any) bool {
	switch v := v.(type) {
	case string:
		return strings.TrimSpace(v) != ""
	case bool:
		return v
	case []string:
		return len(v) > 0
	}
	return false
}

func text(v any) string {
	switch v := v.(type) {
	case string:
		return v
	case bool:
		return strconv.FormatBool(v)
	case []string:
		return strings.Join(v, "\n\n")
	}
	return ""
}

// UserVars turns the run form's values into Vars, checking each against its user input's type: a string, a
// bool, or a list of strings whose blank items are dropped. Missing user inputs are left out.
func (c Chain) UserVars(raw map[string]any) (Vars, error) {
	vars := Vars{}
	for name := range raw {
		if !slices.ContainsFunc(c.Spec.UserInputs, func(in Input) bool { return in.Name == name }) {
			return nil, fmt.Errorf("unknown user input %q", name)
		}
	}
	for _, in := range c.Spec.UserInputs {
		v, ok := raw[in.Name]
		if !ok || v == nil {
			continue
		}
		switch in.Type.orString() {
		case TypeString:
			s, ok := v.(string)
			if !ok {
				return nil, fmt.Errorf("user input %s must be a string", in.Name)
			}
			vars[in.Name] = s
		case TypeBool:
			b, ok := v.(bool)
			if !ok {
				return nil, fmt.Errorf("user input %s must be true or false", in.Name)
			}
			vars[in.Name] = b
		case TypeTexts:
			items, ok := v.([]any)
			if !ok {
				return nil, fmt.Errorf("user input %s must be a list of texts", in.Name)
			}
			texts := []string{}
			for _, it := range items {
				s, ok := it.(string)
				if !ok {
					return nil, fmt.Errorf("user input %s must be a list of texts", in.Name)
				}
				if strings.TrimSpace(s) != "" {
					texts = append(texts, s)
				}
			}
			vars[in.Name] = texts
		}
	}
	return vars, nil
}

// Step is one LLM call of a chain, LangChain's prompt | model | parser: the prompt is rendered, sent to the
// model, and the response is read in the step's Format. Output names the parsed value when a later step reads it (like ADK's
// output_key); a step whose value only goes back to the service needs none.
type Step struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description" yaml:"description"`
	Prompt      `yaml:",inline"`
	Format      Format `json:"format,omitempty" yaml:"format,omitempty"`
	Output      string `json:"output,omitempty" yaml:"output,omitempty"`
}

// Major returns the major part of a version: "v2" for "v2.3", and for "v2" (written before minors).
func Major(version string) string {
	major, _, _ := strings.Cut(version, ".")
	return major
}

// NextVersion bumps a v<major>.<minor> version: the minor, or with major the major (minor back to 0). A version
// without a minor (v2) reads as v2.0.
func NextVersion(version string, major bool) string {
	m, n, _ := strings.Cut(strings.TrimPrefix(version, "v"), ".")
	mi, _ := strconv.Atoi(m)
	ni, _ := strconv.Atoi(n)
	if major {
		return fmt.Sprintf("v%d.0", mi+1)
	}
	return fmt.Sprintf("v%d.%d", mi, ni+1)
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
// Inputs, which are the chain's variables or outputs of earlier steps.
type Prompt struct {
	Inputs   []string  `json:"inputs" yaml:"inputs"`
	Messages []Message `json:"messages" yaml:"messages"`
}

// Message is one message of a prompt. If sends it only when that variable is given (present). Each sends it once
// per item of a texts variable, with {name} as the item and {index} as its number from 1; an empty list sends
// none. Rendered messages carry neither.
type Message struct {
	Role    string `json:"role" yaml:"role"`
	Content string `json:"content" yaml:"content"`
	If      string `json:"if,omitempty" yaml:"if,omitempty"`
	Each    string `json:"each,omitempty" yaml:"each,omitempty"`
}

// Format is what a step asks the model to return, like an API's response_format: text (the default, "" in a
// file) or json. Parse reads a response in it.
type Format string

const (
	FormatText Format = "text"
	FormatJSON Format = "json"
)

// Render returns the messages to send, following each message's If and Each, with each {input} placeholder
// replaced by vars[input]. Values are inserted as they are, so placeholders inside a value are not expanded.
func (p Prompt) Render(vars Vars) []Message {
	var out []Message
	for _, m := range p.Messages {
		if m.If != "" && !present(vars[m.If]) {
			continue
		}
		if m.Each == "" {
			out = append(out, Message{Role: m.Role, Content: p.replace(m.Content, vars, nil)})
			continue
		}
		items, _ := vars[m.Each].([]string)
		for i, item := range items {
			own := map[string]string{m.Each: item, "index": strconv.Itoa(i + 1)}
			out = append(out, Message{Role: m.Role, Content: p.replace(m.Content, vars, own)})
		}
	}
	return out
}

func (p Prompt) replace(content string, vars Vars, own map[string]string) string {
	pairs := make([]string, 0, 2*(len(p.Inputs)+1))
	for _, in := range p.Inputs {
		v, ok := own[in]
		if !ok {
			v = text(vars[in])
		}
		pairs = append(pairs, "{"+in+"}", v)
	}
	if idx, ok := own["index"]; ok {
		pairs = append(pairs, "{index}", idx)
	}
	return strings.NewReplacer(pairs...).Replace(content)
}

// Parse reads a model response, with a code fence around the whole of it stripped: for FormatJSON a JSON value
// into what out points to, for FormatText the trimmed text into a *string.
func (f Format) Parse(raw string, out any) error {
	switch f {
	case FormatJSON:
		return json.Unmarshal([]byte(stripCodeFence(raw)), out)
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
	if err := c.Validate(); err != nil {
		panic(fmt.Sprintf("chains/%s.yaml: %v", name, err))
	}
	return c
}

var (
	nameRe        = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]*$`)
	placeholderRe = regexp.MustCompile(`\{([A-Za-z][A-Za-z0-9_]*)\}`)
)

// Validate checks a chain's structure: variable and output names and types; each step's inputs are variables
// or earlier outputs and each is used by its messages; placeholders, If and Each name the step's inputs, Each a texts
// variable, and a texts variable appears as a placeholder only in its own Each message; every task input is used.
func (c Chain) Validate() error {
	types := map[string]InputType{}
	for _, in := range slices.Concat(c.Spec.UserInputs, c.Spec.Inputs) {
		if !nameRe.MatchString(in.Name) || in.Name == "index" {
			return fmt.Errorf("invalid variable name %q", in.Name)
		}
		if _, dup := types[in.Name]; dup {
			return fmt.Errorf("duplicate variable %q", in.Name)
		}
		t := in.Type.orString()
		if t != TypeString && t != TypeBool && t != TypeTexts {
			return fmt.Errorf("variable %s: unknown type %q", in.Name, in.Type)
		}
		types[in.Name] = t
	}
	if len(c.Spec.Steps) == 0 {
		return fmt.Errorf("no steps")
	}
	used := map[string]bool{}
	for _, st := range c.Spec.Steps {
		if err := st.validate(types, used, true); err != nil {
			return fmt.Errorf("step %s: %w", st.Name, err)
		}
		if st.Output == "" {
			continue
		}
		if _, dup := types[st.Output]; dup || !nameRe.MatchString(st.Output) {
			return fmt.Errorf("step %s: invalid or duplicate output %q", st.Name, st.Output)
		}
		types[st.Output] = TypeString
	}
	for _, in := range c.Spec.Inputs {
		if !used[in.Name] {
			return fmt.Errorf("input %q is not used", in.Name)
		}
	}
	return nil
}

// ValidateMessages checks messages written to replace those of the chain's first step (a custom template): the
// same rules as Validate, except that they need not use every input of the step.
func (c Chain) ValidateMessages(messages []Message) error {
	types := c.variableTypes()
	st := c.Spec.Steps[0]
	st.Messages = messages
	return st.validate(types, map[string]bool{}, false)
}

// WithUserInputs returns the chain with inputs as its user inputs (a custom template's run form), checked like
// Validate checks them and not clashing with the task's inputs. The first step may then read every user input,
// besides the task inputs it read.
func (c Chain) WithUserInputs(inputs []Input) (Chain, error) {
	seen := map[string]bool{}
	for _, in := range c.Spec.Inputs {
		seen[in.Name] = true
	}
	for _, in := range inputs {
		if !nameRe.MatchString(in.Name) || in.Name == "index" {
			return c, fmt.Errorf("invalid user input name %q", in.Name)
		}
		if seen[in.Name] {
			return c, fmt.Errorf("user input %q is already a variable", in.Name)
		}
		seen[in.Name] = true
		if t := in.Type.orString(); t != TypeString && t != TypeBool && t != TypeTexts {
			return c, fmt.Errorf("user input %s: unknown type %q", in.Name, in.Type)
		}
	}
	steps := slices.Clone(c.Spec.Steps)
	var stepInputs []string
	for _, name := range steps[0].Inputs {
		if slices.ContainsFunc(c.Spec.Inputs, func(in Input) bool { return in.Name == name }) {
			stepInputs = append(stepInputs, name)
		}
	}
	for _, in := range inputs {
		stepInputs = append(stepInputs, in.Name)
	}
	steps[0].Inputs = stepInputs
	c.Spec.UserInputs = slices.Clone(inputs)
	c.Spec.Steps = steps
	return c, nil
}

// Template is the part of a chain a custom template rewrites, as written in YAML: the run form and the first step's
// messages.
type Template struct {
	UserInputs []Input   `json:"userInputs" yaml:"userInputs"`
	Messages   []Message `json:"messages" yaml:"messages"`
}

// Template returns the chain's own run form and first step's messages.
func (c Chain) Template() Template {
	return Template{UserInputs: c.Spec.UserInputs, Messages: c.Spec.Steps[0].Messages}
}

// YAML writes the template as the chain files do (multi-line contents as blocks).
func (t Template) YAML() string {
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(t); err != nil {
		return ""
	}
	return buf.String()
}

// ParseTemplate reads a template written in YAML, rejecting unknown fields. It does not check it against a chain
// (WithUserInputs, ValidateMessages).
func ParseTemplate(text string) (Template, error) {
	var t Template
	dec := yaml.NewDecoder(strings.NewReader(text))
	dec.KnownFields(true)
	if err := dec.Decode(&t); err != nil {
		return t, fmt.Errorf("YAML: %w", err)
	}
	return t, nil
}

func (c Chain) variableTypes() map[string]InputType {
	types := map[string]InputType{}
	for _, in := range slices.Concat(c.Spec.UserInputs, c.Spec.Inputs) {
		types[in.Name] = in.Type.orString()
	}
	return types
}

// validate checks a step; strict also requires every input of the step to be used by its messages.
func (st Step) validate(types map[string]InputType, used map[string]bool, strict bool) error {
	if st.Name == "" || len(st.Messages) == 0 {
		return fmt.Errorf("name and messages are required")
	}
	if f := st.Format.OrText(); f != FormatJSON && f != FormatText {
		return fmt.Errorf("unknown format %q", st.Format)
	}
	for _, in := range st.Inputs {
		if _, ok := types[in]; !ok {
			return fmt.Errorf("input %q is neither a variable nor an earlier output", in)
		}
	}
	stepUsed := map[string]bool{}
	for _, m := range st.Messages {
		if m.Content == "" || (m.Role != "system" && m.Role != "user") {
			return fmt.Errorf("invalid %q message", m.Role)
		}
		for _, name := range []string{m.If, m.Each} {
			if name != "" && !slices.Contains(st.Inputs, name) {
				return fmt.Errorf("%q is not an input of the step", name)
			}
			stepUsed[name] = true
		}
		if m.Each != "" && types[m.Each] != TypeTexts {
			return fmt.Errorf("each %q is not a texts variable", m.Each)
		}
		for _, sub := range placeholderRe.FindAllStringSubmatch(m.Content, -1) {
			name := sub[1]
			if name == "index" && m.Each != "" {
				continue
			}
			if !slices.Contains(st.Inputs, name) {
				return fmt.Errorf("placeholder {%s} is not an input of the step", name)
			}
			if types[name] == TypeTexts && name != m.Each {
				return fmt.Errorf("texts variable {%s} is used outside its each message", name)
			}
			stepUsed[name] = true
		}
	}
	for _, in := range st.Inputs {
		if strict && !stepUsed[in] {
			return fmt.Errorf("input %q is not used by its messages", in)
		}
		used[in] = true
	}
	return nil
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
