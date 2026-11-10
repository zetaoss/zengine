package aiedit

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/zetaoss/zengine/goapp/app/config"
	"github.com/zetaoss/zengine/goapp/models"
	"github.com/zetaoss/zengine/goapp/prompts"
	"github.com/zetaoss/zengine/goapp/services/wikisvc"
)

// Request is what a requester asks AI edit for: a page, its request type, an AI edit prompt template or none
// (the chain's own messages), and the run form's values (the chain's user inputs).
type Request struct {
	RequestType string
	Title       string
	// Wikitext is the page as it stands in the editor, for update, even when emptied; when nil the saved page is
	// read.
	Wikitext   *string
	Prompt     *models.AIEditPrompt
	UserInputs map[string]any
}

// Rendered is a request rendered on its chain: the variables it filled and the messages to send.
type Rendered struct {
	Chain    prompts.Chain
	Vars     prompts.Vars
	Messages []prompts.Message
}

// InputError is a request the requester has to fix, answered with 422.
type InputError struct{ msg string }

func (e InputError) Error() string { return e.msg }

// IsInputError reports whether err is an InputError.
func IsInputError(err error) bool {
	var ie InputError
	return errors.As(err, &ie)
}

// ChainOf returns the chain of a request type: doc-update for update, doc-create otherwise.
func ChainOf(requestType string) prompts.Chain {
	if requestType == models.AIEditRequestUpdate {
		return prompts.DocUpdate
	}
	return prompts.DocCreate
}

// Render fills the chain's variables (the run form's values as given; the title and, for update, the wikitext as
// body) and renders its first step. Backlinks are a run form value: the run form fetches them (Backlinks).
func Render(ctx context.Context, cfg *config.Config, req Request) (Rendered, error) {
	chain := ChainOf(req.RequestType)
	if req.Prompt != nil {
		var err error
		if chain, err = TemplateChain(*req.Prompt); err != nil {
			return Rendered{}, fmt.Errorf("template %d: %w", req.Prompt.ID, err)
		}
	}
	vars, err := chain.UserVars(req.UserInputs)
	if err != nil {
		return Rendered{}, InputError{err.Error()}
	}
	if err := checkSizes(vars); err != nil {
		return Rendered{}, InputError{err.Error()}
	}
	for _, in := range chain.Inputs {
		switch in.Name {
		case "title":
			vars["title"] = req.Title
		case "body":
			var text string
			if req.Wikitext != nil {
				text = *req.Wikitext
			} else if text, _, err = wikisvc.PageText(ctx, cfg, req.Title); err != nil {
				return Rendered{}, fmt.Errorf("read page: %w", err)
			}
			vars["body"] = text
		default:
			return Rendered{}, fmt.Errorf("chain %s: no way to fill input %q", chain.Name, in.Name)
		}
	}
	messages := chain.PromptStep().Render(vars)
	if !slices.ContainsFunc(messages, func(m prompts.Message) bool { return m.Role == "user" }) {
		return Rendered{}, InputError{"보낼 user 메시지가 없습니다. 템플릿의 조건(if, for)이 걸린 입력을 채워 주세요."}
	}
	return Rendered{Chain: chain, Vars: vars, Messages: messages}, nil
}

// Limits on run form values, which the requester's browser sends: the items of a list and the characters of an
// item or a string.
const (
	MaxListItems  = 20
	MaxValueChars = 100000
)

func checkSizes(vars prompts.Vars) error {
	for name, v := range vars {
		switch v := v.(type) {
		case string:
			if utf8.RuneCountInString(v) > MaxValueChars {
				return fmt.Errorf("%s는 %d자까지 쓸 수 있습니다", name, MaxValueChars)
			}
		case []string:
			if len(v) > MaxListItems {
				return fmt.Errorf("%s는 %d개까지 쓸 수 있습니다", name, MaxListItems)
			}
			for _, s := range v {
				if utf8.RuneCountInString(s) > MaxValueChars {
					return fmt.Errorf("%s의 항목은 %d자까지 쓸 수 있습니다", name, MaxValueChars)
				}
			}
		}
	}
	return nil
}

// Backlinks lists the titles of the pages that link to title, for the run form's backlinks field.
func Backlinks(ctx context.Context, cfg *config.Config, title string) ([]string, error) {
	return wikisvc.BacklinkTitles(ctx, cfg, title)
}

// TemplatePrompt reads a template's document (aiedit_prompts.spec, prompts.PromptTemplate).
func TemplatePrompt(p models.AIEditPrompt) (prompts.PromptTemplate, error) {
	if strings.TrimSpace(p.Spec) == "" {
		return prompts.PromptTemplate{}, errors.New("the template has no spec")
	}
	return prompts.ParsePromptTemplate(p.Spec)
}

// TemplateChain is the chain a template runs as: its request type's chain with the template's run form and
// template (prompts.Chain.WithTemplate).
func TemplateChain(p models.AIEditPrompt) (prompts.Chain, error) {
	t, err := TemplatePrompt(p)
	if err != nil {
		return prompts.Chain{}, err
	}
	return ChainOf(p.RequestType).WithTemplate(t)
}

// MessagesText joins rendered messages for people to read (aiedit_tasks.llm_input).
func MessagesText(messages []prompts.Message) string {
	parts := make([]string, len(messages))
	for i, m := range messages {
		parts[i] = m.Content
	}
	return strings.Join(parts, "\n\n")
}
