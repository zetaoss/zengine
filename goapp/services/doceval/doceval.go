package doceval

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/zetaoss/zengine/goapp/app/config"
	"github.com/zetaoss/zengine/goapp/models"
	"github.com/zetaoss/zengine/goapp/prompts"
	"github.com/zetaoss/zengine/goapp/services/llmsvc"
	"gorm.io/gorm"
)

const maxWikitextLen = 6000

type Criterion struct {
	ID        string `json:"id"`
	Criterion string `json:"criterion"`
}

type Rubrics struct {
	GeneralQuality []Criterion `json:"general_quality"`
	TextQuality    []Criterion `json:"text_quality"`
}

type CriterionVerdict struct {
	ID      string `json:"id"`
	Reason  string `json:"reason"`
	Verdict string `json:"verdict"`
}

type AccuracyError struct {
	Sentence string `json:"sentence"`
	Reason   string `json:"reason"`
}

type Judgment struct {
	GeneralQuality []CriterionVerdict `json:"general_quality"`
	TextQuality    []CriterionVerdict `json:"text_quality"`
	Safety         struct {
		Reason   string   `json:"reason"`
		Violated bool     `json:"violated"`
		Policies []string `json:"policies"`
	} `json:"safety"`
	Accuracy struct {
		Supported     uint            `json:"supported"`
		Contradictory uint            `json:"contradictory"`
		NoRad         uint            `json:"no_rad"`
		Errors        []AccuracyError `json:"errors"`
	} `json:"accuracy"`
	Summary string `json:"summary"`
}

type Evaluator struct {
	llm *llmsvc.LLMService
}

// New returns an Evaluator whose model calls are recorded in db's llm_calls.
func New(cfg *config.Config, db *gorm.DB) *Evaluator {
	return &Evaluator{llm: llmsvc.New(cfg).WithRecorder(llmsvc.DBRecorder(db))}
}

// Evaluate generates rubrics for the title and definition (unless cachedRubrics holds a valid set) and judges
// the wikitext against them.
func (e *Evaluator) Evaluate(ctx context.Context, page Page, definition, wikitext, model, cachedRubrics string) (*models.DocEvaluation, error) {
	targetModel := modelOrDefault(model)
	title := page.Title

	// title is the stored DB key (Linux_명령어); the LLM reads it as displayed (Linux 명령어).
	displayTitle := strings.ReplaceAll(title, "_", " ")
	start := time.Now()
	var promptTokens, completionTokens int
	run := func(st prompts.Step, vars prompts.Vars, out any) (llmsvc.StepResult, error) {
		res, err := e.llm.RunStep(ctx, targetModel, prompts.DocEval, st, vars, out)
		promptTokens += res.Usage.PromptTokens
		completionTokens += res.Usage.CompletionTokens
		return res, err
	}

	rubrics, ok := parseRubrics(cachedRubrics)
	if !ok {
		var generated struct {
			GeneralQuality []string `json:"general_quality"`
			TextQuality    []string `json:"text_quality"`
		}
		_, err := run(prompts.DocEvalRubric, prompts.Vars{"title": displayTitle, "definition": definition}, &generated)
		if err != nil {
			return nil, fmt.Errorf("llm rubric generation failed: %w", err)
		}
		rubrics = Rubrics{GeneralQuality: toCriteria("g", generated.GeneralQuality), TextQuality: toCriteria("t", generated.TextQuality)}
		if len(rubrics.GeneralQuality) == 0 || len(rubrics.TextQuality) == 0 {
			return nil, fmt.Errorf("llm rubric generation returned empty rubrics")
		}
	}
	rubricsJSON, _ := json.Marshal(rubrics)

	trimmedWikitext := wikitext
	if len(trimmedWikitext) > maxWikitextLen {
		trimmedWikitext = strings.ToValidUTF8(trimmedWikitext[:maxWikitextLen], "")
	}

	var judgment Judgment
	resp, err := run(prompts.DocEvalJudge, prompts.Vars{
		"title":    displayTitle,
		"rubrics":  string(rubricsJSON),
		"wikitext": trimmedWikitext,
	}, &judgment)
	if err != nil {
		return nil, fmt.Errorf("llm evaluation failed: %w", err)
	}
	elapsed := time.Since(start)

	policies := validPolicies(judgment.Safety.Policies)
	safetyViolated := judgment.Safety.Violated || len(policies) > 0
	judgment.Accuracy.Errors = quotedErrors(judgment.Accuracy.Errors, trimmedWikitext)
	contradicted := uint(len(judgment.Accuracy.Errors))
	judgment.Accuracy.Contradictory = contradicted
	judgmentJSON, _ := json.Marshal(judgment)

	generalQuality := passRate(rubrics.GeneralQuality, judgment.GeneralQuality)
	textQuality := passRate(rubrics.TextQuality, judgment.TextQuality)
	score := models.CompositeScore(generalQuality, textQuality)

	return &models.DocEvaluation{
		PageID:               page.ID,
		PageTitle:            title,
		Revid:                page.Revid,
		Definition:           definition,
		Grade:                Grade(safetyViolated, score),
		GeneralQuality:       generalQuality,
		TextQuality:          textQuality,
		Score:                score,
		SafetyViolated:       safetyViolated,
		SafetyPolicies:       strings.Join(policies, ","),
		AccuracyContradicted: contradicted,
		Summary:              truncateRunes(judgment.Summary, 500),
		Rubrics:              string(rubricsJSON),
		ChainVersion:         prompts.DocEval.Metadata.Version,
		RawJSON:              string(judgmentJSON),
		Model:                resp.Model,
		PromptTokens:         uint(promptTokens),
		CompletionTokens:     uint(completionTokens),
		ElapsedMs:            uint(elapsed.Milliseconds()),
		CreatedAt:            time.Now(),
	}, nil
}

