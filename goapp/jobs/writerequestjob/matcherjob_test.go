package writerequestjob

import "testing"

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
