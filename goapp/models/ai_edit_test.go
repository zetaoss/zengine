package models

import "testing"

func TestNormalizeAIEditRequestType(t *testing.T) {
	for in, want := range map[string]string{"create": AIEditRequestCreate, " update ": AIEditRequestUpdate, "edit": AIEditRequestUpdate} {
		if got, ok := NormalizeAIEditRequestType(in); !ok || got != want {
			t.Errorf("NormalizeAIEditRequestType(%q) = %q, %v; want %q", in, got, ok, want)
		}
	}
	for _, in := range []string{"", "delete", "Edit"} {
		if _, ok := NormalizeAIEditRequestType(in); ok {
			t.Errorf("NormalizeAIEditRequestType(%q) must be rejected", in)
		}
	}
}
