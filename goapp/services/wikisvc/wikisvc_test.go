package wikisvc

import (
	"strings"
	"testing"
)

func TestLinkPattern(t *testing.T) {
	re := linkPattern("KBS_2TV")
	for _, s := range []string{"[[KBS 2TV]]", "[[kBS_2TV|방송]]", "[[ KBS 2TV#역사 ]]"} {
		if !re.MatchString(s) {
			t.Errorf("expected a link in %q", s)
		}
	}
	for _, s := range []string{"[[KBS 2TV 드라마]]", "[[KBS]]", "KBS 2TV"} {
		if re.MatchString(s) {
			t.Errorf("unexpected link in %q", s)
		}
	}
	if !linkPattern("사과").MatchString("[[사과|과일]]") {
		t.Error("expected a link to a title without letter case")
	}
}

func TestExcerpt(t *testing.T) {
	text := strings.Repeat("가", 400) + "[[사과]]<ref>출처</ref>는 과일<!-- 메모 -->이다." + strings.Repeat("나", 400)
	got := Excerpt(text, linkPattern("사과"))
	if !strings.HasPrefix(got, "…가") || !strings.HasSuffix(got, "나…") || !strings.Contains(got, "[[사과]]는 과일이다.") {
		t.Fatalf("unexpected excerpt %q", got)
	}
	if strings.Contains(got, "출처") || strings.Contains(got, "메모") {
		t.Fatalf("references and comments must be removed: %q", got)
	}
	lead := Excerpt("{{정보}}\n'''배'''는 과일이다.", linkPattern("사과"))
	if lead != "{{정보}}\n'''배'''는 과일이다." {
		t.Fatalf("without a link the beginning stands in, got %q", lead)
	}
}
