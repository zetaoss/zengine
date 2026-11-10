package jinja

import (
	"strings"
	"testing"
)

func render(t *testing.T, src string, vars map[string]any) string {
	t.Helper()
	var names []string
	for k := range vars {
		names = append(names, k)
	}
	tpl, err := Parse(src, names)
	if err != nil {
		t.Fatalf("Parse(%q): %v", src, err)
	}
	out, err := tpl.Render(vars)
	if err != nil {
		t.Fatalf("Render(%q): %v", src, err)
	}
	return out
}

func TestRender(t *testing.T) {
	vars := map[string]any{
		"title":     "자바 (섬)",
		"hint":      "island",
		"blank":     "  ",
		"on":        true,
		"off":       false,
		"texts":     []string{"첫째", "둘째"},
		"none":      []string{},
		"wikitext":  "가나다라마바사",
		"backlinks": []string{"인도네시아", "자카르타"},
	}
	cases := []struct{ name, src, want string }{
		{"output", "<title>\n{{ title }}\n</title>", "<title>\n자바 (섬)\n</title>"},
		{"no spaces", "{{title}}", "자바 (섬)"},
		{"slice by characters", "{{ wikitext[:3] }}|{{ wikitext[5:] }}|{{ wikitext[-2:] }}", "가나다|바사|바사"},
		{"slice past the end", "{{ wikitext[:100] }}", "가나다라마바사"},
		{"list prints lines", "{{ backlinks }}", "인도네시아\n자카르타"},
		{"if", "{% if hint %}H{% endif %}{% if blank %}B{% endif %}", "H"},
		{"if bool", "{% if on %}1{% endif %}{% if off %}2{% endif %}{% if not off %}3{% endif %}", "13"},
		{"if empty list", "{% if none %}x{% else %}empty{% endif %}", "empty"},
		{"elif", "{% if off %}a{% elif blank %}b{% elif hint %}c{% else %}d{% endif %}", "c"},
		{"and or", "{% if off or hint and on %}y{% endif %}{% if off and on %}n{% endif %}", "y"},
		{"for", "{% for t in texts %}[{{ loop.index }}:{{ t }}]{% endfor %}", "[1:첫째][2:둘째]"},
		{"loop first last", "{% for t in texts %}{% if loop.first %}<{% endif %}{{ t }}{% if loop.last %}>{% endif %}{% endfor %}", "<첫째둘째>"},
		{"trim_blocks and lstrip_blocks", "a\n  {% if hint %}\n  <hint>\n  {{ hint }}\n  </hint>\n  {% endif %}\nb", "a\n  <hint>\n  island\n  </hint>\nb"},
		{"block lines vanish when false", "a\n{% if blank %}\nx\n{% endif %}\nb", "a\nb"},
		{"for lines", "<texts>\n{% for t in texts %}\n{{ t }}\n{% endfor %}\n</texts>", "<texts>\n첫째\n둘째\n</texts>"},
		{"minus trims", "a  {{- title -}}  b\n{%- if hint %} c{% endif %}", "a자바 (섬)b c"},
		{"comment", "a{# 메모 #}b\n{# 줄 #}\nc", "ab\nc"},
		{"raw", "{% raw %}{{ title }} {% if x %}{% endraw %}", "{{ title }} {% if x %}"},
		{"raw keeps its body", "a\n{% raw %}\n{{PAGENAME}}\n{% endraw %}\nb", "a\n\n{{PAGENAME}}\nb"},
		{"raw inline", "a\n{% raw %}{{PAGENAME}}{% endraw %}\nb", "a\n{{PAGENAME}}b"},
		{"values are not parsed again", "{{ hint }}", "island"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := render(t, c.src, vars); got != c.want {
				t.Errorf("got %q, want %q", got, c.want)
			}
		})
	}
}

func TestValuesStayAsTheyAre(t *testing.T) {
	v := "{{ title }} {% if x %} <math>a<b</math> &amp; {{틀}}"
	if got := render(t, "{{ body }}", map[string]any{"body": v, "title": "T"}); got != v {
		t.Errorf("got %q, want %q", got, v)
	}
}

