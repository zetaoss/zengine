package aiedit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
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

// Render fills the chain's variables (the user inputs as given; the title and, as body, the wikitext), runs its function
// steps (with useBacklinks, the titles of the pages that link here) and renders its LLM step.
func Render(ctx context.Context, cfg *config.Config, req Request) (Rendered, error) {
	chain := ChainOf(req.RequestType)
	if req.Prompt != nil {
		var err error
		if chain, err = TemplateChain(*req.Prompt); err != nil {
			return Rendered{}, fmt.Errorf("template %d: %w", req.Prompt.ID, err)
		}
	}
	vars, err := chain.UserVars(runInputs(chain, req.UserInputs))
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
	if vars, err = chain.RunFunctions(ctx, vars, functions(cfg)); err != nil {
		return Rendered{}, err
	}
	vars = chain.Fill(vars)
	messages := prompts.MergeUserMessages(chain.PromptStep().Render(vars))
	if len(messages) == 0 {
		return Rendered{}, InputError{"보낼 메시지가 없습니다. 조건(if, each)이 걸린 입력을 채워 주세요."}
	}
	return Rendered{Chain: chain, Vars: vars, Messages: messages}, nil
}

// functions are the code of the function steps AI edit's chains use (prompts.Functions).
func functions(cfg *config.Config) map[string]prompts.Func {
	return map[string]prompts.Func{
		"wiki/backlinkTitles": func(ctx context.Context, vars prompts.Vars) (any, error) {
			title, _ := vars["title"].(string)
			titles, err := wikisvc.BacklinkTitles(ctx, cfg, title)
			if err != nil {
				return nil, fmt.Errorf("read backlinks: %w", err)
			}
			return titles, nil
		},
	}
}

// TemplateUserInputs are a template's run form: the user inputs of its chain it picks by name
// (aiedit_prompts.user_inputs), each as the chain defines it and in the chain's order, or all of the chain's. A stored name from before a
// rename takes its name now; one the chain does not have (request, a template's own input) is left out. Values of a
// user input reach the messages as they are; useBacklinks also has the task collect titles of pages that link here.
func TemplateUserInputs(p models.AIEditPrompt) []prompts.Input {
	var own []prompts.Input
	if strings.TrimSpace(p.UserInputs) == "" || json.Unmarshal([]byte(p.UserInputs), &own) != nil {
		return ChainOf(p.RequestType).Spec.UserInputs
	}
	names := map[string]bool{}
	for _, in := range own {
		names[renamedUserInput(in.Name)] = true
	}
	return inChainOrder(p.RequestType, names)
}

// PickUserInputs reads a template's user inputs as an editor writes them, names of its chain's user inputs alone
// (- hint), and gives each as the chain defines it, in the chain's order (which the messages follow). An input may
// also come as the chain defines it (a skin loaded before templates picked by name sends the whole definition). A
// name the chain does not have, a repeated name, or attributes other than the chain's are refused.
func PickUserInputs(requestType string, given []prompts.Input) ([]prompts.Input, error) {
	defs := userInputDefs(requestType)
	seen := map[string]bool{}
	for _, in := range given {
		name := renamedUserInput(in.Name)
		def, ok := defs[name]
		if !ok {
			return nil, fmt.Errorf("쓸 수 없는 사용자 입력입니다: %s (쓸 수 있는 입력: %s)", in.Name, strings.Join(userInputNames(ChainOf(requestType).Spec.UserInputs), ", "))
		}
		if seen[name] {
			return nil, fmt.Errorf("사용자 입력이 두 번 있습니다: %s", name)
		}
		in.Name = name
		if !reflect.DeepEqual(in, prompts.Input{Name: name}) && !sameInput(in, def) {
			return nil, fmt.Errorf("사용자 입력은 이름만 씁니다: %s", in.Name)
		}
		seen[name] = true
	}
	return inChainOrder(requestType, seen), nil
}

// inChainOrder are the chain's user inputs whose names are given, in the chain's order.
func inChainOrder(requestType string, names map[string]bool) []prompts.Input {
	picked := []prompts.Input{}
	for _, in := range ChainOf(requestType).Spec.UserInputs {
		if names[in.Name] {
			picked = append(picked, in)
		}
	}
	return picked
}

// sameInput reports whether an input as given (decoded from JSON or YAML) is the chain's definition.
func sameInput(given, def prompts.Input) bool {
	a, _ := json.Marshal(given)
	b, _ := json.Marshal(def)
	return string(a) == string(b)
}

// UserInputNames are the inputs by name alone, as a template stores the ones it picks (aiedit_prompts.user_inputs).
func UserInputNames(inputs []prompts.Input) []prompts.Input {
	names := make([]prompts.Input, len(inputs))
	for i, in := range inputs {
		names[i] = prompts.Input{Name: in.Name}
	}
	return names
}

