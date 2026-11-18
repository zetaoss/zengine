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
	"doc-define": "v2.6 a35a00347434",
	"doc-eval":   "v2.6 9bbc9cf0e073",
	"doc-create": "v2.8 2e6750143fa9",
	"doc-update": "v2.7 691c7c5d0014",
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

func user(st Step) string {
	return st.Render(Vars{})[1].Content
}

func TestBuiltinRolesAndInputs(t *testing.T) {
	for _, st := range []Step{DocCreate.PromptStep(), DocUpdate.PromptStep(), DocDefineStep, DocEvalRubric, DocEvalJudge} {
		messages := st.Render(Vars{"title": "제목", "wikitext": "본문", "body": "본문", "definition": "정의", "rubrics": "기준"})
		if len(messages) != 2 || messages[0].Role != "system" || messages[1].Role != "user" {
			t.Fatalf("%s: unexpected roles: %+v", st.Name, messages)
		}
		if strings.Contains(messages[0].Content, "## 과제") || strings.Contains(messages[0].Content, "## 출력") {
			t.Errorf("%s: task instructions must be in the user message", st.Name)
		}
		if !strings.Contains(messages[1].Content, "## 출력") || !strings.Contains(messages[1].Content, "제목") {
			t.Errorf("%s: user message must include input and output instructions", st.Name)
		}
	}
}

func TestDocEvalPrompts(t *testing.T) {
	if DocEval.Version == "" {
		t.Fatal("doc-eval needs a version: evaluations record it and rubric reuse matches on it")
	}

	for _, kw := range []string{"general_quality", "text_quality", "Coherence", "Fluency", "최신성", "길이로 감점"} {
		if !strings.Contains(user(DocEvalRubric), kw) {
			t.Errorf("rubric user message missing expected keyword %q", kw)
		}
	}

	judgeKeywords := []string{"general_quality", "text_quality", "safety", "accuracy", "supported", "contradictory", "no_rad"}
	judgeKeywords = append(judgeKeywords, DocEvalSafetyPolicies...)
	for _, kw := range judgeKeywords {
		if !strings.Contains(user(DocEvalJudge), kw) {
			t.Errorf("judge user message missing expected keyword %q", kw)
		}
	}

	if got := DocEvalRubric.Render(Vars{"title": "테스트 문서", "definition": "테스트 문서는 시험용 문서이다."}); !strings.Contains(got[1].Content, "<definition>\n테스트 문서는 시험용 문서이다.\n</definition>") || !strings.Contains(got[1].Content, "제목이 '테스트 문서'인 문서") {
		t.Errorf("rubric user message rendered incorrectly: %s", got[1].Content)
	}

	got := DocEvalJudge.Render(Vars{
		"title":    "테스트 문서",
		"rubrics":  `{"general_quality":[]}`,
		"wikitext": "== 개요 ==\n테스트 내용" + strings.Repeat("가", 3000),
	})
	user := got[1].Content
	if want := "<body>\n" + "== 개요 ==\n테스트 내용" + strings.Repeat("가", 3000) + "\n</body>"; !strings.Contains(user, want) {
		t.Error("judge must receive the full wikitext")
	}
	if !strings.Contains(user, "제목이 '테스트 문서'인 문서의 본문") || !strings.Contains(user, `{"general_quality":[]}`) || !strings.Contains(user, "테스트 내용") {
		t.Errorf("judge user message rendered incorrectly: %s", user)
	}
}

func TestDocDefinePrompt(t *testing.T) {
	if DocDefine.Version == "" {
		t.Fatal("doc-define needs a version: each definition records it")
	}
	got := DocDefineStep.Render(Vars{"title": "자바", "wikitext": "'''자바'''는 인도네시아의 섬이다."})
	if !strings.Contains(got[1].Content, "## 입력\n\n<bodyHead>\n") || !strings.Contains(got[1].Content, "제목이 '자바'인 문서") || !strings.Contains(got[1].Content, "인도네시아의 섬") {
		t.Errorf("define user message rendered incorrectly: %s", got[1].Content)
	}
	if !strings.Contains(user(DocDefineStep), `"definition"`) {
		t.Error("define user message must ask for the definition JSON key")
	}
}
