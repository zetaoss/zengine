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
	// legacyHeading is the start of a heading the editor wrote above a placeholder (-----wikitext-----).
	legacyHeading = regexp.MustCompile(`-----[A-Za-z][A-Za-z0-9_]*`)
)

// LegacySpec builds a spec for a template saved before specs. Its messages are kept in their roles and order,
// each if and each message becoming a block around it (a template saved before messages has its content as the
// system message), with {name} placeholders as {{ name }} and wiki markup that would read as template syntax
// ({{PAGENAME}}, {%, {#) inside raw. The user message then lists in tags the variables the template does not
// already read, in the chain's order, as the built-in chains do.
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

	known := map[string]bool{}
	for _, in := range slices.Concat(chain.Inputs, chain.UserInputs) {
		known[in.Name] = true
	}
	used := map[string]bool{}
	// convert writes stored text as the template reads it: wiki markup that would read as template syntax inside
	// raw, and {name} placeholders as {{ name }} (inside an each message, the item as {{ item }} and {index} as
	// {{ loop.index }}).
	convert := func(s, each string) string {
		s = legacyHeading.ReplaceAllStringFunc(s, func(m string) string { return "-----" + legacyName(m[5:]) })
		s = jinjaTagStart.ReplaceAllStringFunc(s, func(m string) string { return "{% raw %}" + m + "{% endraw %}" })
		s = wikiBareName.ReplaceAllStringFunc(s, func(m string) string { return "{% raw %}" + m + "{% endraw %}" })
		return legacyPlaceholder.ReplaceAllStringFunc(s, func(m string) string {
			name := legacyName(m[1 : len(m)-1])
			switch {
			case each != "" && name == each:
				return "{{ item }}"
			case each != "" && name == "index":
				return "{{ loop.index }}"
			case !known[name]:
				return m
			}
			used[name] = true
			return "{{ " + name + " }}"
		})
	}

	var systemParts, userParts []string
	if r.Messages != nil && strings.TrimSpace(*r.Messages) != "" {
		var messages []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
			If      string `json:"if"`
			Each    string `json:"each"`
		}
		if err := json.Unmarshal([]byte(*r.Messages), &messages); err != nil {
			return "", fmt.Errorf("messages: %w", err)
		}
		// Every message is kept as it ran: its if and each become blocks around it. One conditioned on a variable the
		// chain no longer has (request) is left out, as a run that does not send that variable left it out.
		for _, m := range messages {
			cond, each := legacyName(m.If), legacyName(m.Each)
			if (m.If != "" && !known[cond]) || (m.Each != "" && !known[each]) {
				continue
			}
			part := convert(m.Content, each)
			switch {
			case m.Each != "":
				used[each] = true
				part = "{% for item in " + each + " %}\n" + part + "\n\n{% endfor %}\n"
			case m.If != "":
				used[cond] = true
				part = "{% if " + cond + " %}\n" + part + "\n\n{% endif %}\n"
			default:
				part += "\n\n"
			}
			if m.Role == "system" {
				systemParts = append(systemParts, part)
			} else {
				userParts = append(userParts, part)
			}
		}
	} else if r.Content != nil {
		systemParts = append(systemParts, convert(*r.Content, ""))
	}
	system := strings.Join(systemParts, "")

	// The user message ends with the variables the template does not read yet, in tags under "# 변수"; the heading
	// stands only when one of them is sent.
	var blocks []string
	var conds []string
	for _, in := range chain.Inputs {
		if !used[in.Name] {
			blocks = append(blocks, "<"+in.Name+">\n{{ "+in.Name+" }}\n</"+in.Name+">")
		}
	}
	always := len(blocks) > 0
	for _, in := range chain.UserInputs {
		if !picked[in.Name] || used[in.Name] {
			continue
		}
		conds = append(conds, in.Name)
		switch {
		case in.Name == "backlinks":
			blocks = append(blocks, "{% if backlinks %}\n<backlinks>\n{% for b in backlinks %}\n{{ b }}\n{% endfor %}\n</backlinks>\n{% endif %}")
		case in.Type() == prompts.TypeTextList:
			blocks = append(blocks, "{% for t in "+in.Name+" %}\n<"+in.Name+">\n{{ t }}\n</"+in.Name+">\n{% endfor %}")
		default:
			blocks = append(blocks, "{% if "+in.Name+" %}\n<"+in.Name+">\n{{ "+in.Name+" }}\n</"+in.Name+">\n{% endif %}")
		}
	}
	if len(blocks) > 0 {
		heading := "# 변수\n\n"
		if !always {
			heading = "{% if " + strings.Join(conds, " or ") + " %}\n# 변수\n\n{% endif %}\n"
		}
		userParts = append(userParts, heading+strings.Join(blocks, "\n"))
	}
	var tpl strings.Builder
	if s := strings.TrimSpace(system); s != "" {
		tpl.WriteString("<message role=\"system\">\n" + s + "\n</message>\n")
	}
	tpl.WriteString("<message role=\"user\">\n" + strings.TrimSpace(strings.Join(userParts, "")) + "\n</message>")
	t := prompts.PromptTemplate{
		Name:           r.Title,
		Description:    r.Description,
		TemplateFormat: prompts.TemplateFormatJinja2,
		Template:       tpl.String(),
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
