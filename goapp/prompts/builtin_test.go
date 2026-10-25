package prompts

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"
)

// specVersions pins each built-in chain's version to a hash of its spec, so a spec change fails here until the
// version is bumped and the entry updated: a minor bump keeps stored results, a major bump makes them stale.
var specVersions = map[string]string{
	"doc-define": "v2.0 95600152bd0d",
	"doc-eval":   "v2.0 3e49e3826803",
	"doc-create": "v2.0 2e111e3e7d5b",
	"doc-update": "v2.0 9fb3d325d755",
}

func specHash(c Chain) string {
	b, err := json.Marshal(c.Spec)
	if err != nil {
		panic(err)
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])[:12]
}

func TestSpecVersions(t *testing.T) {
	for _, c := range BuiltinChains() {
		got := c.Metadata.Version + " " + specHash(c)
		want := specVersions[c.Name]
		switch {
		case got == want:
		case strings.HasPrefix(want, c.Metadata.Version+" "):
			t.Errorf("%s: spec changed but version is still %s; bump metadata.version (minor keeps stored results, major makes them stale), then update specVersions", c.Name, c.Metadata.Version)
		default:
			t.Errorf("%s: set specVersions[%q] = %q", c.Name, c.Name, got)
		}
	}
}

func system(st Step) string {
	return st.Messages[0].Content
}

func TestDocEvalPrompts(t *testing.T) {
	if DocEval.Metadata.Version == "" {
		t.Fatal("doc-eval needs a version: evaluations record it and rubric reuse matches on it")
	}

	for _, kw := range []string{"general_quality", "text_quality", "Coherence", "Fluency", "최신성", "길이로 감점"} {
		if !strings.Contains(system(DocEvalRubric), kw) {
			t.Errorf("rubric system message missing expected keyword %q", kw)
		}
	}

	judgeKeywords := []string{"general_quality", "text_quality", "safety", "accuracy", "supported", "contradictory", "no_rad"}
	judgeKeywords = append(judgeKeywords, DocEvalSafetyPolicies...)
	for _, kw := range judgeKeywords {
		if !strings.Contains(system(DocEvalJudge), kw) {
			t.Errorf("judge system message missing expected keyword %q", kw)
		}
	}

	if got := DocEvalRubric.Render(Vars{"title": "테스트 문서", "definition": "테스트 문서는 시험용 문서이다."}); got[1].Content != "### 문서 제목: 테스트 문서\n### 대상 정의: 테스트 문서는 시험용 문서이다." {
		t.Errorf("rubric user message rendered incorrectly: %s", got[1].Content)
	}

	got := DocEvalJudge.Render(Vars{
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
	if DocDefine.Metadata.Version == "" {
		t.Fatal("doc-define needs a version: each definition records it")
	}
	got := DocDefineStep.Render(Vars{"title": "자바", "lead": "'''자바'''는 인도네시아의 섬이다."})
	if !strings.Contains(got[1].Content, "문서 제목: 자바") || !strings.Contains(got[1].Content, "인도네시아의 섬") {
		t.Errorf("define user message rendered incorrectly: %s", got[1].Content)
	}
	if !strings.Contains(system(DocDefineStep), `"definition"`) {
		t.Error("define system message must ask for the definition JSON key")
	}
}
