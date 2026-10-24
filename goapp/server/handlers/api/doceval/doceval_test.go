package doceval

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/zetaoss/zengine/goapp/models"
)

func TestDocEvaluationSerialization(t *testing.T) {
	eval := models.DocEvaluation{
		ID:                   1,
		PageID:               6276,
		PageTitle:            "파이썬",
		Revid:                845848,
		Grade:                models.DocEvalGradeFail,
		GeneralQuality:       83,
		TextQuality:          100,
		SafetyViolated:       true,
		SafetyPolicies:       "spam",
		AccuracyContradicted: 0,
		Summary:              "광고 링크 포함",
		Rubrics:              "{}",
		ChainVersion:         "v1",
		RawJSON:              "{}",
		Model:                "gemini/gemini-3-flash-preview",
		PromptTokens:         1000,
		CompletionTokens:     250,
		ElapsedMs:            5000,
		CreatedAt:            time.Now(),
		LatestRevid:          845850,
		IsLatest:             false,
		RevisionsAfter:       2,
	}

	b, err := json.Marshal(eval)
	if err != nil {
		t.Fatalf("failed to marshal DocEvaluation: %v", err)
	}

	var decoded models.DocEvaluation
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("failed to unmarshal DocEvaluation: %v", err)
	}

	if decoded.Grade != models.DocEvalGradeFail || !decoded.SafetyViolated || decoded.GeneralQuality != 83 || decoded.PageTitle != "파이썬" {
		t.Fatalf("unexpected decoded value: %+v", decoded)
	}
	if decoded.LatestRevid != 845850 || decoded.IsLatest || decoded.RevisionsAfter != 2 {
		t.Fatalf("unexpected revision fields: %+v", decoded)
	}
}

func TestPopulateRevisionsInfoNilSafe(t *testing.T) {
	items := []models.DocEvaluation{
		{ID: 1, PageID: 6276, Revid: 845848},
	}
	populateRevisionsInfo(nil, items)
	if items[0].LatestRevid != 0 || items[0].IsLatest {
		t.Fatalf("expected untouched items on nil db, got %+v", items[0])
	}

	populateRevisionsInfo(nil, nil)
}
