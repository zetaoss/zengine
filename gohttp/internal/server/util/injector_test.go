package util

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestInjectScript(t *testing.T) {
	t.Parallel()

	inj := NewInjector(`"gaMeasurementId":"G-AAAA"`)
	req := httptest.NewRequest("GET", "http://example.test/", nil)
	req.Header.Set("X-Policy", "standard")

	html := []byte("<html><head><title>Home</title></head><body>ok</body></html>")
	got := string(inj.InjectScript(html, req))

	if !strings.Contains(got, `window.ZCONF={"gaMeasurementId":"G-AAAA","policy":"standard"};`) {
		t.Fatalf("injected payload missing or malformed: %s", got)
	}
	if !strings.Contains(got, `</script></head>`) {
		t.Fatalf("script not injected at title boundary: %s", got)
	}
}

func TestInjectScriptWithoutTitle(t *testing.T) {
	t.Parallel()

	inj := NewInjector(`"x":"y"`)
	req := httptest.NewRequest("GET", "http://example.test/", nil)

	html := []byte("<html><head></head><body>ok</body></html>")
	got := string(inj.InjectScript(html, req))
	if got != string(html) {
		t.Fatalf("expected unchanged html when </title> is missing, got: %s", got)
	}
}

func TestInjectScriptPolicyFallsBackToStrict(t *testing.T) {
	t.Parallel()

	inj := NewInjector(`"x":"y"`)
	req := httptest.NewRequest("GET", "http://example.test/", nil)
	req.Header.Set("X-Policy", `</script><script>alert(1)</script>`)

	html := []byte("<html><head><title>Home</title></head><body>ok</body></html>")
	got := string(inj.InjectScript(html, req))

	if !strings.Contains(got, `,"policy":"strict"`) {
		t.Fatalf("policy should fall back to strict, got: %s", got)
	}
}