func TestWikiSyntaxIsText(t *testing.T) {
	vars := map[string]any{"title": "T"}
	cases := []struct{ src, want string }{
		{"{{풀기:사전링크}}", "{{풀기:사전링크}}"},
		{"틀(`{{…}}`)을 지우지 않는다", "틀(`{{…}}`)을 지우지 않는다"},
		{"{{{1}}} {{{1|기본}}}", "{{{1}}} {{{1|기본}}}"},
		{"{{!}} {{#if:x|y}} {{lc:ABC}}", "{{!}} {{#if:x|y}} {{lc:ABC}}"},
		{"{{Infobox|name=x}}", "{{Infobox|name=x}}"},
		{"{{DISPLAYTITLE:x}}", "{{DISPLAYTITLE:x}}"},
		{"{{틀|{{ title }}}}", "{{틀|T}}"},
		{"{{{ title }}}", "{T}"},
		{"{| class=\"wikitable\"\n|}", "{| class=\"wikitable\"\n|}"},
		{"{\"definition\": \"정의\"}", "{\"definition\": \"정의\"}"},
		{"{%| %} {%", "{%| %} {%"},
		{"}} %} #}", "}} %} #}"},
	}
	for _, c := range cases {
		if got := render(t, c.src, vars); got != c.want {
			t.Errorf("%q: got %q, want %q", c.src, got, c.want)
		}
	}
}

func TestParseErrors(t *testing.T) {
	cases := []struct{ src, want string }{
		{"{{ titel }}", `"titel" is not a variable here`},
		{"{{PAGENAME}}", "{% raw %}"},
		{"{{ title | upper }}", "is not supported"},
		{"{{ title() }}", "is not supported"},
		{"{{ title[1] }}", "name[start:end]"},
		{"{% set x = 1 %}", "{% set %} is not supported"},
		{"{% macro m() %}", "{% macro %} is not supported"},
		{"{% if %}x{% endif %}", "needs a condition"},
		{"{% if title == 'a' %}x{% endif %}", "is not supported"},
		{"{% if title %}x", "not closed by {% endif %}"},
		{"{% endif %}", "unexpected {% endif %}"},
		{"{% for t in texts %}{{ t }}", "not closed by {% endfor %}"},
		{"{% for t in texts %}{% else %}{% endfor %}", "inside {% for %} is not supported"},
		{"{% for title in texts %}{% endfor %}", "hides another variable"},
		{"{% for t of texts %}{% endfor %}", "write {% for item in list %}"},
		{"{{ loop.index }}", `"loop" is not a variable here`},
		{"{% for t in texts %}{{ loop.length }}{% endfor %}", "loop has index"},
		{"{% raw %}x", "not closed by {% endraw %}"},
		{"{# x", "not closed by #}"},
		{"a\n{% if x", "line 2, column 1"},
	}
	for _, c := range cases {
		_, err := Parse(c.src, []string{"title", "texts"})
		if err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%q: got %v, want an error with %q", c.src, err, c.want)
		}
	}
}

func TestUses(t *testing.T) {
	src := "{{ title }}{% if hint %}{{ hint }}{% endif %}{% for t in texts %}{{ t }}{{ loop.index }}{% endfor %}{% raw %}{{ body }}{% endraw %}"
	tpl, err := Parse(src, []string{"title", "hint", "texts", "body"})
	if err != nil {
		t.Fatal(err)
	}
	want := []Use{{"title", UseOutput}, {"hint", UseTest}, {"hint", UseOutput}, {"texts", UseLoop}}
	if got := tpl.Uses(); len(got) != len(want) {
		t.Fatalf("Uses() = %v, want %v", got, want)
	} else {
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("Uses() = %v, want %v", got, want)
			}
		}
	}
	if got := strings.Join(tpl.Variables(), ","); got != "title,hint,texts" {
		t.Errorf("Variables() = %s", got)
	}
}

func TestSpans(t *testing.T) {
	src := "a {{ title }} {{틀}} {% if x %}b{% endif %}{# c #}"
	tpl, err := Parse(src, []string{"title", "x"})
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, s := range tpl.Spans() {
		got = append(got, string(s.Kind)+":"+src[s.Start:s.End])
	}
	want := "output:{{ title }}|tag:{% if x %}|tag:{% endif %}|comment:{# c #}"
	if strings.Join(got, "|") != want {
		t.Errorf("Spans() = %v, want %s", got, want)
	}
}
