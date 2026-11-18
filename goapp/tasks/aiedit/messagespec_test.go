package aiedit

import (
	"strings"
	"testing"

	"github.com/zetaoss/zengine/goapp/prompts"
)

func TestMessageSpecCarriesOverTaggedTemplates(t *testing.T) {
	cases := map[string]struct {
		spec   string
		system string
		user   string
	}{
		"as the editor wrote it": {
			spec: "name: 고치기\ndescription: 설명\ntemplate_format: jinja2\ntemplate: |-\n  <message role=\"system\">\n  규칙 {{풀기:사전링크}}\n  둘째 줄\n  </message>\n  <message role=\"user\">\n  {{ title }}\n  </message>\ninput_variables:\n  - name: title\n",
			system: "규칙 {{풀기:사전링크}}\n둘째 줄",
			user:   "{{ title }}",
		},
		"clip block, single quotes, spaces and blank lines": {
			spec:   "name: 고치기\ntemplate: |\n  <message  role = 'system' >  \n  규칙\n  </message>\n\n  <message role='user'>\n  {{ title }}\n\n  </message>\ninput_variables:\n  - name: title\n",
			system: "규칙",
			user:   "{{ title }}\n",
		},
		"quoted template": {
			spec:   "name: 고치기\ntemplate: \"<message role=\\\"system\\\">\\n규칙\\n</message>\\n<message role=\\\"user\\\">\\n{{ title }}\\n</message>\"\ninput_variables:\n  - name: title\n",
			system: "규칙",
			user:   "{{ title }}",
		},
		"user message only": {
			spec: "name: 고치기\ntemplate: |-\n  <message role=\"user\">\n  {{ title }}\n  </message>\ninput_variables:\n  - name: title\n",
			user: "{{ title }}",
		},
	}
	for name, c := range cases {
		got, err := MessageSpec(c.spec)
		if err != nil {
			t.Errorf("%s: %v", name, err)
			continue
		}
		pt, err := prompts.ParsePromptTemplate(got)
		if err != nil {
			t.Errorf("%s: the rewritten spec does not read: %v\n%s", name, err, got)
			continue
		}
		if pt.Name != "고치기" || pt.Message.System != c.system || pt.Message.User != c.user {
			t.Errorf("%s: got %+v\n%s", name, pt, got)
		}
		if len(pt.InputVariables) != 1 || pt.InputVariables[0].Name != "title" {
			t.Errorf("%s: input variables %+v", name, pt.InputVariables)
		}
		if strings.Contains(got, "template") {
			t.Errorf("%s: template and template_format should be gone:\n%s", name, got)
		}
	}
}

func TestMessageSpecRefusesWhatItCannotCarryOver(t *testing.T) {
	wrap := func(template string) string {
		return "name: 고치기\ntemplate: |-\n" + template + "input_variables: []\n"
	}
	cases := map[string]string{
		"two user messages": wrap("  <message role=\"user\">\n  하나\n  </message>\n  <message role=\"user\">\n  둘\n  </message>\n"),
		"assistant message": wrap("  <message role=\"user\">\n  질문\n  </message>\n  <message role=\"assistant\">\n  답\n  </message>\n"),
		"user before system": wrap("  <message role=\"user\">\n  질문\n  </message>\n  <message role=\"system\">\n  규칙\n  </message>\n"),
		"text outside a message": wrap("  규칙\n"),
		"unclosed message":       wrap("  <message role=\"user\">\n  질문\n"),
		"another format":         "name: 고치기\ntemplate_format: handlebars\ntemplate: |-\n  <message role=\"user\">\n  질문\n  </message>\n",
		"already message keys":   "name: 고치기\nmessage:\n  user: 질문\ninput_variables: []\n",
	}
	for name, spec := range cases {
		if got, err := MessageSpec(spec); err == nil {
			t.Errorf("%s: expected an error, got\n%s", name, got)
		}
	}
}
