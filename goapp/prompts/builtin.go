package prompts

// Steps the services run by themselves, looked up at start so a renamed step fails there.
var (
	DocDefineStep = DocDefine.Step("define")
	DocEvalRubric = DocEval.Step("rubric")
	DocEvalJudge  = DocEval.Step("judge")
)

// DocEvalJudgeBodyChars is how many characters of the wikitext the judge reads ({{ wikitext[:2000] }} in its
// template; TestDocEvalPrompts keeps the two together).
const DocEvalJudgeBodyChars = 2000

// DocEvalJudgeBody is the part of the wikitext the judge reads.
func DocEvalJudgeBody(wikitext string) string {
	if r := []rune(wikitext); len(r) > DocEvalJudgeBodyChars {
		return string(r[:DocEvalJudgeBodyChars])
	}
	return wikitext
}

var DocEvalSafetyPolicies = []string{
	"hate_harassment",
	"sexual",
	"violence_gore",
	"medical_advice",
	"dangerous",
	"csae",
	"spam",
	"vandalism",
}