func userInputDefs(requestType string) map[string]prompts.Input {
	defs := map[string]prompts.Input{}
	for _, in := range ChainOf(requestType).Spec.UserInputs {
		defs[in.Name] = in
	}
	return defs
}

func userInputNames(inputs []prompts.Input) []string {
	names := make([]string, len(inputs))
	for i, in := range inputs {
		names[i] = in.Name
	}
	return names
}

func renamedUserInput(name string) string {
	if now, ok := userInputRenames[name]; ok {
		return now
	}
	return name
}

// TemplateChain is the chain a template runs as: its request type's chain with the template's user inputs and,
// for the LLM step, its messages.
func TemplateChain(p models.AIEditPrompt) (prompts.Chain, error) {
	chain, err := ChainOf(p.RequestType).WithUserInputs(TemplateUserInputs(p))
	if err != nil {
		return chain, err
	}
	chain.Spec.Steps[chain.PromptIndex()].Messages = TemplateMessages(p)
	return chain, nil
}

// TemplateMessages are the messages a template runs with: its own (aiedit_prompts.messages) or, for a template
// saved before it had them, its content in place of the chain step's unconditional messages (the instructions
// and the page) followed by the step's messages that carry the run form's values (If, Each).
func TemplateMessages(p models.AIEditPrompt) []prompts.Message {
	var own []prompts.Message
	if strings.TrimSpace(p.Messages) != "" && json.Unmarshal([]byte(p.Messages), &own) == nil && len(own) > 0 {
		for i := range own {
			if name, ok := varRenames[own[i].Each]; ok {
				own[i].Each = name
			}
			if name, ok := varRenames[own[i].If]; ok {
				own[i].If = name
			}
			if name, ok := userInputRenames[own[i].If]; ok {
				own[i].If = name
			}
			own[i].Content = renamedVars(own[i].Content)
		}
		return own
	}
	messages := []prompts.Message{{Role: "user", Content: renamedVars(p.Content)}}
	for _, m := range ChainOf(p.RequestType).PromptStep().Messages {
		if m.If != "" || m.Each != "" {
			messages = append(messages, m)
		}
	}
	return messages
}

// userInputRenames maps user inputs renamed since templates were saved, and since skins still running sent their
// run forms, to their names now.
var userInputRenames = map[string]string{"readBacklinks": "useBacklinks", "extraTexts": "texts", "copiedTexts": "texts"}

// droppedUserInputs are user inputs the built-in chains no longer have (request). A run form loaded before they
// went may still send them; a chain that does not declare one ignores it rather than refusing the run.
var droppedUserInputs = []string{"request"}

// runInputs are the run form values as the chain takes them: values sent under an older user input name get its
// name now (userInputRenames), and dropped inputs the chain does not declare are left out.
func runInputs(chain prompts.Chain, values map[string]any) map[string]any {
	declared := map[string]bool{}
	for _, in := range chain.Spec.UserInputs {
		declared[in.Name] = true
	}
	out := make(map[string]any, len(values))
	for k, v := range values {
		out[k] = v
	}
	for old, name := range userInputRenames {
		if v, ok := out[old]; ok {
			delete(out, old)
			if _, given := out[name]; !given {
				out[name] = v
			}
		}
	}
	for _, name := range droppedUserInputs {
		if !declared[name] {
			delete(out, name)
		}
	}
	return out
}

// varRenames maps variables a stored template may still use under an older name to their names now: the backlink
// output was backlinkTexts and then backlinkTitles, the texts input was extraTexts and then copiedTexts, and
// doc-update's wikitext is now body.
var varRenames = map[string]string{"backlinkTexts": "backlinks", "backlinkTitles": "backlinks", "extraTexts": "texts", "copiedTexts": "texts", "wikitext": "body"}

// renamedVars rewrites the older variable names (varRenames) in a stored template's content: its placeholders
// (prompts.RenamePlaceholders, so wiki markup such as {{wikitext}} stays) and the headings the editor writes above
// them (-----wikitext-----).
func renamedVars(content string) string {
	var pairs []string
	for old, name := range varRenames {
		pairs = append(pairs, "-----"+old+"-----", "-----"+name+"-----")
	}
	return strings.NewReplacer(pairs...).Replace(prompts.RenamePlaceholders(content, varRenames))
}

// MessagesText joins rendered messages for people to read (aiedit_tasks.llm_input).
func MessagesText(messages []prompts.Message) string {
	parts := make([]string, len(messages))
	for i, m := range messages {
		parts[i] = m.Content
	}
	return strings.Join(parts, "\n\n")
}
