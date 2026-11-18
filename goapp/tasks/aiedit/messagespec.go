package aiedit

import (
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.com/zetaoss/zengine/goapp/models"
	"github.com/zetaoss/zengine/goapp/prompts"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
)

// taggedPromptTemplate is a template spec as it was saved before message.system and message.user: one template
// block of <message role="..."> sections, and a template_format whose only value was jinja2.
type taggedPromptTemplate struct {
	Name           string                  `yaml:"name"`
	Description    string                  `yaml:"description"`
	TemplateFormat string                  `yaml:"template_format"`
	Template       string                  `yaml:"template"`
	InputVariables []prompts.InputVariable `yaml:"input_variables"`
}

// taggedMessageOpen and the roles are those of the parser that read tagged templates: an open tag on a line of
// its own, at its start, with the role in either kind of quotes.
var (
	taggedMessageOpen = regexp.MustCompile(`^<message\s+role\s*=\s*(?:"([a-z]+)"|'([a-z]+)')\s*>$`)
	taggedRoles       = []string{"system", "user", "assistant"}
)

type taggedMessage struct {
	role string
	body string
}

// splitTaggedMessages splits a tagged template into its messages as that parser did: blank lines between
// messages are skipped, any other text outside one is an error, and each body is the text between its tags.
func splitTaggedMessages(src string) ([]taggedMessage, error) {
	var messages []taggedMessage
	lines := strings.SplitAfter(src, "\n")
	offset, bodyStart, openLine := 0, 0, 0
	role := ""
	for i, raw := range lines {
		line := strings.TrimRight(raw, " \t\r\n")
		n := i + 1
		switch {
		case role == "" && line == "":
		case role == "":
			m := taggedMessageOpen.FindStringSubmatch(line)
			if m == nil {
				return nil, fmt.Errorf("line %d: text outside a message", n)
			}
			role = m[1] + m[2]
			found := false
			for _, r := range taggedRoles {
				found = found || r == role
			}
			if !found {
				return nil, fmt.Errorf("line %d: role %q is not one of %s", n, role, strings.Join(taggedRoles, ", "))
			}
			bodyStart, openLine = offset+len(raw), n
		case line == "</message>":
			messages = append(messages, taggedMessage{role: role, body: src[bodyStart:offset]})
			role = ""
		case strings.HasPrefix(line, "<message"):
			return nil, fmt.Errorf("line %d: a message opened at line %d is not closed by </message>", n, openLine)
		}
		offset += len(raw)
	}
	if role != "" {
		return nil, fmt.Errorf("line %d: a message is not closed by </message>", openLine)
	}
	if len(messages) == 0 {
		return nil, errors.New("the template has no message")
	}
	return messages, nil
}

// MessageSpec rewrites a tagged template spec with message.system and message.user. It refuses, rather than
// changing what the model is sent, a template it cannot carry over as it is: one with an assistant message, two
// messages of one role, or a user message before the system one.
func MessageSpec(spec string) (string, error) {
	var t taggedPromptTemplate
	dec := yaml.NewDecoder(strings.NewReader(spec))
	dec.KnownFields(true)
	if err := dec.Decode(&t); err != nil {
		return "", fmt.Errorf("YAML: %w", err)
	}
	if t.TemplateFormat != "" && t.TemplateFormat != "jinja2" {
		return "", fmt.Errorf("template_format %q is not jinja2", t.TemplateFormat)
	}
	if strings.TrimSpace(t.Template) == "" {
		return "", errors.New("the spec has no template")
	}
	messages, err := splitTaggedMessages(t.Template)
	if err != nil {
		return "", fmt.Errorf("template: %w", err)
	}
	var m prompts.MessageTemplates
	for i, msg := range messages {
		body := strings.TrimSuffix(strings.TrimSuffix(msg.body, "\n"), "\r")
		switch {
		case msg.role == "assistant":
			return "", fmt.Errorf("message %d: an assistant message cannot be carried over", i+1)
		case msg.role == "system" && i > 0:
			return "", fmt.Errorf("message %d: a system message after another message cannot be carried over", i+1)
		case msg.role == "user" && m.User != "":
			return "", fmt.Errorf("message %d: a second user message cannot be carried over", i+1)
		case msg.role == "system":
			m.System = body
		default:
			m.User = body
		}
	}
	if t.InputVariables == nil {
		t.InputVariables = []prompts.InputVariable{}
	}
	out := prompts.PromptTemplate{Name: t.Name, Description: t.Description, Message: m, InputVariables: t.InputVariables}
	return out.YAML(), nil
}

// ConvertMessageSpecs rewrites every template spec still written as tagged messages (MessageSpec), and reports
// how many it rewrote. A spec it cannot rewrite is logged and left as it is; its owner repairs it in the chains
// tab, which edits an unreadable spec as YAML. `tool migrate` runs it after ConvertLegacyPrompts.
func ConvertMessageSpecs(db *gorm.DB) (int, error) {
	var rows []models.AIEditPrompt
	if err := db.Where("spec IS NOT NULL AND spec <> ''").Find(&rows).Error; err != nil {
		return 0, err
	}
	converted := 0
	for _, r := range rows {
		if _, err := prompts.ParsePromptTemplate(r.Spec); err == nil {
			continue
		}
		spec, err := MessageSpec(r.Spec)
		if err != nil {
			slog.Error("aiedit: cannot rewrite template spec with message keys", "id", r.ID, "err", err)
			continue
		}
		if err := db.Model(&models.AIEditPrompt{}).Where("id = ?", r.ID).Update("spec", spec).Error; err != nil {
			return converted, err
		}
		converted++
	}
	return converted, nil
}
