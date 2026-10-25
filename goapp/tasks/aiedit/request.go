package aiedit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

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

// Render fills the chain's variables (the user inputs as given; the title, the wikitext and, with readBacklinks,
// the excerpts of the pages that link here) and renders its step.
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
	for _, in := range chain.Spec.Inputs {
		if in.From != "" {
			continue
		}
		switch in.Name {
		case "title":
			vars["title"] = req.Title
		case "wikitext":
			var text string
			if req.Wikitext != nil {
				text = *req.Wikitext
			} else if text, _, err = wikisvc.PageText(ctx, cfg, req.Title); err != nil {
				return Rendered{}, fmt.Errorf("read page: %w", err)
			}
			vars["wikitext"] = text
		case "backlinkTexts":
			texts := []string{}
			if read, _ := vars["readBacklinks"].(bool); read {
				if texts, err = wikisvc.BacklinkTexts(ctx, cfg, req.Title); err != nil {
					return Rendered{}, fmt.Errorf("read backlinks: %w", err)
				}
			}
			vars["backlinkTexts"] = texts
		default:
			return Rendered{}, fmt.Errorf("chain %s: no way to fill input %q", chain.Name, in.Name)
		}
	}
	vars = chain.Fill(vars)
	messages := chain.Spec.Steps[0].Render(vars)
	if len(messages) == 0 {
		return Rendered{}, InputError{"보낼 메시지가 없습니다. 조건(if, each)이 걸린 입력을 채워 주세요."}
	}
	return Rendered{Chain: chain, Vars: vars, Messages: messages}, nil
}

// TemplateUserInputs are a template's run form: its own (aiedit_prompts.user_inputs) or its chain's. Values of a
// user input reach the messages as they are; readBacklinks also has the task read the pages that link here.
func TemplateUserInputs(p models.AIEditPrompt) []prompts.Input {
	var own []prompts.Input
	if strings.TrimSpace(p.UserInputs) != "" && json.Unmarshal([]byte(p.UserInputs), &own) == nil {
		if own == nil {
			own = []prompts.Input{}
		}
		return own
	}
	return ChainOf(p.RequestType).Spec.UserInputs
}

// TemplateChain is the chain a template runs as: its request type's chain with the template's user inputs and,
// for the first step, its messages.
func TemplateChain(p models.AIEditPrompt) (prompts.Chain, error) {
	chain, err := ChainOf(p.RequestType).WithUserInputs(TemplateUserInputs(p))
	if err != nil {
		return chain, err
	}
	chain.Spec.Steps[0].Messages = TemplateMessages(p)
	return chain, nil
}

// TemplateMessages are the messages a template runs with: its own (aiedit_prompts.messages) or, for a template
// saved before it had them, its content in place of the chain step's unconditional messages (the instructions
// and the page) followed by the step's messages that carry the run form's values (If, Each).
func TemplateMessages(p models.AIEditPrompt) []prompts.Message {
	var own []prompts.Message
	if strings.TrimSpace(p.Messages) != "" && json.Unmarshal([]byte(p.Messages), &own) == nil && len(own) > 0 {
		return own
	}
	messages := []prompts.Message{{Role: "user", Content: p.Content}}
	for _, m := range ChainOf(p.RequestType).Spec.Steps[0].Messages {
		if m.If != "" || m.Each != "" {
			messages = append(messages, m)
		}
	}
	return messages
}

// MessagesText joins rendered messages for people to read (aiedit_tasks.llm_input).
func MessagesText(messages []prompts.Message) string {
	parts := make([]string, len(messages))
	for i, m := range messages {
		parts[i] = m.Content
	}
	return strings.Join(parts, "\n\n")
}
