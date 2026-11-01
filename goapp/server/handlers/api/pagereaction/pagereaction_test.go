package pagereaction

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/zetaoss/zengine/goapp/models"
	"github.com/zetaoss/zengine/goapp/server/serverctx"
)

func TestEmojiCodeConversion(t *testing.T) {
	t.Parallel()

	cases := []struct {
		inputEmoji string
		wantCode   int
		wantCanonical string
	}{
		{"👍", 128077, "👍"},
		{"❤️", 10084, "❤️"},
		{"❤", 10084, "❤️"},
		{"🎉", 127881, "🎉"},
		{"🚀", 128640, "🚀"},
		{"👀", 128064, "👀"},
		{"😄", 128516, "😄"},
		{"😆", 128516, "😄"},
	}

	for _, tc := range cases {
		code := emoji2code(tc.inputEmoji)
		if code != tc.wantCode {
			t.Errorf("emoji2code(%q) = %d, want %d", tc.inputEmoji, code, tc.wantCode)
		}
		canonical := code2emoji(code)
		if canonical != tc.wantCanonical {
			t.Errorf("code2emoji(%d) = %q, want %q", code, canonical, tc.wantCanonical)
		}
	}

	// Unknown emoji should return 0
	if code := emoji2code("unknown"); code != 0 {
		t.Errorf("emoji2code(unknown) = %d, want 0", code)
	}

	// Unknown code should return empty string
	if em := code2emoji(999999); em != "" {
		t.Errorf("code2emoji(999999) = %q, want empty string", em)
	}
}

func TestShowInvalidPage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		path string
	}{
		{"zero page", "0"},
		{"negative page", "-5"},
		{"non-numeric page", "abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/api/reactions/page/"+tt.path, nil)
			req.SetPathValue("page", tt.path)

			c := &serverctx.Context{W: w, R: req}
			Show(c)

			if w.Code != http.StatusNotFound {
				t.Fatalf("Show(%q) status = %d, want %d", tt.path, w.Code, http.StatusNotFound)
			}
		})
	}
}

func TestStoreValidation(t *testing.T) {
	t.Parallel()

	t.Run("unauthenticated user returns 401", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/reactions/page", nil)

		c := &serverctx.Context{W: w, R: req}
		Store(c)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("Store() without user status = %d, want 401", w.Code)
		}
	})

	t.Run("blocked user returns 403", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/reactions/page", nil)

		c := &serverctx.Context{W: w, R: req}
		blockID := 1
		c.SetUser(models.MWUser{ID: 42, BlockID: &blockID})
		Store(c)

		if w.Code != http.StatusForbidden {
			t.Fatalf("Store() blocked user status = %d, want 403", w.Code)
		}
	})

	t.Run("invalid payload returns 422", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/reactions/page", bytes.NewBufferString("{invalid json}"))

		c := &serverctx.Context{W: w, R: req}
		c.SetUser(models.MWUser{ID: 42})
		Store(c)

		if w.Code != http.StatusUnprocessableEntity {
			t.Fatalf("Store() invalid payload status = %d, want 422", w.Code)
		}
	})

	t.Run("invalid emoji returns 422", func(t *testing.T) {
		payload, _ := json.Marshal(map[string]any{
			"pageid": 100,
			"emoji":  "👻",
			"enable": true,
		})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/reactions/page", bytes.NewBuffer(payload))

		c := &serverctx.Context{W: w, R: req}
		c.SetUser(models.MWUser{ID: 42})
		Store(c)

		if w.Code != http.StatusUnprocessableEntity {
			t.Fatalf("Store() invalid emoji status = %d, want 422", w.Code)
		}
	})
}
