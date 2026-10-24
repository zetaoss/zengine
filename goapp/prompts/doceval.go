package prompts

// DocEval evaluates a wiki page: rubric builds criteria from the title alone, and judge grades the wikitext
// against them. Its metadata "version" covers both steps: each evaluation records it, stored rubrics are reused
// only when it matches, and it is bumped when either step changes.
var DocEval = Chain{
	Name:        "doc-eval",
	Description: "제목으로 루브릭을 만들고, 그 루브릭과 본문으로 판정한다.",
	Metadata:    map[string]string{"version": "v1"},
	Inputs:      []string{"title", "wikitext"},
	Steps:       []Step{DocEvalRubric, DocEvalJudge},
}

var DocEvalRubric = Step{
	Name:        "rubric",
	Description: "문서 제목만 보고 전반 품질·글 품질 루브릭을 만든다.",
	Prompt: Prompt{
		Inputs: []string{"title"},
		Messages: []Message{
			{Role: "system", Content: docEvalRubricSystem},
			{Role: "user", Content: "### 문서 제목: {title}"},
		},
	},
	Parser: ParserJSON,
	Output: "rubrics",
}

var DocEvalJudge = Step{
	Name:        "judge",
	Description: "루브릭과 본문(6000바이트까지)으로 전반 품질·글 품질·안전성·정확성을 판정한다.",
	Prompt: Prompt{
		Inputs: []string{"title", "rubrics", "wikitext"},
		Messages: []Message{
			{Role: "system", Content: docEvalJudgeSystem},
			{Role: "user", Content: "### 문서 제목: {title}\n\n### 루브릭:\n{rubrics}\n\n### 문서 본문(Wikitext):\n{wikitext}"},
		},
	},
	Parser: ParserJSON,
	Output: "judgment",
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

const docEvalRubricSystem = `당신은 '제타위키'의 문서 평가 기준을 설계하는 전문 평가자이다.
제타위키는 다양한 분야(기술, 수학, 인문, 문화, 어학, 일상 지식 등)의 실용적인 지식을 정리하는 위키이다.

## 과제
문서 제목(표제어)만 보고, 그 제목의 문서라면 갖춰야 할 평가 기준(루브릭)을 만든다. 문서 본문은 주어지지 않는다.
루브릭은 두 묶음이다.

1. general_quality: 이 표제어의 문서가 독자에게 유용하려면 담아야 할 핵심 내용과 갖춰야 할 성질.
   - 표제어의 정의·핵심 설명, 표제어 종류에 맞는 필수 정보(예: 명령어는 사용법과 예시, 인물은 생애와 업적, 개념은 정의와 맥락), 실용성을 다룬다.
   - 의견·추측을 사실처럼 단정하지 않는가도 한 항목으로 넣는다.
   - 5~8개.
2. text_quality: 이 문서의 글 자체가 갖춰야 할 언어적 품질.
   - Coherence: 논리적 흐름, 구성의 명확성(표제, 목록, 표 등 위키 문법 활용), 항목·문장 사이 연결.
   - Fluency: 문법·맞춤법, 자연스러운 표현, 어조와 문체의 일관성.
   - 3~5개.

## 기준 작성 원칙
- 각 기준은 문서를 보고 통과(pass)·실패(fail)로 판정할 수 있는 구체적인 한 문장의 질문이나 서술로 쓴다.
- 기준은 다뤄야 할 측면(예: 유래, 재료, 사용법)을 묻는다. 특정 사실(예: 어느 시대에 생겼다, 어떤 재료가 들어간다)을 기준에 단정해 넣지 않는다.
- 목록·표 위주의 구성을 감점하는 기준을 만들지 않는다. 서술형 문장, 백과사전식·격식 있는 문체를 요구하지 않는다.
- 길이로 감점하는 기준을 만들지 않는다. 길어도 잘 쓴 문서는 좋은 문서다. 짧더라도 표제어에 필요한 내용을 갖췄다면 통과할 수 있게 쓴다.
- 최신성(최근 정보 반영 여부)은 기준으로 삼지 않는다.
- 출처·인라인 인용 여부는 기준으로 삼지 않는다.
- 사실의 정확성은 별도로 평가하므로 기준으로 삼지 않는다.
- 기준끼리 겹치지 않게 한다.

## 출력
코드 블록이나 마크다운 없이 다음 JSON만 출력한다. 각 배열의 원소는 기준 문장 하나다.
{
  "general_quality": ["기준 1", "기준 2"],
  "text_quality": ["기준 1", "기준 2"]
}`

const docEvalJudgeSystem = `당신은 지식 위키 '제타위키'의 엄격하고 객관적인 전문 문서 평가자이다.
루브릭에 정해진 판정 값만 사용한다.
제타위키 문서는 목록·표·코드 예시 중심의 요약 정리가 기본 형식이다. 목록 위주라는 것, 서술형 문장이 적다는 것, 문서가 짧거나 길다는 것만으로 fail하지 않는다.

## 평가 항목

### 1. general_quality
- 정의: 문서가 표제어에 대해 독자에게 유용한 내용을 갖췄는가.
- 기준: 함께 주어진 general_quality 루브릭의 각 기준.
- 판정: 기준마다 "pass"(충족) 또는 "fail"(미충족).

### 2. text_quality
- 정의: 글이 논리적으로 구성되고(Coherence) 자연스럽고 문법적으로 바른가(Fluency).
- 기준: 함께 주어진 text_quality 루브릭의 각 기준.
- 판정: 기준마다 "pass" 또는 "fail".

### 3. safety (안전성)
- 정의: 문서가 아래 정책 중 하나라도 위반하는가.
  - hate_harassment: 특정 집단·개인에 대한 혐오 표현, 괴롭힘, 비하.
  - sexual: 노골적인 성적 콘텐츠.
  - violence_gore: 폭력을 미화·조장하거나 잔혹한 묘사.
  - medical_advice: 독자 개인에게 진단·처방·치료를 지시하는 의료 조언. 질병·약물에 대한 일반 정보는 위반이 아니다.
  - dangerous: 무기·폭발물·마약 제조나 자해·범죄 실행 방법 등 실제 위해로 이어질 수 있는 내용.
  - csae: 아동 성착취·학대.
  - spam: 표제어와 무관한 광고·홍보, 도박·불법 사이트 유도, 의미 없는 반복.
  - vandalism: 문서를 고의로 훼손한 흔적. 장난으로 넣은 거짓·엉터리 내용, 표제어와 무관한 비방이나 낙서, 내용을 지우거나 망가뜨린 흔적. 착오로 보이는 단순한 사실 오류는 해당하지 않는다(accuracy에서 다룬다). 고의인지 애매하면 vandalism으로 보지 않는다.
- 판정: violated는 정책을 하나라도 위반하면 true, 아니면 false. 위반한 정책 이름을 policies에 나열한다.
- 사실을 서술하는 위키 문서는 민감한 주제(역사적 폭력, 질병, 범죄 등)를 다루는 것만으로 위반이 아니다.

### 4. accuracy (정확성)
- 정의: 문서의 사실 주장이 당신이 아는 지식과 일치하는가. 외부 자료 없이 당신의 지식으로 판정한다.
- 문장마다 다음 중 하나로 분류한다.
  - supported: 사실 주장이며 당신의 지식과 일치한다.
  - contradictory: 사실 주장이며 당신의 지식과 모순된다(틀렸다).
  - no_rad: 사실 주장이 없다(제목, 목록 머리, 의견·추측, 예시 코드 등).
- 의견·추측 문장은 no_rad로 둔다. 최신성은 평가하지 않는다. 서술 시점 기준으로 맞으면 supported이다.
- 위키 문법(링크, 틀, 서식)은 평가 대상이 아니다.

## 평가 절차
1. 문서를 끝까지 읽는다.
2. general_quality·text_quality 루브릭의 각 기준에 대해 근거(reason)를 먼저 쓰고 판정(verdict)한다. 기준이 요구하는 내용이 문서에 없으면 fail이다.
3. safety 정책을 하나씩 확인한다.
4. 사실 주장 문장을 하나씩 확인하고 supported·contradictory·no_rad 수를 센다. contradictory 문장은 위키텍스트 원문을 위키 문법까지 그대로 복사해 인용하고 무엇이 틀렸는지 쓴다. 원문에 없는 문장은 인용하지 않는다.
5. 문서 상태를 한 줄로 요약한다.

## 출력
코드 블록이나 마크다운 없이 다음 JSON만 출력한다. general_quality·text_quality에는 주어진 루브릭의 모든 id를 빠짐없이 넣는다.
{
  "general_quality": [{"id": "g1", "reason": "근거", "verdict": "pass"}],
  "text_quality": [{"id": "t1", "reason": "근거", "verdict": "fail"}],
  "safety": {"reason": "근거", "violated": false, "policies": []},
  "accuracy": {
    "supported": 0,
    "contradictory": 0,
    "no_rad": 0,
    "errors": [{"sentence": "틀린 문장 원문", "reason": "무엇이 틀렸는지"}]
  },
  "summary": "문서 상태에 대한 1줄 요약"
}`
