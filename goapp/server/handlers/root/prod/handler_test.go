package prod

import "testing"

func TestShouldServeInjectedIndex(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		cleanPath string
		exists    bool
		want      bool
	}{
		{
			name:      "root always injected",
			cleanPath: "/",
			exists:    true,
			want:      true,
		},
		{
			name:      "existing static file served by file server",
			cleanPath: "/assets/app.js",
			exists:    true,
			want:      false,
		},
		{
			name:      "missing path falls back to injected index",
			cleanPath: "/missing/path",
			exists:    false,
			want:      true,
		},
		{
			name:      "missing dotted route falls back to injected index",
			cleanPath: "/user/alice.smith",
			exists:    false,
			want:      true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := shouldServeInjectedIndex(tc.cleanPath, tc.exists)
			if got != tc.want {
				t.Fatalf("shouldServeInjectedIndex(%q, %v) = %v, want %v", tc.cleanPath, tc.exists, got, tc.want)
			}
		})
	}
}
