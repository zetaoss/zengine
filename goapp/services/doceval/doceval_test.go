package doceval

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/zetaoss/zengine/goapp/models"
	"github.com/zetaoss/zengine/goapp/prompts"
	"github.com/zetaoss/zengine/goapp/services/llmsvc"
	"github.com/zetaoss/zengine/goapp/services/llmsvc/client"
)

const testRubrics = `{"general_quality":[{"id":"g1","criterion":"정의"},{"id":"g2","criterion":"예시"},{"id":"g3","criterion":"옵션"},{"id":"g4","criterion":"단정 없음"}],"text_quality":[{"id":"t1","criterion":"흐름"},{"id":"t2","criterion":"문법"}]}`

const testGeneratedRubrics = `{"general_quality":["정의","예시"," ","옵션","단정 없음"],"text_quality":["흐름","문법"]}`

func llmPayload(content string) []byte {
	b, _ := json.Marshal(map[string]any{
		"model":   "gemini/gemini-3-flash-preview",
		"choices": []map[string]any{{"message": map[string]any{"content": content}}},
		"usage":   map[string]any{"prompt_tokens": 100, "completion_tokens": 50, "total_tokens": 150},
	})
	return b
}

// newServer answers rubric requests with rubrics and judge requests with judgment, counting calls.
func newServer(t *testing.T, rubrics, judgment string, calls *atomic.Int32) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		var req struct {
			Messages []client.Message `json:"messages"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(req.Messages[0].Content, "평가 기준(루브릭)을 만든다") {
			_, _ = w.Write(llmPayload(rubrics))
			return
		}
		_, _ = w.Write(llmPayload(judgment))
	}))
}

func newEvaluator(url string) *Evaluator {
	return &Evaluator{llm: llmsvc.NewWithEndpoint(url)}
}

func TestEvaluateGeneratesRubricsAndScores(t *testing.T) {
	judgment := "```json\n" + `{
		"general_quality":[{"id":"g1","reason":"","verdict":"pass"},{"id":"g2","reason":"","verdict":"PASS"},{"id":"g3","reason":"","verdict":"fail"}],
		"text_quality":[{"id":"t1","reason":"","verdict":"pass"},{"id":"t2","reason":"","verdict":"pass"}],
		"safety":{"reason":"","violated":false,"policies":[]},
		"accuracy":{"supported":5,"contradictory":0,"no_rad":3,"errors":[]},
		"summary":"쓸 만한 문서"
	}` + "\n```"
	var calls atomic.Int32
	ts := newServer(t, testGeneratedRubrics, judgment, &calls)
	defer ts.Close()

	res, err := newEvaluator(ts.URL).Evaluate(context.Background(), Page{ID: 123, Title: "리눅스 whoami", Revid: 456}, "리눅스 whoami는 사용자 이름을 출력하는 명령어이다.", "본문", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls.Load() != 2 {
		t.Fatalf("expected 2 llm calls, got %d", calls.Load())
	}
	if res.PageID != 123 || res.PageTitle != "리눅스 whoami" || res.Revid != 456 || !strings.HasPrefix(res.Definition, "리눅스 whoami는") {
		t.Fatalf("unexpected page meta: %+v", res)
	}
	// g4 is missing from the judgment and counts as fail: 2/4.
	if res.GeneralQuality != 50 || res.TextQuality != 100 {
		t.Fatalf("unexpected pass rates: %d / %d", res.GeneralQuality, res.TextQuality)
	}
	// Composite score: 50*0.8 + 100*0.2 = 60, which is grade B.
	if res.Grade != "B" || res.Score != 60 || res.SafetyViolated || res.AccuracyContradicted != 0 {
		t.Fatalf("unexpected gates: %+v", res)
	}
	if res.PromptTokens != 200 || res.CompletionTokens != 100 {
		t.Fatalf("unexpected token usage: %d / %d", res.PromptTokens, res.CompletionTokens)
	}
	if res.Summary != "쓸 만한 문서" || !strings.Contains(res.Rubrics, `{"id":"g4","criterion":"단정 없음"}`) {
		t.Fatalf("unexpected summary/rubrics: %+v", res)
	}
}

