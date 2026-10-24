package prompts

// DocDefine writes a one-line definition of what a wiki page is about, from its title and the beginning of
// its wikitext. Doc-eval builds rubrics from it, so a title alone (a homonym, an abbreviation) no longer decides
// the subject. Its metadata "version" is stored with each definition; bump it when the step changes.
var DocDefine = Chain{
	Name:        "doc-define",
	Description: "제목과 본문 앞부분으로 문서가 다루는 대상을 한 줄로 정의한다.",
	Metadata:    map[string]string{"version": "v1"},
	Inputs:      []string{"title", "lead"},
	Steps:       []Step{DocDefineStep},
}

var DocDefineStep = Step{
	Name:        "define",
	Description: "제목과 본문 앞부분(1500바이트까지)을 보고 문서의 대상을 한 문장으로 정의한다.",
	Prompt: Prompt{
		Inputs: []string{"title", "lead"},
		Messages: []Message{
			{Role: "system", Content: docDefineSystem},
			{Role: "user", Content: "### 문서 제목: {title}\n\n### 문서 본문 앞부분(Wikitext):\n{lead}"},
		},
	},
	Parser: ParserJSON,
	Output: "definition",
}

const docDefineSystem = `당신은 지식 위키 '제타위키'의 문서 분류 담당자이다.

## 과제
문서 제목과 본문 앞부분을 보고, 이 문서가 다루는 대상이 무엇인지 한 문장으로 정의한다.
- 제목이 여러 뜻을 가질 수 있으면(동음이의어, 약어 등) 본문이 실제로 다루는 뜻으로 정의한다.
- 본문에 쓰인 내용을 근거로 한다. 본문에 없는 사실을 덧붙이지 않고, 내용이 맞는지는 판단하지 않는다.
- "제목은 무엇이다" 꼴의 한국어 한 문장으로, 80자 안팎으로 쓴다. 위키 문법은 쓰지 않는다.
- 본문이 비었거나 대상을 알 수 없으면 제목만으로 가장 그럴듯한 뜻을 쓴다.

## 출력
코드 블록이나 마크다운 없이 다음 JSON만 출력한다.
{"definition": "정의 문장"}`
