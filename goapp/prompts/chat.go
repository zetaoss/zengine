package prompts

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/zetaoss/zengine/goapp/prompts/jinja"
)

// ChatTemplate is a template split into messages: lines <message role="..."> and </message> at the start of a
// line open and close each message, and the body between them is a jinja template. The split comes before any
// value is filled in, so a value holding </message> stays inside its message.
type ChatTemplate struct {
	Messages []MessageTemplate
}

// MessageTemplate is one message of a ChatTemplate.
type MessageTemplate struct {
	Role     string
	Template *jinja.Template
	// Offset is the byte offset of the body in the source, and Line its first line, from 1.
	Offset int
	Line   int
}

var (
	messageOpen  = regexp.MustCompile(`^<message\s+role\s*=\s*(?:"([a-z]+)"|'([a-z]+)')\s*>$`)
	messageRoles = []string{"system", "user", "assistant"}
)

// ParseChat splits src into messages and parses each body; vars are the names the bodies may read.
func ParseChat(src string, vars []string) (ChatTemplate, error) {
	var ct ChatTemplate
	lines := strings.SplitAfter(src, "\n")
	offset := 0
	role := ""
	bodyStart, bodyLine, openLine := 0, 0, 0
	for i, raw := range lines {
		line := strings.TrimRight(raw, " \t\r\n")
		n := i + 1
		switch {
		case role == "" && line == "":
		case role == "":
			m := messageOpen.FindStringSubmatch(line)
			if m == nil {
				return ChatTemplate{}, fmt.Errorf("line %d: text outside a message; open one with <message role=\"user\">", n)
			}
			role = m[1] + m[2]
			if !containsString(messageRoles, role) {
				return ChatTemplate{}, fmt.Errorf("line %d: role %q is not one of %s", n, role, strings.Join(messageRoles, ", "))
			}
			bodyStart, bodyLine, openLine = offset+len(raw), n+1, n
		case line == "</message>":
			body := src[bodyStart:offset]
			tpl, err := jinja.Parse(body, vars)
			if err != nil {
				return ChatTemplate{}, shiftError(err, bodyLine-1)
			}
			ct.Messages = append(ct.Messages, MessageTemplate{Role: role, Template: tpl, Offset: bodyStart, Line: bodyLine})
			role = ""
		case strings.HasPrefix(line, "<message"):
			return ChatTemplate{}, fmt.Errorf("line %d: a message opened at line %d is not closed by </message>", n, openLine)
		}
		offset += len(raw)
	}
	if role != "" {
		return ChatTemplate{}, fmt.Errorf("line %d: a message is not closed by </message>", openLine)
	}
	if len(ct.Messages) == 0 {
		return ChatTemplate{}, errors.New("the template has no message; write <message role=\"system\">...</message>")
	}
	return ct, nil
}

func shiftError(err error, lines int) error {
	var je *jinja.Error
	if errors.As(err, &je) {
		return &jinja.Error{Line: je.Line + lines, Col: je.Col, Msg: je.Msg}
	}
	return err
}

func containsString(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

// Uses lists the variables the messages read, each once per kind, in order of first use.
func (ct ChatTemplate) Uses() []jinja.Use {
	var uses []jinja.Use
	for _, m := range ct.Messages {
		for _, u := range m.Template.Uses() {
			if !containsUse(uses, u) {
				uses = append(uses, u)
			}
		}
	}
	return uses
}

func containsUse(list []jinja.Use, u jinja.Use) bool {
	for _, x := range list {
		if x == u {
			return true
		}
	}
	return false
}

// Spans lists the template syntax of every message, at byte offsets of the whole source.
func (ct ChatTemplate) Spans() []jinja.Span {
	var spans []jinja.Span
	for _, m := range ct.Messages {
		for _, s := range m.Template.Spans() {
			spans = append(spans, jinja.Span{Start: s.Start + m.Offset, End: s.End + m.Offset, Kind: s.Kind})
		}
	}
	return spans
}

// Render fills each message and trims its content; a message left blank is not sent.
func (ct ChatTemplate) Render(vars Vars) ([]Message, error) {
	var out []Message
	for _, m := range ct.Messages {
		content, err := m.Template.Render(vars)
		if err != nil {
			return nil, err
		}
		if content = strings.TrimSpace(content); content != "" {
			out = append(out, Message{Role: m.Role, Content: content})
		}
	}
	return out, nil
}
