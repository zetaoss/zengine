package prompts

import (
	"errors"
	"fmt"
	"strings"

	"github.com/zetaoss/zengine/goapp/prompts/jinja"
)

// MessageTemplates are a step's messages as written, a jinja template per role: System (the instructions) and
// User (the request, with the values it reads). Either may be left out, but not both. Each is its own template,
// so a value holding the other's text stays inside its message.
type MessageTemplates struct {
	System string `json:"system,omitempty" yaml:"system,omitempty"`
	User   string `json:"user,omitempty" yaml:"user,omitempty"`
}

// ChatTemplate is a step's messages parsed, in the order they are sent: system, then user.
type ChatTemplate struct {
	Messages []MessageTemplate
}

// MessageTemplate is one message of a ChatTemplate.
type MessageTemplate struct {
	Role     string
	Template *jinja.Template
}

// ParseChat parses each message of m; vars are the names they may read. An error names the message it is in
// (message.system or message.user), with its line counted from the message's first line.
func ParseChat(m MessageTemplates, vars []string) (ChatTemplate, error) {
	var ct ChatTemplate
	for _, msg := range []struct{ role, src string }{{"system", m.System}, {"user", m.User}} {
		if strings.TrimSpace(msg.src) == "" {
			continue
		}
		tpl, err := jinja.Parse(msg.src, vars)
		if err != nil {
			return ChatTemplate{}, fmt.Errorf("message.%s: %w", msg.role, err)
		}
		ct.Messages = append(ct.Messages, MessageTemplate{Role: msg.role, Template: tpl})
	}
	if len(ct.Messages) == 0 {
		return ChatTemplate{}, errors.New("no message; write message.system, message.user or both")
	}
	return ct, nil
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
