package doceval

import (
	"encoding/json"
	"testing"
)

func TestRunNowTaskPayload(t *testing.T) {
	task := runNowTask(LoopTaskType, 42, "  Linux 명령어 ")
	if task.Type() != LoopTaskType {
		t.Fatalf("unexpected task type %q", task.Type())
	}
	var p map[string]any
	if err := json.Unmarshal(task.Payload(), &p); err != nil {
		t.Fatal(err)
	}
	if p["force"] != true || p["page_id"] != float64(42) || p["page_title"] != "Linux 명령어" {
		t.Fatalf("unexpected payload %v", p)
	}

	p = nil
	if err := json.Unmarshal(runNowTask(DefineLoopTaskType, 0, " ").Payload(), &p); err != nil {
		t.Fatal(err)
	}
	if len(p) != 1 || p["force"] != true {
		t.Fatalf("expected only force, got %v", p)
	}
}
