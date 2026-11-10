package prompts

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strconv"
	"strings"
	"testing"
)

// specVersions pins each built-in chain's version to a hash of its spec, so a spec change fails here until the
// version is bumped and the entry updated: a minor bump keeps stored results, a major bump makes them stale.
var specVersions = map[string]string{
	"doc-define": "v2.3 d4e35d9cdf90",
	"doc-eval":   "v2.3 98e1cd44b79c",
	"doc-create": "v2.5 b985a5f59f81",
	"doc-update": "v2.5 22cb61eb7137",
}

func specHash(c Chain) string {
	b, err := json.Marshal([]any{c.UserInputs, c.Inputs, c.Steps})
	if err != nil {
		panic(err)
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])[:12]
}

func TestSpecVersions(t *testing.T) {
	for _, c := range BuiltinChains() {
		got := c.Version + " " + specHash(c)
		want := specVersions[c.Name]
		switch {
		case got == want:
		case strings.HasPrefix(want, c.Version+" "):
			t.Errorf("%s: spec changed but version is still %s; bump version (minor keeps stored results, major makes them stale), then update specVersions", c.Name, c.Version)
		default:
			t.Errorf("%s: set specVersions[%q] = %q", c.Name, c.Name, got)
		}
	}
}

func system(st Step) string {
	return st.Render(Vars{})[0].Content
}

func TestDocEvalPrompts(t *testing.T) {
	if DocEval.Version == "" {
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

	if got := DocEvalRubric.Render(Vars{"title": "테스트 문서", "definition": "테스트 문서는 시험용 문서이다."}); got[1].Content != "# 변수\n\n<title>\n테스트 문서\n</title>\n<definition>\n테스트 문서는 시험용 문서이다.\n</definition>" {
		t.Errorf("rubric user message rendered incorrectly: %s", got[1].Content)
	}

	got := DocEvalJudge.Render(Vars{
		"title":    "테스트 문서",
		"rubrics":  `{"general_quality":[]}`,
		"wikitext": "== 개요 ==\n테스트 내용" + strings.Repeat("가", 3000),
	})
	user := got[1].Content
	if !strings.Contains(DocEvalJudge.Template, "{{ wikitext[:"+strconv.Itoa(DocEvalJudgeBodyChars)+"] }}") {
		t.Errorf("the judge template must cut the wikitext at DocEvalJudgeBodyChars")
	}
	if want := "<body>\n" + DocEvalJudgeBody("== 개요 ==\n테스트 내용"+strings.Repeat("가", 3000)) + "\n</body>"; !strings.Contains(user, want) {
		t.Errorf("judge must read the first %d characters of the wikitext", DocEvalJudgeBodyChars)
	}
	if !strings.Contains(user, "<title>\n테스트 문서\n</title>") || !strings.Contains(user, `{"general_quality":[]}`) || !strings.Contains(user, "테스트 내용") {
		t.Errorf("judge user message rendered incorrectly: %s", user)
	}
}

func TestDocDefinePrompt(t *testing.T) {
	if DocDefine.Version == "" {
		t.Fatal("doc-define needs a version: each definition records it")
	}
	got := DocDefineStep.Render(Vars{"title": "자바", "wikitext": "'''자바'''는 인도네시아의 섬이다."})
	if !strings.Contains(got[1].Content, "<title>\n자바\n</title>\n<bodyHead>\n") || !strings.Contains(got[1].Content, "인도네시아의 섬") {
		t.Errorf("define user message rendered incorrectly: %s", got[1].Content)
	}
	if !strings.Contains(system(DocDefineStep), `"definition"`) {
		t.Error("define system message must ask for the definition JSON key")
	}
}
