package prompts

// Steps the services run by themselves, looked up at start so a renamed step fails there.
var (
	DocDefineStep = DocDefine.Step("define")
	DocEvalRubric = DocEval.Step("rubric")
	DocEvalJudge  = DocEval.Step("judge")
)

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
