package prompts

import (
	"bytes"
	"context"
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

// Input is a variable of a chain. A user input shows in the run form under its name (readBacklinks as "Read
// backlinks"), with Placeholder in its empty field. An input with From is derived from that variable before a step
// renders (Fill), cut to MaxChars characters when set: the task fills the source, the messages read the cut.
//
// A user input may also be Required (a run without it, or with it blank, is refused; not for a bool), have a
// Default (its value when the run form leaves it out: a string, or true or false for a bool; not for a textList),
// and, as a choice, list its Options (the run form shows them to pick one; the default, if any, is one of them),
// like a GitHub Actions workflow_dispatch input.
type Input struct {
	Name        string    `json:"name" yaml:"name"`
	Placeholder string    `json:"placeholder,omitempty" yaml:"placeholder,omitempty"`
	Type        InputType `json:"type,omitempty" yaml:"type,omitempty"`
	Required    bool      `json:"required,omitempty" yaml:"required,omitempty"`
	Default     any       `json:"default,omitempty" yaml:"default,omitempty"`
	Options     []string  `json:"options,omitempty" yaml:"options,omitempty"`
	From        string    `json:"from,omitempty" yaml:"from,omitempty"`
	MaxChars    int       `json:"maxChars,omitempty" yaml:"maxChars,omitempty"`
}

// validateUser checks a user input's name, type and attributes (Required, Default, Options).
func (in Input) validateUser() error {
	if !nameRe.MatchString(in.Name) || in.Name == "index" {
		return fmt.Errorf("invalid user input name %q", in.Name)
	}
	t := in.Type.orString()
	if !in.Type.valid() {
		return fmt.Errorf("user input %s: unknown type %q", in.Name, in.Type)
	}
	if in.From != "" || in.MaxChars != 0 {
		return fmt.Errorf("user input %s: from and maxChars are for task inputs", in.Name)
	}
	if in.Required && t == TypeBool {
		return fmt.Errorf("user input %s: a bool cannot be required", in.Name)
	}
	if (t == TypeChoice) != (in.Options != nil) {
		return fmt.Errorf("user input %s: options are for a choice, and a choice needs them", in.Name)
	}
	if t == TypeChoice {
		if len(in.Options) < 2 {
			return fmt.Errorf("user input %s: a choice needs two or more options", in.Name)
		}
		seen := map[string]bool{}
		for _, o := range in.Options {
			if strings.TrimSpace(o) == "" || seen[o] {
				return fmt.Errorf("user input %s: options must be distinct and not blank", in.Name)
			}
			seen[o] = true
		}
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
		if t == TypeString || t == TypeText || (t == TypeChoice && slices.Contains(in.Options, d)) {
			return nil
		}
	}
	return fmt.Errorf("user input %s: default %v does not fit a %s", in.Name, in.Default, t)
}

// Fill returns vars with every declared variable that has no value set empty (an empty string or list, false),
// so its placeholders render as nothing, and each derived input set from its source (From), cut to MaxChars
// characters. A {name} that is not a declared variable stays as written.
func (c Chain) Fill(vars Vars) Vars {
	out := Vars{}
	for k, v := range vars {
		out[k] = v
	}
	for _, in := range slices.Concat(c.Spec.UserInputs, c.Spec.Inputs) {
		if _, ok := out[in.Name]; ok || in.From != "" {
			continue
		}
		out[in.Name] = empty(in.Type)
	}
	for _, in := range c.Spec.Inputs {
		if in.From == "" {
			continue
		}
		s := text(vars[in.From])
		if r := []rune(s); in.MaxChars > 0 && len(r) > in.MaxChars {
			s = string(r[:in.MaxChars])
		}
		out[in.Name] = s
	}
	return out
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

// InputType is the type of a variable's value and, for a user input, its field in the run form: string (the
// default, "" in a file; one line), text (a string over lines), textList ([]string, a text per item), bool, or
// choice (a string picked from the input's Options; user inputs only).
type InputType string

const (
	TypeString   InputType = "string"
	TypeText     InputType = "text"
	TypeTextList InputType = "textList"
	TypeBool     InputType = "bool"
	TypeChoice   InputType = "choice"
)

func (t InputType) valid() bool {
	switch t.orString() {
	case TypeString, TypeText, TypeTextList, TypeBool, TypeChoice:
		return true
	}
	return false
}

func (t InputType) orString() InputType {
	if t == "" {
		return TypeString
	}
	return t
}

// empty is the value of a variable of type t that is not given: an empty string or list, or false.
func empty(t InputType) any {
	switch t.orString() {
	case TypeTextList:
		return []string{}
	case TypeBool:
		return false
	}
	return ""
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

// UserVars turns the run form's values into Vars, checking each against its user input's type: a string (for a
// choice, one of its options), a bool, or a list of strings whose blank items are dropped. A missing user input
// takes its Default or is left out; a Required one that is missing or blank is refused.
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
			v = in.Default
		}
		if v == nil {
			if in.Required {
				return nil, fmt.Errorf("user input %s is required", in.Name)
			}
			continue
		}
		switch in.Type.orString() {
		case TypeString, TypeText, TypeChoice:
			s, ok := v.(string)
			if !ok {
				return nil, fmt.Errorf("user input %s must be a string", in.Name)
			}
			if in.Type == TypeChoice && !slices.Contains(in.Options, s) {
				return nil, fmt.Errorf("user input %s must be one of %s", in.Name, strings.Join(in.Options, ", "))
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
		if in.Required && !present(vars[in.Name]) {
			return nil, fmt.Errorf("user input %s is required", in.Name)
		}
	}
	return vars, nil
}

// Step is one LLM call of a chain, LangChain's prompt | model | parser: the prompt is rendered, sent to the
// model, and the response is read in the step's Format. Output names the parsed value when a later step reads it (like ADK's
// output_key); a step whose value only goes back to the service needs none.
//
// A step with Uses is a function step instead, like LangChain's RunnableLambda or a GitHub Actions step's uses: it
// calls the Go function named there (Functions) with no model and no messages, and saves its value under Output.
// If runs it only when that variable is given (present); otherwise its output is empty.
type Step struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description" yaml:"description"`
	Uses        string `json:"uses,omitempty" yaml:"uses,omitempty"`
	If          string `json:"if,omitempty" yaml:"if,omitempty"`
	Prompt      `yaml:",inline"`
	Format      Format `json:"format,omitempty" yaml:"format,omitempty"`
	Output      string `json:"output,omitempty" yaml:"output,omitempty"`
}

// Function is the form of a function step's Go function: the variables it reads and the type of its value.
type Function struct {
	Inputs []string
	Output InputType
}

// Functions are the functions a function step may use. Their code is given to RunFunctions by the task that runs
// the chain, so this package does not depend on the services they call.
var Functions = map[string]Function{
	// wiki/backlinkTexts reads, in the pages that link to the page titled title, the text around each link.
	"wiki/backlinkTexts": {Inputs: []string{"title"}, Output: TypeTextList},
}

// Func is the code of a function step: it reads the run's variables and returns the step's value.
type Func func(ctx context.Context, vars Vars) (any, error)

// RunFunctions runs the chain's function steps in order with funcs and returns vars with their outputs set. A step
// whose If is not given is skipped and its output set empty, as Fill sets unset variables.
func (c Chain) RunFunctions(ctx context.Context, vars Vars, funcs map[string]Func) (Vars, error) {
	out := Vars{}
	for k, v := range vars {
		out[k] = v
	}
	for _, st := range c.Spec.Steps {
		if st.Uses == "" {
			continue
		}
		if st.If != "" && !present(out[st.If]) {
			out[st.Output] = empty(Functions[st.Uses].Output)
			continue
		}
		fn, ok := funcs[st.Uses]
		if !ok {
			return nil, fmt.Errorf("step %s: no code for %s", st.Name, st.Uses)
		}
		v, err := fn(ctx, out)
		if err != nil {
			return nil, fmt.Errorf("step %s: %w", st.Name, err)
		}
		out[st.Output] = v
	}
	return out, nil
}

// PromptIndex is the index of the chain's first LLM step: the one AI edit sends and a custom template rewrites.
func (c Chain) PromptIndex() int {
	for i, st := range c.Spec.Steps {
		if st.Uses == "" {
			return i
		}
	}
	panic(fmt.Sprintf("chain %s has no LLM step", c.Name))
}

// PromptStep returns the chain's first LLM step (PromptIndex).
func (c Chain) PromptStep() Step {
	return c.Spec.Steps[c.PromptIndex()]
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

// Prompt is a chat prompt template. Message contents may hold {name} placeholders for the chain's variables or
// outputs of earlier steps. Inputs are the variables its messages use (placeholders, If, Each), found when the
// chain loads, like LangChain's input_variables.
type Prompt struct {
	Inputs   []string  `json:"inputs" yaml:"-"`
	Messages []Message `json:"messages" yaml:"messages"`
}

// messageVariables are the variables messages use, in the order they first appear: placeholders ({index} inside
// an Each aside), If and Each.
func messageVariables(messages []Message) []string {
	var out []string
	add := func(name string) {
		if name != "" && !slices.Contains(out, name) {
			out = append(out, name)
		}
	}
	for _, m := range messages {
		add(m.If)
		add(m.Each)
		for _, ph := range placeholders(m.Content) {
			if ph.name != "index" || m.Each == "" {
				add(ph.name)
			}
		}
	}
	return out
}

// Message is one message of a prompt. If sends it only when that variable is given (present). Each sends it once
// per item of a textList variable, with {name} as the item and {index} as its number from 1; an empty list sends
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
	var b strings.Builder
	last := 0
	for _, ph := range placeholders(content) {
		v, ok := own[ph.name]
		if !ok {
			var given any
			if given, ok = vars[ph.name]; ok {
				v = text(given)
			}
		}
		if !ok {
			continue
		}
		b.WriteString(content[last:ph.start])
		b.WriteString(v)
		last = ph.end
	}
	b.WriteString(content[last:])
	return b.String()
}

type placeholder struct {
	name       string
	start, end int
}

// placeholders finds the {name} placeholders of content. One right after another brace is wiki markup
// ({{Infobox}}, {{title}}, {{{1}}}) and is left out; one inside a template's parameter ({{틀|{title}}}) is not.
func placeholders(content string) []placeholder {
	var out []placeholder
	for _, m := range placeholderRe.FindAllStringSubmatchIndex(content, -1) {
		if m[0] > 0 && content[m[0]-1] == '{' {
			continue
		}
		out = append(out, placeholder{name: content[m[2]:m[3]], start: m[0], end: m[1]})
	}
	return out
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
	for i := range c.Spec.Steps {
		c.Spec.Steps[i].Inputs = c.Spec.Steps[i].variables()
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
// or earlier outputs and each is used by its messages; placeholders, If and Each name the step's inputs, Each a
// textList variable, and a textList variable appears as a placeholder only in its own Each message; every task input is used.
func (c Chain) Validate() error {
	for _, in := range c.Spec.UserInputs {
		if err := in.validateUser(); err != nil {
			return err
		}
	}
	for _, in := range c.Spec.Inputs {
		if in.Required || in.Default != nil || in.Options != nil || in.Type == TypeChoice {
			return fmt.Errorf("input %s: required, default and choice are for user inputs", in.Name)
		}
	}
	types := map[string]InputType{}
	for _, in := range slices.Concat(c.Spec.UserInputs, c.Spec.Inputs) {
		if !nameRe.MatchString(in.Name) || in.Name == "index" {
			return fmt.Errorf("invalid variable name %q", in.Name)
		}
		if _, dup := types[in.Name]; dup {
			return fmt.Errorf("duplicate variable %q", in.Name)
		}
		t := in.Type.orString()
		if !in.Type.valid() {
			return fmt.Errorf("variable %s: unknown type %q", in.Name, in.Type)
		}
		types[in.Name] = t
	}
	for _, in := range c.Spec.Inputs {
		if in.From == "" {
			if in.MaxChars != 0 {
				return fmt.Errorf("input %s: maxChars needs from", in.Name)
			}
			continue
		}
		if t, ok := types[in.From]; !ok || (t != TypeString && t != TypeText) || in.From == in.Name {
			return fmt.Errorf("input %s: from %q is not a string variable", in.Name, in.From)
		}
		if in.MaxChars < 0 || types[in.Name] != TypeString {
			return fmt.Errorf("input %s: a derived input is a string with maxChars of 0 or more", in.Name)
		}
	}
	if len(c.Spec.Steps) == 0 {
		return fmt.Errorf("no steps")
	}
	used := map[string]bool{}
	prompt := false
	for _, st := range c.Spec.Steps {
		if err := st.validate(types, used); err != nil {
			return fmt.Errorf("step %s: %w", st.Name, err)
		}
		prompt = prompt || st.Uses == ""
		if st.Output == "" {
			continue
		}
		if _, dup := types[st.Output]; dup || !nameRe.MatchString(st.Output) || st.Output == "index" {
			return fmt.Errorf("step %s: invalid or duplicate output %q", st.Name, st.Output)
		}
		types[st.Output] = st.outputType()
	}
	if !prompt {
		return fmt.Errorf("no LLM step")
	}
	for _, in := range c.Spec.Inputs {
		if in.From != "" && used[in.Name] {
			used[in.From] = true
		}
	}
	for _, in := range c.Spec.Inputs {
		if !used[in.Name] {
			return fmt.Errorf("input %q is not used", in.Name)
		}
	}
	return nil
}

// ValidateMessages checks messages written to replace those of the chain's first LLM step (a custom template) by
// the rules of Validate.
func (c Chain) ValidateMessages(messages []Message) error {
	i := c.PromptIndex()
	st := c.Spec.Steps[i]
	st.Messages = messages
	st.Inputs = messageVariables(messages)
	return st.validate(c.typesBefore(i), map[string]bool{})
}

// WithUserInputs returns the chain with inputs as its user inputs (a custom template's run form), checked like
// Validate checks them and not clashing with the task's inputs.
func (c Chain) WithUserInputs(inputs []Input) (Chain, error) {
	seen := map[string]bool{}
	for _, in := range c.Spec.Inputs {
		seen[in.Name] = true
	}
	for _, st := range c.Spec.Steps {
		if st.Output != "" {
			seen[st.Output] = true
		}
	}
	for _, in := range inputs {
		if err := in.validateUser(); err != nil {
			return c, err
		}
		if seen[in.Name] {
			return c, fmt.Errorf("user input %q is already a variable", in.Name)
		}
		seen[in.Name] = true
	}
	c.Spec.UserInputs = slices.Clone(inputs)
	c.Spec.Steps = slices.Clone(c.Spec.Steps)
	return c, nil
}

// Template is the part of a chain a custom template rewrites, as written in YAML: the run form and the first LLM
// step's messages.
type Template struct {
	UserInputs []Input   `json:"userInputs" yaml:"userInputs"`
	Messages   []Message `json:"messages" yaml:"messages"`
}

// Template returns the chain's own run form and first LLM step's messages.
func (c Chain) Template() Template {
	return Template{UserInputs: c.Spec.UserInputs, Messages: c.PromptStep().Messages}
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
	if t.UserInputs == nil {
		t.UserInputs = []Input{}
	}
	return t, nil
}

// typesBefore are the types of the chain's variables and of the outputs of its steps before step i.
func (c Chain) typesBefore(i int) map[string]InputType {
	types := map[string]InputType{}
	for _, in := range slices.Concat(c.Spec.UserInputs, c.Spec.Inputs) {
		types[in.Name] = in.Type.orString()
	}
	for _, st := range c.Spec.Steps[:i] {
		if st.Output != "" {
			types[st.Output] = st.outputType()
		}
	}
	return types
}

// outputType is the type of the step's output: its function's for a function step, a string for an LLM step.
func (st Step) outputType() InputType {
	if st.Uses != "" {
		return Functions[st.Uses].Output
	}
	return TypeString
}

// variables are the variables a step reads: its function's inputs and If, or what its messages use.
func (st Step) variables() []string {
	if st.Uses == "" {
		return messageVariables(st.Messages)
	}
	vars := slices.Clone(Functions[st.Uses].Inputs)
	if st.If != "" && !slices.Contains(vars, st.If) {
		vars = append(vars, st.If)
	}
	return vars
}

// validate checks a step: the variables its messages use are chain variables or earlier outputs, If and Each name them,
// Each a textList one, and a textList variable is a placeholder only in its own Each message.
func (st Step) validate(types map[string]InputType, used map[string]bool) error {
	if st.Uses != "" {
		return st.validateFunction(types, used)
	}
	if st.If != "" {
		return fmt.Errorf("if is for function steps; a message has its own if")
	}
	if st.Name == "" || len(st.Messages) == 0 {
		return fmt.Errorf("name and messages are required")
	}
	if f := st.Format.OrText(); f != FormatJSON && f != FormatText {
		return fmt.Errorf("unknown format %q", st.Format)
	}
	for _, name := range messageVariables(st.Messages) {
		if _, ok := types[name]; !ok {
			return fmt.Errorf("{%s} is neither a variable nor an earlier output", name)
		}
		used[name] = true
	}
	for _, m := range st.Messages {
		if m.Content == "" || (m.Role != "system" && m.Role != "user") {
			return fmt.Errorf("invalid %q message", m.Role)
		}
		if m.Each != "" && types[m.Each] != TypeTextList {
			return fmt.Errorf("each %q is not a textList variable", m.Each)
		}
		for _, ph := range placeholders(m.Content) {
			if types[ph.name] == TypeTextList && ph.name != m.Each {
				return fmt.Errorf("textList variable {%s} is used outside its each message", ph.name)
			}
		}
	}
	return nil
}

// validateFunction checks a function step: its function exists, its inputs and If are variables or earlier outputs
// (If of any type, inputs of the type they are read as), it names an output, and it has no messages or format.
func (st Step) validateFunction(types map[string]InputType, used map[string]bool) error {
	fn, ok := Functions[st.Uses]
	if !ok {
		return fmt.Errorf("unknown function %q", st.Uses)
	}
	if st.Name == "" || st.Output == "" || len(st.Messages) > 0 || st.Format != "" {
		return fmt.Errorf("a function step has a name and an output, and no messages or format")
	}
	for _, name := range st.variables() {
		if _, ok := types[name]; !ok {
			return fmt.Errorf("%s reads %q, which is neither a variable nor an earlier output", st.Uses, name)
		}
		used[name] = true
	}
	for _, name := range fn.Inputs {
		if t := types[name]; t != TypeString && t != TypeText {
			return fmt.Errorf("%s reads %q as a string, but it is %s", st.Uses, name, t)
		}
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