// quotedErrors keeps only accuracy errors whose quoted sentence appears in the wikitext, dropping misquotes.
func quotedErrors(errs []AccuracyError, wikitext string) []AccuracyError {
	text := strings.Join(strings.Fields(wikitext), " ")
	var out []AccuracyError
	for _, e := range errs {
		quote := strings.Join(strings.Fields(e.Sentence), " ")
		if quote != "" && strings.Contains(text, quote) {
			out = append(out, e)
		}
	}
	return out
}

func toCriteria(prefix string, items []string) []Criterion {
	var out []Criterion
	for _, item := range items {
		if item = strings.TrimSpace(item); item != "" {
			out = append(out, Criterion{ID: fmt.Sprintf("%s%d", prefix, len(out)+1), Criterion: item})
		}
	}
	return out
}

// Grade is F for a safety violation, otherwise S, A, B, C, D by the composite score.
func Grade(safetyViolated bool, score uint8) string {
	switch {
	case safetyViolated:
		return models.DocEvalGradeFail
	case score >= 90:
		return "S"
	case score >= 75:
		return "A"
	case score >= 60:
		return "B"
	case score >= 40:
		return "C"
	default:
		return "D"
	}
}

func parseRubrics(raw string) (Rubrics, bool) {
	var r Rubrics
	if strings.TrimSpace(raw) == "" || json.Unmarshal([]byte(raw), &r) != nil {
		return Rubrics{}, false
	}
	return r, len(r.GeneralQuality) > 0 && len(r.TextQuality) > 0
}

// passRate returns the percentage of rubric criteria judged "pass". A criterion the judge omitted counts as fail.
func passRate(criteria []Criterion, verdicts []CriterionVerdict) uint8 {
	if len(criteria) == 0 {
		return 0
	}
	passed := map[string]bool{}
	for _, v := range verdicts {
		passed[v.ID] = strings.EqualFold(strings.TrimSpace(v.Verdict), "pass")
	}
	n := 0
	for _, c := range criteria {
		if passed[c.ID] {
			n++
		}
	}
	return uint8((n*100 + len(criteria)/2) / len(criteria))
}

func validPolicies(policies []string) []string {
	var out []string
	for _, p := range policies {
		p = strings.ToLower(strings.TrimSpace(p))
		if slices.Contains(prompts.DocEvalSafetyPolicies, p) && !slices.Contains(out, p) {
			out = append(out, p)
		}
	}
	return out
}

func truncateRunes(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n])
}
