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

	"github.com/zetaoss/zengine/goapp/prompts/jinja"
	"gopkg.in/yaml.v3"
)

// Chain is a pipeline of steps, following LangChain. Each step is a Semantic Kernel style prompt (a jinja2
// template of messages); a step's output is saved under its output variable's name and can be read by later
// steps. Built-in chains are YAML files in chains/, embedded and loaded at start, and listed read-only in the AI
// tool's chains tab (the AI edit prompt templates are stored in the database and editable).
//
// Version is v<major>.<minor>, recorded with each call and result of the chain. A stored result of the same
// major is current (SameMajor): a minor bump keeps earlier results, a major bump makes them all stale. UserInputs
// are what the requester of a run gives (the run form); Inputs are what the task running the chain fills (a
// page's title, its wikitext). Together they are the chain's variables.
type Chain struct {
	Name        string  `json:"name" yaml:"name"`
	Description string  `json:"description" yaml:"description"`
	Version     string  `json:"version" yaml:"version"`
	UserInputs  []Input `json:"user_inputs,omitempty" yaml:"user_inputs,omitempty"`
	Inputs      []Input `json:"inputs" yaml:"inputs"`
	Steps       []Step  `json:"steps" yaml:"steps"`
}

// Input is a variable of a chain. Its value has the shape of JSONSchema, a closed subset of JSON Schema: a string
// (the default), a string from Enum (a choice), a boolean, or an array of strings. UI tells the run form how to
// show it, with the names of A2UI's TextField: Variant shortText (one line, the default) or longText (several
// lines), and Placeholder in an empty field. A user input may be IsRequired (a run without it, or with it blank,
// is refused; not for a boolean) and have a Default (its value when the run form leaves it out; not for an array).
// A task input is a string, written as its name alone or with a description.
type Input struct {
	Name        string      `json:"name" yaml:"name"`
	Description string      `json:"description,omitempty" yaml:"description,omitempty"`
	JSONSchema  *JSONSchema `json:"json_schema,omitempty" yaml:"json_schema,omitempty"`
	UI          *UI         `json:"ui,omitempty" yaml:"ui,omitempty"`
	IsRequired  bool        `json:"is_required,omitempty" yaml:"is_required,omitempty"`
	Default     any         `json:"default,omitempty" yaml:"default,omitempty"`
}

// JSONSchema is the subset of JSON Schema an input's value may take (see Input).
type JSONSchema struct {
	Type  string      `json:"type" yaml:"type"`
	Items *JSONSchema `json:"items,omitempty" yaml:"items,omitempty"`
	Enum  []string    `json:"enum,omitempty" yaml:"enum,omitempty"`
}

// UI is how the run form shows an input (see Input).
type UI struct {
	Variant     string `json:"variant,omitempty" yaml:"variant,omitempty"`
	Placeholder string `json:"placeholder,omitempty" yaml:"placeholder,omitempty"`
}

// InputType is what an input is, read from its JSONSchema and UI: the field the run form shows and the Go type
// of its value (string, bool or []string).
type InputType string

const (
	TypeString   InputType = "string"
	TypeText     InputType = "text"
	TypeTextList InputType = "textList"
	TypeBool     InputType = "bool"
	TypeChoice   InputType = "choice"
)

const (
	variantShort = "shortText"
	variantLong  = "longText"
)

// Type is the input's InputType.
func (in Input) Type() InputType {
	s := in.JSONSchema
	switch {
	case s == nil || s.Type == "string" && s.Enum == nil:
		if in.UI != nil && in.UI.Variant == variantLong {
			return TypeText
		}
		return TypeString
	case s.Type == "string":
		return TypeChoice
	case s.Type == "boolean":
		return TypeBool
	}
	return TypeTextList
}

// Options are a choice's values.
func (in Input) Options() []string {
	if in.JSONSchema == nil {
		return nil
	}
	return in.JSONSchema.Enum
}

