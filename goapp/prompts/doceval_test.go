package prompts

import (
	"strings"
	"testing"
)

func TestDocEvalPrompts(t *testing.T) {
	if DocEval.Metadata["version"] == "" {
		t.Fatal("doc-eval needs a metadata version: evaluations record it and rubric reuse matches on it")
	}

	for _, kw := range []string{"general_quality", "text_quality", "Coherence", "Fluency", "최신성", "길이로 감점"} {
		if !strings.Contains(docEvalRubricSystem, kw) {
			t.Errorf("rubric system message missing expected keyword %q", kw)
		}
	}

	judgeKeywords := []string{"general_quality", "text_quality", "safety", "accuracy", "supported", "contradictory", "no_rad"}
	judgeKeywords = append(judgeKeywords, DocEvalSafetyPolicies...)
	for _, kw := range judgeKeywords {
		if !strings.Contains(docEvalJudgeSystem, kw) {
			t.Errorf("judge system message missing expected keyword %q", kw)
		}
	}

	if got := DocEvalRubric.Prompt.Render(map[string]string{"title": "테스트 문서", "definition": "테스트 문서는 시험용 문서이다."}); got[1].Content != "### 문서 제목: 테스트 문서\n### 대상 정의: 테스트 문서는 시험용 문서이다." {
		t.Errorf("rubric user message rendered incorrectly: %s", got[1].Content)
	}

	got := DocEvalJudge.Prompt.Render(map[string]string{
		"title":    "테스트 문서",
		"rubrics":  `{"general_quality":[]}`,
		"wikitext": "== 개요 ==\n테스트 내용",
	})
	user := got[1].Content
	if !strings.Contains(user, "문서 제목: 테스트 문서") || !strings.Contains(user, `{"general_quality":[]}`) || !strings.Contains(user, "테스트 내용") {
		t.Errorf("judge user message rendered incorrectly: %s", user)
	}
}

func TestDocDefinePrompt(t *testing.T) {
	if DocDefine.Metadata["version"] == "" {
		t.Fatal("doc-define needs a metadata version: each definition records it")
	}
	got := DocDefineStep.Prompt.Render(map[string]string{"title": "자바", "lead": "'''자바'''는 인도네시아의 섬이다."})
	if !strings.Contains(got[1].Content, "문서 제목: 자바") || !strings.Contains(got[1].Content, "인도네시아의 섬") {
		t.Errorf("define user message rendered incorrectly: %s", got[1].Content)
	}
	if !strings.Contains(docDefineSystem, `"definition"`) {
		t.Error("define system message must ask for the definition JSON key")
	}
}
