package aiedit

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/zetaoss/zengine/goapp/models"
)

func TestOriginalWikitext(t *testing.T) {
	tests := []struct {
		name   string
		inputs string
		want   string
	}{
		{name: "body", inputs: `{"body":"[[Page]]"}`, want: "[[Page]]"},
		{name: "legacy wikitext", inputs: `{"wikitext":"old content"}`, want: "old content"},
		{name: "body takes precedence", inputs: `{"body":"current","wikitext":"legacy"}`, want: "current"},
		{name: "missing body", inputs: `{"title":"Page"}`},
		{name: "invalid JSON", inputs: `{`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := originalWikitext(tt.inputs); got != tt.want {
				t.Fatalf("originalWikitext() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTaskResponseExposesOriginalWithoutInputs(t *testing.T) {
	response := newTaskResponse(models.AIEdit{
		Inputs:      `{"body":"[[Page]]","private":"value"}`,
		PageTitle:   "Page",
		RequestType: "update",
		LLMOutput:   "output",
	})

	encoded, err := json.Marshal(response)
	if err != nil {
		t.Fatal(err)
	}
	result := string(encoded)
	if !strings.Contains(result, `"original_wikitext":"[[Page]]"`) {
		t.Fatalf("response does not include the source wikitext: %s", result)
	}
	if strings.Contains(result, "private") || strings.Contains(result, `"inputs"`) {
		t.Fatalf("response exposes stored inputs: %s", result)
	}
}