// MarshalJSON adds the input's type, so the run form need not read its schema.
func (in Input) MarshalJSON() ([]byte, error) {
	type plain Input
	return json.Marshal(struct {
		plain
		Type InputType `json:"type"`
	}{plain(in), in.Type()})
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

func (in Input) validateSchema() error {
	if s := in.JSONSchema; s != nil {
		switch {
		case s.Type == "string" && s.Items == nil:
			if s.Enum != nil {
				if len(s.Enum) < 2 {
					return fmt.Errorf("a choice needs two or more enum values")
				}
				seen := map[string]bool{}
				for _, o := range s.Enum {
					if strings.TrimSpace(o) == "" || seen[o] {
						return fmt.Errorf("enum values must be distinct and not blank")
					}
					seen[o] = true
				}
			}
		case s.Type == "boolean" && s.Items == nil && s.Enum == nil:
		case s.Type == "array" && s.Enum == nil && s.Items != nil && s.Items.Type == "string" && s.Items.Items == nil && s.Items.Enum == nil:
		default:
			return fmt.Errorf("json_schema is a string, a string with enum, a boolean, or an array of strings")
		}
	}
	if in.UI != nil {
		t := in.Type()
		if in.UI.Variant != "" && in.UI.Variant != variantShort && in.UI.Variant != variantLong {
			return fmt.Errorf("ui.variant is %s or %s", variantShort, variantLong)
		}
		if t == TypeBool || t == TypeChoice {
			return fmt.Errorf("ui is for strings and arrays of strings")
		}
	}
	return nil
}

// validateUser checks a user input's name, schema and attributes.
func (in Input) validateUser() error {
	if !nameRe.MatchString(in.Name) || in.Name == "loop" {
		return fmt.Errorf("invalid user input name %q", in.Name)
	}
	if err := in.validateSchema(); err != nil {
		return fmt.Errorf("user input %s: %w", in.Name, err)
	}
	t := in.Type()
	if in.IsRequired && t == TypeBool {
		return fmt.Errorf("user input %s: a boolean cannot be required", in.Name)
	}
	if in.Default == nil {
		return nil
	}
	switch d := in.Default.(type) {
	case bool:
		if t == TypeBool {
			return nil
		}
	case string:
		if t == TypeString || t == TypeText || (t == TypeChoice && slices.Contains(in.Options(), d)) {
			return nil
		}
	}
	return fmt.Errorf("user input %s: default %v does not fit a %s", in.Name, in.Default, t)
}

// Vars are the values of a chain's variables for one run: a string, a bool, or a []string per name.
type Vars map[string]any

func present(v any) bool { return jinja.Truthy(v) }

// UserVars turns the run form's values into Vars, checking each against its user input's type: a string (for a
// choice, one of its options), a bool, or a list of strings whose blank items are dropped. A missing user input
// takes its Default or is left out; a required one that is missing or blank is refused.
func (c Chain) UserVars(raw map[string]any) (Vars, error) {
	vars := Vars{}
	for name := range raw {
		if !slices.ContainsFunc(c.UserInputs, func(in Input) bool { return in.Name == name }) {
			return nil, fmt.Errorf("unknown user input %q", name)
		}
	}
	for _, in := range c.UserInputs {
		v, ok := raw[in.Name]
		if !ok || v == nil {
			v = in.Default
		}
		if v == nil {
			if in.IsRequired {
				return nil, fmt.Errorf("user input %s is required", in.Name)
			}
			continue
		}
		switch t := in.Type(); t {
		case TypeString, TypeText, TypeChoice:
			s, ok := v.(string)
			if !ok {
				return nil, fmt.Errorf("user input %s must be a string", in.Name)
			}
			if t == TypeChoice && !slices.Contains(in.Options(), s) {
				return nil, fmt.Errorf("user input %s must be one of %s", in.Name, strings.Join(in.Options(), ", "))
			}
			vars[in.Name] = s
		case TypeBool:
			b, ok := v.(bool)
			if !ok {
				return nil, fmt.Errorf("user input %s must be true or false", in.Name)
			}
			vars[in.Name] = b
		case TypeTextList:
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
		if in.IsRequired && !present(vars[in.Name]) {
			return nil, fmt.Errorf("user input %s is required", in.Name)
		}
	}
	return vars, nil
}

// Step is one LLM call of a chain, LangChain's prompt | model | parser: its Template, a jinja2 template of
// messages (ParseChat), is rendered and sent to the model, and the response is read in its output variable's
// Format. The output's Name is given only when a later step reads it (like ADK's output_key). Inputs are the
// variables the template reads, found when the chain loads, like LangChain's input_variables.
type Step struct {
	Name           string          `json:"name" yaml:"name"`
	Description    string          `json:"description,omitempty" yaml:"description,omitempty"`
	TemplateFormat string          `json:"template_format" yaml:"template_format"`
	Template       string          `json:"template" yaml:"template"`
	OutputVariable *OutputVariable `json:"output_variable,omitempty" yaml:"output_variable,omitempty"`
	Inputs         []string        `json:"inputs" yaml:"-"`
	chat           ChatTemplate
}

// OutputVariable is what a step returns: Format, and Name when a later step reads it.
type OutputVariable struct {
	Name   string `json:"name,omitempty" yaml:"name,omitempty"`
	Format Format `json:"format,omitempty" yaml:"format,omitempty"`
}

// TemplateFormatJinja2 is the only template format: the dialect of package jinja.
const TemplateFormatJinja2 = "jinja2"

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
	for _, st := range c.Steps {
		if st.Name == name {
			return st
		}
	}
	panic(fmt.Sprintf("chain %s has no step %q", c.Name, name))
}

// PromptStep returns the chain's first step: the one AI edit sends and a custom template rewrites.
func (c Chain) PromptStep() Step {
	return c.Steps[0]
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

// Compile checks the chain and parses its steps' templates. Variable names are distinct and start with a
// letter; task inputs are plain strings; each step's template is jinja2, its messages are system or user
// messages and read only the chain's variables and earlier outputs, a for loops over an array; every task input
// is read by some step.
func (c *Chain) Compile() error {
	if !versionRe.MatchString(c.Version) {
		return fmt.Errorf("version %q is not v<major>.<minor>", c.Version)
	}
	for _, in := range c.UserInputs {
		if err := in.validateUser(); err != nil {
			return err
		}
	}
	for _, in := range c.Inputs {
		if in.JSONSchema != nil || in.UI != nil || in.IsRequired || in.Default != nil {
			return fmt.Errorf("input %s: a task input has a name and a description only", in.Name)
		}
	}
	types := map[string]InputType{}
	var names []string
	for _, in := range slices.Concat(c.UserInputs, c.Inputs) {
		if !nameRe.MatchString(in.Name) || in.Name == "loop" {
			return fmt.Errorf("invalid variable name %q", in.Name)
		}
		if _, dup := types[in.Name]; dup {
			return fmt.Errorf("duplicate variable %q", in.Name)
		}
		types[in.Name] = in.Type()
		names = append(names, in.Name)
	}
	if len(c.Steps) == 0 {
		return fmt.Errorf("no steps")
	}
	c.Steps = slices.Clone(c.Steps)
	used := map[string]bool{}
	for i := range c.Steps {
		st := &c.Steps[i]
		if err := st.compile(types, names, used); err != nil {
			return fmt.Errorf("step %s: %w", st.Name, err)
		}
		if out := st.Output(); out != "" {
			if _, dup := types[out]; dup || !nameRe.MatchString(out) || out == "loop" {
				return fmt.Errorf("step %s: invalid or duplicate output %q", st.Name, out)
			}
			types[out] = TypeString
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

func (st *Step) compile(types map[string]InputType, names []string, used map[string]bool) error {
	if st.Name == "" {
		return fmt.Errorf("a step needs a name")
	}
	if st.TemplateFormat != TemplateFormatJinja2 {
		return fmt.Errorf("template_format is %s", TemplateFormatJinja2)
	}
	if st.OutputVariable != nil {
		if f := st.OutputVariable.Format.OrText(); f != FormatJSON && f != FormatText {
			return fmt.Errorf("unknown format %q", st.OutputVariable.Format)
		}
	}
	chat, err := ParseChat(st.Template, names)
	if err != nil {
		return err
	}
	for _, m := range chat.Messages {
		if m.Role != "system" && m.Role != "user" {
			return fmt.Errorf("a %s message; a step has system and user messages", m.Role)
		}
	}
	for _, u := range chat.Uses() {
		if u.Kind == jinja.UseLoop && types[u.Name] != TypeTextList {
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

// PromptTemplate is an AI edit prompt template as written: a Semantic Kernel prompt document without
// execution_settings (a template's model is a setting of its own, outside the document). Template replaces the
// template of its chain's first step; InputVariables pick, by name, the chain's variables it reads (each as the
// chain defines it).
type PromptTemplate struct {
	Name           string          `json:"name" yaml:"name"`
	Description    string          `json:"description,omitempty" yaml:"description,omitempty"`
	TemplateFormat string          `json:"template_format" yaml:"template_format"`
	Template       string          `json:"template" yaml:"template"`
	InputVariables []InputVariable `json:"input_variables" yaml:"input_variables"`
}

// InputVariable is a variable a template reads, by name.
type InputVariable struct {
	Name string `json:"name" yaml:"name"`
}

// ParsePromptTemplate reads a template written in YAML, rejecting unknown fields. It does not check it against a
// chain (WithTemplate).
func ParsePromptTemplate(text string) (PromptTemplate, error) {
	var t PromptTemplate
	dec := yaml.NewDecoder(strings.NewReader(text))
	dec.KnownFields(true)
	if err := dec.Decode(&t); err != nil {
		return t, fmt.Errorf("YAML: %w", err)
	}
	if t.InputVariables == nil {
		t.InputVariables = []InputVariable{}
	}
	return t, nil
}

// YAML writes the template as the chain files do (multi-line strings as blocks).
func (t PromptTemplate) YAML() string {
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(t); err != nil {
		return ""
	}
	return buf.String()
}

// PromptTemplate is the chain as a template to clone: its first step's template, reading the variables it reads,
// with the chain's name and description.
func (c Chain) PromptTemplate() PromptTemplate {
	st := c.PromptStep()
	t := PromptTemplate{Name: c.Name, Description: c.Description, TemplateFormat: st.TemplateFormat, Template: st.Template, InputVariables: []InputVariable{}}
	for _, in := range slices.Concat(c.Inputs, c.UserInputs) {
		if slices.Contains(st.Inputs, in.Name) {
			t.InputVariables = append(t.InputVariables, InputVariable{Name: in.Name})
		}
	}
	return t
}

// WithTemplate returns the chain as the template runs it: its user inputs are those the template picks (in the
// chain's order), and its first step's template is the template's, read with only the variables the template
// picks. A name the chain does not have or a repeated name is refused.
func (c Chain) WithTemplate(t PromptTemplate) (Chain, error) {
	if t.TemplateFormat != TemplateFormatJinja2 {
		return c, fmt.Errorf("template_format is %s", TemplateFormatJinja2)
	}
	picked := map[string]bool{}
	for _, v := range t.InputVariables {
		known := slices.ContainsFunc(slices.Concat(c.UserInputs, c.Inputs), func(in Input) bool { return in.Name == v.Name })
		if !known {
			return c, fmt.Errorf("input_variables: %s is not a variable of %s (it has %s)", v.Name, c.Name, strings.Join(c.variableNames(), ", "))
		}
		if picked[v.Name] {
			return c, fmt.Errorf("input_variables: %s is there twice", v.Name)
		}
		picked[v.Name] = true
	}
	out := c
	out.UserInputs = []Input{}
	for _, in := range c.UserInputs {
		if picked[in.Name] {
			out.UserInputs = append(out.UserInputs, in)
		}
	}
	out.Steps = slices.Clone(c.Steps)
	out.Steps[0].Template = t.Template
	types := map[string]InputType{}
	var names []string
	for _, in := range slices.Concat(c.UserInputs, c.Inputs) {
		if picked[in.Name] {
			types[in.Name] = in.Type()
			names = append(names, in.Name)
		}
	}
	used := map[string]bool{}
	if err := out.Steps[0].compile(types, names, used); err != nil {
		return c, err
	}
	for _, in := range c.Inputs {
		if !used[in.Name] {
			return c, fmt.Errorf("the template does not read %s; list it in input_variables and use it", in.Name)
		}
	}
	return out, nil
}

func (c Chain) variableNames() []string {
	var names []string
	for _, in := range slices.Concat(c.Inputs, c.UserInputs) {
		names = append(names, in.Name)
	}
	return names
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
