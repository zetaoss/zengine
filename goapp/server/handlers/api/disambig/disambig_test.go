package disambig

import (
	"encoding/json"
	"testing"
	"time"
)

func TestDisambigResponseSerialization(t *testing.T) {
	t.Parallel()

	item := DisambigItem{
		ID:        100,
		Title:     "테스트 문서",
		Entries:   2,
		CreatedAt: time.Date(2026, 8, 4, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 8, 4, 1, 0, 0, 0, time.UTC),
		Nodes: []DisambigNode{
			{Title: "테스트 문서 (1)", Text: "테스트 문서 (1)", Href: "/w/%ED%85%8C%EC%8A%A4%ED%8A%B8_%EB%AC%B8%EC%84%9C_(1)", ID: 101},
			{Title: "테스트 문서 (2)", Text: "테스트 문서 (2)", Href: "/w/%ED%85%8C%EC%8A%A4%ED%8A%B8_%EB%AC%B8%EC%84%9C_(2)", New: 1},
		},
	}

	resp := DisambigResponse{
		Data:       []DisambigItem{item},
		Total:      1,
		Page:       1,
		Limit:      25,
		TotalPages: 1,
	}

	bytes, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal DisambigResponse: %v", err)
	}

	var decoded DisambigResponse
	if err := json.Unmarshal(bytes, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal DisambigResponse: %v", err)
	}

	if decoded.Total != 1 || decoded.Page != 1 || decoded.Limit != 25 || decoded.TotalPages != 1 {
		t.Fatalf("DisambigResponse metadata mismatch: %+v", decoded)
	}

	if len(decoded.Data) != 1 || decoded.Data[0].Title != "테스트 문서" {
		t.Fatalf("DisambigItem title mismatch: %+v", decoded.Data)
	}

	if len(decoded.Data[0].Nodes) != 2 || decoded.Data[0].Nodes[1].New != 1 {
		t.Fatalf("DisambigNode new flag mismatch: %+v", decoded.Data[0].Nodes)
	}
}

func TestPaginationBounds(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		page           int
		limit          int
		total          int64
		wantPage       int
		wantTotalPages int
	}{
		{name: "page in range", page: 3, limit: 25, total: 137, wantPage: 3, wantTotalPages: 6},
		{name: "page above range", page: 999, limit: 25, total: 137, wantPage: 6, wantTotalPages: 6},
		{name: "empty result", page: 999, limit: 25, total: 0, wantPage: 1, wantTotalPages: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			page, totalPages := paginationBounds(tt.page, tt.limit, tt.total)
			if page != tt.wantPage || totalPages != tt.wantTotalPages {
				t.Fatalf("paginationBounds() = (%d, %d), want (%d, %d)", page, totalPages, tt.wantPage, tt.wantTotalPages)
			}
		})
	}
}
