package prompts

import (
	"strings"
	"testing"
)

var chatMessages = MessageTemplates{
	System: "지침 {{풀기:사전링크}}",
	User: `## 입력

<title>
{{ title }}
</title>
{% if hint %}
<hint>
{{ hint }}
</hint>
{% endif %}`,
}

func TestParseChat(t *testing.T) {
	ct, err := ParseChat(chatMessages, []string{"title", "hint"})
	if err != nil {
		t.Fatal(err)
	}
	msgs, err := ct.Render(Vars{"title": "눈", "hint": ""})
	if err != nil {
		t.Fatal(err)
	}
	if len(msgs) != 2 {
		t.Fatalf("got %d messages, want 2: %v", len(msgs), msgs)
	}
	if msgs[0].Role != "system" || msgs[0].Content != "지침 {{풀기:사전링크}}" {
		t.Errorf("system = %+v", msgs[0])
	}
	if want := "## 입력\n\n<title>\n눈\n</title>"; msgs[1].Role != "user" || msgs[1].Content != want {
		t.Errorf("user = %q, want %q", msgs[1].Content, want)
	}
}

func TestParseChatBlankMessage(t *testing.T) {
	ct, err := ParseChat(MessageTemplates{User: "{% if hint %}{{ hint }}{% endif %}"}, []string{"hint"})
	if err != nil {
		t.Fatal(err)
	}
	if msgs, _ := ct.Render(Vars{}); len(msgs) != 0 {
		t.Errorf("a message rendered blank must not be sent: %+v", msgs)
	}
}

func TestParseChatErrors(t *testing.T) {
	cases := []struct {
		m    MessageTemplates
		want string
	}{
		{MessageTemplates{}, "no message"},
		{MessageTemplates{System: " \n"}, "no message"},
		{MessageTemplates{User: "a\n{{ titel }}"}, "message.user: line 2, column 1"},
		{MessageTemplates{System: "{{ titel }}", User: "{{ title }}"}, "message.system: line 1"},
	}
	for _, c := range cases {
		_, err := ParseChat(c.m, []string{"title"})
		if err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%+v: got %v, want %q", c.m, err, c.want)
		}
	}
}
