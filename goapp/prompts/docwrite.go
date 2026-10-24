package prompts

// DocCreate and DocUpdate write a page's wikitext for AI edit: create for a page that does not exist yet,
// update to rewrite an existing one. Each runs the user's prompt as rendered from an AI edit prompt template
// (aiedit_prompts), so their single step passes it through; the templates are to move into the chains later.
var DocCreate = Chain{
	Name:        "doc-create",
	Description: "새 문서의 위키텍스트를 작성한다.",
	Metadata:    map[string]string{"version": "v1"},
	Inputs:      []string{"prompt"},
	Steps:       []Step{DocCreateGenerate},
}

var DocUpdate = Chain{
	Name:        "doc-update",
	Description: "기존 문서의 위키텍스트를 고쳐 쓴다.",
	Metadata:    map[string]string{"version": "v1"},
	Inputs:      []string{"prompt"},
	Steps:       []Step{DocUpdateGenerate},
}

var DocCreateGenerate = generateStep("AI 편집 프롬프트(생성)로 새 문서의 위키텍스트를 쓴다.")

var DocUpdateGenerate = generateStep("AI 편집 프롬프트(편집)로 기존 문서의 위키텍스트를 고쳐 쓴다.")

func generateStep(description string) Step {
	return Step{
		Name:        "generate",
		Description: description,
		Prompt: Prompt{
			Inputs:   []string{"prompt"},
			Messages: []Message{{Role: "user", Content: "{prompt}"}},
		},
		Parser: ParserText,
		Output: "wikitext",
	}
}
