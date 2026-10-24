package llmsvc

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/zetaoss/zengine/goapp/app/config"
	"github.com/zetaoss/zengine/goapp/prompts"
	"github.com/zetaoss/zengine/goapp/services/llmsvc/client"
)

type Input struct {
	Prompt string
	Model  string
}

type Output struct {
	Content string
	Model   string
}

type LLMService struct {
	client *client.LLMClient
}

func New(cfg *config.Config) *LLMService {
	return &LLMService{client: client.New(cfg)}
}

func NewWithEndpoint(endpoint string) *LLMService {
	return &LLMService{client: client.NewClientWithEndpoint(endpoint)}
}

// StepResult is one step run: the model that answered and the token usage summed over its calls.
type StepResult struct {
	Model string
	Usage client.Usage
}

// RunStep renders the step's prompt with vars, calls the model, and parses the response into out, a non-nil
// pointer, with the step's parser. When the response does not parse, it calls once more. Each attempt parses
// into a fresh value and out is set only on success, so a failed attempt leaves nothing behind. When the second
// call fails, its error keeps the first attempt's parse error.
func (s *LLMService) RunStep(ctx context.Context, model string, st prompts.Step, vars map[string]string, out any) (StepResult, error) {
	var res StepResult
	target := reflect.ValueOf(out)
	if target.Kind() != reflect.Pointer || target.IsNil() {
		return res, fmt.Errorf("step %s: out must be a non-nil pointer, got %T", st.Name, out)
	}
	var messages []client.Message
	for _, m := range st.Prompt.Render(vars) {
		messages = append(messages, client.Message{Role: m.Role, Content: m.Content})
	}
	var err error
	for range 2 {
		resp, callErr := s.client.ChatCompletionFull(ctx, model, messages)
		if callErr != nil {
			if err != nil {
				return res, fmt.Errorf("%w (previous attempt: %v)", callErr, err)
			}
			return res, callErr
		}
		res.Model = resp.Model
		res.Usage.PromptTokens += resp.Usage.PromptTokens
		res.Usage.CompletionTokens += resp.Usage.CompletionTokens
		res.Usage.TotalTokens += resp.Usage.TotalTokens
		fresh := reflect.New(target.Elem().Type())
		if err = st.Parser.Parse(resp.Content, fresh.Interface()); err == nil {
			target.Elem().Set(fresh.Elem())
			return res, nil
		}
		err = fmt.Errorf("invalid %s output: %w (raw: %s)", st.Parser, err, truncateRunes(resp.Content, maxRawInError))
	}
	return res, err
}

// maxRawInError caps how much of an unparsable response goes into the error, which ends up in run records.
const maxRawInError = 500

func truncateRunes(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}

func (s *LLMService) Generate(ctx context.Context, in Input) (Output, error) {
	content, model, err := s.client.ChatCompletion(ctx, in.Model, []client.Message{{Role: "user", Content: strings.TrimSpace(in.Prompt)}})
	if err != nil {
		return Output{}, err
	}
	return Output{Content: content, Model: model}, nil
}
