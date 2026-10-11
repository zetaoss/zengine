package writerequest

import (
	"strings"
	"testing"
)

func TestResolveTitleExistsAcceptsSpaceUnderscoreVariants(t *testing.T) {
	existsMap := map[string]bool{"hello world": true, "missing page": false}
	if !resolveTitleExists(existsMap, "Hello_World") {
		t.Fatal("expected underscore title to match space title")
	}
	if resolveTitleExists(existsMap, "Missing_Page") {
		t.Fatal("expected missing page to remain missing")
	}
}

func TestUniqueStringsPreservesOrder(t *testing.T) {
	got := uniqueStrings([]string{"a", "b", "a"})
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("unexpected unique strings: %#v", got)
	}
}

func TestDecodeTitleExists(t *testing.T) {
	body := `{"batchcomplete":true,"query":{"pages":[` +
		`{"ns":0,"title":"깃허브 액션","missing":true},` +
		`{"pageid":7,"ns":0,"title":"데비안"}]}}`
	out := map[string]bool{}
	if err := decodeTitleExists(strings.NewReader(body), out); err != nil {
		t.Fatalf("decodeTitleExists: %v", err)
	}
	if !resolveTitleExists(out, "데비안") {
		t.Fatal("expected 데비안 to exist")
	}
	if resolveTitleExists(out, "깃허브_액션") {
		t.Fatal("expected 깃허브 액션 to be missing")
	}
}

func TestDecodeTitleExistsRequiresPages(t *testing.T) {
	if err := decodeTitleExists(strings.NewReader(`{"batchcomplete":true}`), map[string]bool{}); err == nil {
		t.Fatal("expected an error when query.pages is missing")
	}
}