func TestEvaluateReusesCachedRubrics(t *testing.T) {
	judgment := `{"general_quality":[],"text_quality":[],"safety":{"violated":false},"accuracy":{"contradictory":3,"errors":[{"sentence":"x","reason":"y"},{"sentence":"z 가","reason":"w"},{"sentence":"없는 문장","reason":"misquote"}]},"summary":""}`
	var calls atomic.Int32
	ts := newServer(t, "not json", judgment, &calls)
	defer ts.Close()

	res, err := newEvaluator(ts.URL).Evaluate(context.Background(), Page{ID: 1, Title: "토막글", Revid: 1}, "토막글은 짧은 문서이다.", "* x\n* z  가", "", testRubrics)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls.Load() != 1 {
		t.Fatalf("expected cached rubrics to skip generation, got %d calls", calls.Load())
	}
	// Contradicted sentences are reported beside the grade, not in it: 0% general quality is still D, not F.
	if res.Grade != "D" || res.AccuracyContradicted != 2 {
		t.Fatalf("unexpected accuracy result: %s / %d", res.Grade, res.AccuracyContradicted)
	}
	if res.GeneralQuality != 0 || res.TextQuality != 0 {
		t.Fatalf("unexpected pass rates: %d / %d", res.GeneralQuality, res.TextQuality)
	}
}

