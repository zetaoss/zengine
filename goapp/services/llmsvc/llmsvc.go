package llmsvc

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
	"time"

	"github.com/zetaoss/zengine/goapp/app/config"
	"github.com/zetaoss/zengine/goapp/models"
	"github.com/zetaoss/zengine/goapp/prompts"
	"github.com/zetaoss/zengine/goapp/services/llmsvc/client"
	"gorm.io/gorm"
)

type LLMService struct {
	client *client.LLMClient
	record Recorder
}

// Call is one model call RunStep made: OK means the step got a usable answer, so a response that does not
// parse is a failed call with its parse error.
type Call struct {
	Chain         string
	ChainVersion  string
	Step          string
	Model         string
	ResolvedModel string
	OK            bool
	Err           error
	Usage         client.Usage
	Elapsed       time.Duration
}

// Recorder receives each call RunStep makes. It must not fail the step, so it handles its own errors.
type Recorder func(context.Context, Call)

// WithRecorder makes s pass each RunStep call to r and returns s.
func (s *LLMService) WithRecorder(r Recorder) *LLMService {
	s.record = r
	return s
}

// DBRecorder stores each call as an llm_calls row (models.LLMCall); a failed insert is only logged.
func DBRecorder(db *gorm.DB) Recorder {
	return func(ctx context.Context, c Call) {
		row := models.LLMCall{
			Chain:            c.Chain,
			ChainVersion:     c.ChainVersion,
			Step:             c.Step,
			Model:            c.Model,
			ResolvedModel:    c.ResolvedModel,
			OK:               c.OK,
			PromptTokens:     uint(max(c.Usage.PromptTokens, 0)),
			CompletionTokens: uint(max(c.Usage.CompletionTokens, 0)),
			ElapsedMs:        uint(c.Elapsed.Milliseconds()),
			CreatedAt:        time.Now(),
		}
		if c.Err != nil {
			row.Error = truncateRunes(c.Err.Error(), maxCallError)
		}
		// The call is over, so a canceled task context must not drop its record.
		if err := db.WithContext(context.WithoutCancel(ctx)).Create(&row).Error; err != nil {
			slog.Warn("[llmsvc] Failed to record llm call", "chain", c.Chain, "step", c.Step, "model", c.Model, "err", err)
		}
	}
}

// maxCallError matches llm_calls.error.
const maxCallError = 299

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
// call fails, its error keeps the first attempt's parse error. Each call goes to the recorder, if any, under
// the chain's name and version.
func (s *LLMService) RunStep(ctx context.Context, model string, chain prompts.Chain, st prompts.Step, vars map[string]string, out any) (StepResult, error) {
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
	record := func(start time.Time, resp client.ChatResponse, callErr error) {
		if s.record != nil {
			s.record(ctx, Call{
				Chain: chain.Name, ChainVersion: chain.Metadata["version"], Step: st.Name,
				Model: model, ResolvedModel: resp.Model, OK: callErr == nil, Err: callErr,
				Usage: resp.Usage, Elapsed: time.Since(start),
			})
		}
	}
	for range 2 {
		start := time.Now()
		resp, callErr := s.client.ChatCompletionFull(ctx, model, messages)
		if callErr != nil {
			record(start, resp, callErr)
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
			record(start, resp, nil)
			target.Elem().Set(fresh.Elem())
			return res, nil
		}
		err = fmt.Errorf("invalid %s output: %w (raw: %s)", st.Parser, err, truncateRunes(resp.Content, maxRawInError))
		record(start, resp, err)
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
