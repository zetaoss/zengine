package redirect

import (
	"testing"
)

func TestNewFixerTask(t *testing.T) {
	task := NewFixerTask()
	if task == nil {
		t.Fatal("expected non-nil FixerTask")
	}
}
