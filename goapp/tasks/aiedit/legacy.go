package aiedit

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"
	"slices"
	"strings"

	"github.com/zetaoss/zengine/goapp/models"
	"github.com/zetaoss/zengine/goapp/prompts"
	"gorm.io/gorm"
)

// legacyPrompt is a template saved before templates had a spec: its instructions as content (the oldest) or as
// system messages (messages, JSON), and the user inputs it picked (user_inputs, JSON; none: all of its chain's).
type legacyPrompt struct {
	ID          int
	Title       string
	Description string
	RequestType string
	Content     *string
	Messages    *string
	UserInputs  *string
	Model       string
}

// ConvertLegacyPrompts gives every template without a spec one built from its old columns (LegacySpec), and
// reports how many it converted. A template it cannot convert is logged and left without a spec, so it fails to
// run until it is fixed by hand. `tool migrate` runs it after the SQL migrations.
func ConvertLegacyPrompts(db *gorm.DB) (int, error) {
	var rows []legacyPrompt
	if err := db.Raw("SELECT id, title, description, request_type, content, messages, user_inputs, model FROM aiedit_prompts WHERE spec IS NULL OR spec = ''").Scan(&rows).Error; err != nil {
		return 0, err
	}
	converted := 0
	for _, r := range rows {
		spec, err := LegacySpec(r)
		if err != nil {
			slog.Error("aiedit: cannot convert template", "id", r.ID, "err", err)
			continue
		}
		if err := db.Model(&models.AIEditPrompt{}).Where("id = ?", r.ID).Update("spec", spec).Error; err != nil {
			return converted, err
		}
		converted++
	}
	return converted, nil
}

// legacyRenames maps names a stored template may use to the names of the chains now.
var legacyRenames = map[string]string{
	"wikitext": "body", "backlinkTexts": "backlinks", "backlinkTitles": "backlinks",
	"readBacklinks": "backlinks", "useBacklinks": "backlinks", "extraTexts": "texts", "copiedTexts": "texts",
}

var (
	legacyPlaceholder = regexp.MustCompile(`\{([A-Za-z][A-Za-z0-9_]*)\}`)
	wikiBareName      = regexp.MustCompile(`\{\{\s*[A-Za-z][A-Za-z0-9_.]*\s*\}\}`)
	jinjaTagStart     = regexp.MustCompile(`\{%|\{#`)
)

// LegacySpec builds a spec for a template saved before specs. Its instructions become the system message, with
// {name} placeholders as {{ name }} and wiki markup that would read as template syntax ({{PAGENAME}}, {%, {#)
// inside raw; the user message lists the variables the instructions do not already read, in the chain's order,
// as the built-in chains do.
func LegacySpec(r legacyPrompt) (string, error) {
	chain := ChainOf(r.RequestType)
	picked := map[string]bool{}
	if r.UserInputs == nil || strings.TrimSpace(*r.UserInputs) == "" {
		for _, in := range chain.UserInputs {
			picked[in.Name] = true
		}
	} else {
		var own []struct {
			Name string `json:"name"`
		}
		if err := json.Unmarshal([]byte(*r.UserInputs), &own); err != nil {
			return "", fmt.Errorf("user_inputs: %w", err)
		}
		for _, in := range own {
			picked[legacyName(in.Name)] = true
		}
	}

	system := ""
	if r.Messages != nil && strings.TrimSpace(*r.Messages) != "" {
		var messages []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}
		if err := json.Unmarshal([]byte(*r.Messages), &messages); err != nil {
			return "", fmt.Errorf("messages: %w", err)
		}
		var parts []string
		for _, m := range messages {
			if m.Role == "system" {
				parts = append(parts, m.Content)
			}
		}
		system = strings.Join(parts, "\n\n")
	} else if r.Content != nil {
		system = *r.Content
	}

	known := map[string]bool{}
	for _, in := range slices.Concat(chain.Inputs, chain.UserInputs) {
		known[in.Name] = true
	}
	used := map[string]bool{}
	system = jinjaTagStart.ReplaceAllStringFunc(system, func(s string) string { return "{% raw %}" + s + "{% endraw %}" })
	system = wikiBareName.ReplaceAllStringFunc(system, func(s string) string { return "{% raw %}" + s + "{% endraw %}" })
	system = legacyPlaceholder.ReplaceAllStringFunc(system, func(s string) string {
		name := legacyName(s[1 : len(s)-1])
		if !known[name] {
			return s
		}
		used[name] = true
		return "{{ " + name + " }}"
	})

	var user strings.Builder
	user.WriteString("# 변수\n")
	block := func(name, body string) {
		user.WriteString("\n<" + name + ">\n" + body + "\n</" + name + ">")
	}
	for _, in := range chain.Inputs {
		if !used[in.Name] {
			block(in.Name, "{{ "+in.Name+" }}")
		}
	}
	for _, in := range chain.UserInputs {
		if !picked[in.Name] || used[in.Name] {
			continue
		}
		switch {
		case in.Name == "backlinks":
			user.WriteString("\n{% if backlinks %}")
			block("backlinks", "{% for b in backlinks %}\n{{ b }}\n{% endfor %}")
			user.WriteString("\n{% endif %}")
		case in.Type() == prompts.TypeTextList:
			user.WriteString("\n{% for t in " + in.Name + " %}")
			block(in.Name, "{{ t }}")
			user.WriteString("\n{% endfor %}")
		default:
			user.WriteString("\n{% if " + in.Name + " %}")
			block(in.Name, "{{ "+in.Name+" }}")
			user.WriteString("\n{% endif %}")
		}
	}

	t := prompts.PromptTemplate{
		Name:           r.Title,
		Description:    r.Description,
		TemplateFormat: prompts.TemplateFormatJinja2,
		Template:       "<message role=\"system\">\n" + strings.TrimSpace(system) + "\n</message>\n<message role=\"user\">\n" + strings.TrimSpace(user.String()) + "\n</message>",
	}
	for _, in := range slices.Concat(chain.Inputs, chain.UserInputs) {
		if picked[in.Name] || used[in.Name] || slices.ContainsFunc(chain.Inputs, func(x prompts.Input) bool { return x.Name == in.Name }) {
			t.InputVariables = append(t.InputVariables, prompts.InputVariable{Name: in.Name})
		}
	}
	if _, err := chain.WithTemplate(t); err != nil {
		return "", err
	}
	return t.YAML(), nil
}

func legacyName(name string) string {
	if now, ok := legacyRenames[name]; ok {
		return now
	}
	return name
}
