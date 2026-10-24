package prompts

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Chain is a pipeline of steps, following LangChain. Each step's output is saved under its Output name and
// can be an input of later steps. Metadata holds free key-value pairs, like LangChain's RunnableConfig
// metadata; a chain uses them as it needs (doc-eval keeps a "version"). Chains are defined in code and listed
// read-only in the AI tool's chains tab (the aiedit prompts are stored in the database and editable).
type Chain struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	Inputs      []string          `json:"inputs"`
	Steps       []Step            `json:"steps"`
}

// Step is one LLM call of a chain, LangChain's prompt | model | parser: the prompt is rendered, sent to the
// model, and the response is parsed into the value named by Output.
type Step struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Prompt      Prompt `json:"prompt"`
	Parser      Parser `json:"parser"`
	Output      string `json:"output"`
}

// Prompt is a chat prompt template. Message contents may hold {name} placeholders for the variables listed in
// Inputs, which are chain inputs or outputs of earlier steps.
type Prompt struct {
	Inputs   []string  `json:"inputs"`
	Messages []Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
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

func BuiltinChains() []Chain {
	return []Chain{DocDefine, DocEval, DocCreate, DocUpdate}
}