func TestEvaluateSafetyFailWins(t *testing.T) {
	judgment := `{"safety":{"violated":false,"policies":["Spam","unknown","spam"]},"accuracy":{"contradictory":1,"errors":[{"sentence":"내용","reason":"y"}]},"summary":""}`
	var calls atomic.Int32
	ts := newServer(t, testGeneratedRubrics, judgment, &calls)
	defer ts.Close()

	res, err := newEvaluator(ts.URL).Evaluate(context.Background(), Page{ID: 1, Title: "광고", Revid: 1}, "광고의 정의", "내용", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Grade != models.DocEvalGradeFail || !res.SafetyViolated || res.SafetyPolicies != "spam" {
		t.Fatalf("unexpected safety gate: %s / %v / %q", res.Grade, res.SafetyViolated, res.SafetyPolicies)
	}
}

func TestEvaluateInvalidRubrics(t *testing.T) {
	var calls atomic.Int32
	ts := newServer(t, `{"general_quality":[],"text_quality":["흐름"]}`, "{}", &calls)
	defer ts.Close()

	if _, err := newEvaluator(ts.URL).Evaluate(context.Background(), Page{ID: 1, Title: "제목", Revid: 1}, "제목의 정의", "내용", "", ""); err == nil {
		t.Fatal("expected error for empty rubrics")
	}
}

func TestEvaluateRetriesInvalidJSON(t *testing.T) {
	var calls atomic.Int32
	ts := newServer(t, `{"general_quality": ["a": "b"]}`, "{}", &calls)
	defer ts.Close()

	if _, err := newEvaluator(ts.URL).Evaluate(context.Background(), Page{ID: 1, Title: "제목", Revid: 1}, "제목의 정의", "내용", "", ""); err == nil {
		t.Fatal("expected error for invalid rubric json")
	}
	if calls.Load() != 2 {
		t.Fatalf("expected one retry, got %d calls", calls.Load())
	}
}

func TestCompositeScore(t *testing.T) {
	cases := []struct {
		general, text uint8
		want          uint8
	}{
		{100, 100, 100},
		{0, 0, 0},
		{50, 100, 60}, // 50*0.8 + 100*0.2 = 60
		{0, 25, 5},    // 0*0.8 + 25*0.2 = 5
		{80, 90, 82},  // 80*0.8 + 90*0.2 = 64 + 18 = 82
		{70, 75, 71},  // 56 + 15 = 71
	}
	for _, c := range cases {
		if got := models.CompositeScore(c.general, c.text); got != c.want {
			t.Errorf("CompositeScore(%d, %d) = %d, want %d", c.general, c.text, got, c.want)
		}
	}
}

func TestGrade(t *testing.T) {
	cases := []struct {
		violated bool
		score    uint8
		want     string
	}{
		{true, 100, "F"}, {false, 100, "S"}, {false, 90, "S"}, {false, 86, "A"}, {false, 75, "A"},
		{false, 71, "B"}, {false, 60, "B"}, {false, 57, "C"}, {false, 40, "C"}, {false, 29, "D"}, {false, 0, "D"},
	}
	for _, c := range cases {
		if got := Grade(c.violated, c.score); got != c.want {
			t.Errorf("Grade(%v, %d) = %s, want %s", c.violated, c.score, got, c.want)
		}
	}
}

func TestRubricPromptCarriesDefinition(t *testing.T) {
	var rubricUser string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Messages []client.Message `json:"messages"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		if strings.Contains(req.Messages[0].Content, "평가 기준(루브릭)을 만든다") {
			rubricUser = req.Messages[1].Content
			_, _ = w.Write(llmPayload(testGeneratedRubrics))
			return
		}
		_, _ = w.Write(llmPayload(`{"general_quality":[],"text_quality":[],"safety":{"violated":false},"accuracy":{},"summary":""}`))
	}))
	defer ts.Close()

	if _, err := newEvaluator(ts.URL).Evaluate(context.Background(), Page{ID: 1, Title: "자바_(섬)", Revid: 1}, "자바는 인도네시아의 섬이다.", "본문", "", ""); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(rubricUser, "문서 제목: 자바 (섬)") || !strings.Contains(rubricUser, "대상 정의: 자바는 인도네시아의 섬이다.") {
		t.Fatalf("rubric prompt must carry the display title and the definition, got %q", rubricUser)
	}
}

func TestDefinitionCurrent(t *testing.T) {
	page := Page{ID: 1, Title: "자바", Revid: 10}
	d := &models.DocDefinition{PageID: 1, PageTitle: "자바", Revid: 10, ChainVersion: prompts.DocDefine.Metadata.Version}
	if !definitionCurrent(d, page) {
		t.Fatal("expected a definition of the same revision, title and major version to be current")
	}
	for name, stale := range map[string]Page{"revision": {ID: 1, Title: "자바", Revid: 11}, "title": {ID: 1, Title: "자바_(섬)", Revid: 10}} {
		if definitionCurrent(d, stale) {
			t.Errorf("expected another %s to make the definition stale", name)
		}
	}
	major := prompts.Major(prompts.DocDefine.Metadata.Version)
	for _, v := range []string{major, major + ".0", major + ".99"} {
		same := *d
		same.ChainVersion = v
		if !definitionCurrent(&same, page) {
			t.Errorf("expected a definition of version %s to be current", v)
		}
	}
	old := *d
	old.ChainVersion = "v0.9"
	if definitionCurrent(&old, page) || definitionCurrent(nil, page) {
		t.Error("expected another major version or no definition to be stale")
	}
}

func TestDefineMajorArg(t *testing.T) {
	if arg := defineMajorArg(); arg.Name != "defineMajor" || arg.Value != prompts.Major(prompts.DocDefine.Metadata.Version) {
		t.Fatalf("unexpected arg %+v", arg)
	}
	if !strings.Contains(currentDefinitionSQL, "SUBSTRING_INDEX(d.chain_version, '.', 1) = @defineMajor") {
		t.Fatal("currentDefinitionSQL must compare the major of d.chain_version with @defineMajor")
	}
}

func TestNormalizeDefinition(t *testing.T) {
	for _, c := range []struct{ title, raw, want string }{
		{"김치", "배추 등을 소금에 절여 양념에 버무린 한국의 발효 음식", "배추 등을 소금에 절여 양념에 버무린 한국의 발효 음식"},
		{"파이썬", "파이썬은 귀도 반 로섬이 만든 범용 프로그래밍 언어이다.", "귀도 반 로섬이 만든 범용 프로그래밍 언어"},
		{"서울", "서울은  대한민국의 수도이다.", "대한민국의 수도"},
		{"피타고라스_정리", "피타고라스 정리란 직각삼각형의 세 변 사이의 관계를 나타내는 정리이다", "직각삼각형의 세 변 사이의 관계를 나타내는 정리"},
		{"바다", "지구 표면의 짠물이 모인 넓은 곳. ", "지구 표면의 짠물이 모인 넓은 곳"},
		{"이다", "이다", "이다"},
	} {
		if got := normalizeDefinition(c.title, c.raw); got != c.want {
			t.Errorf("normalizeDefinition(%q, %q) = %q, want %q", c.title, c.raw, got, c.want)
		}
	}
}
