package registry

import (
	"context"
	"testing"
	"time"

	"github.com/zetaoss/zengine/goapp/app/taskctx"
)

func TestLoopOptionWithLooperInterface(t *testing.T) {
	r := &Registry{specs: make(map[string]Spec)}
	register(r, "dummy-loop", time.Minute, dummyLooperTask{}, loop())

	spec, ok := r.FindSpec("dummy-loop")
	if !ok {
		t.Fatal("expected dummy-loop to be registered")
	}
	if spec.Loop == nil {
		t.Fatal("expected LoopConfig to be set")
	}
	if spec.MaxRetries != 0 {
		t.Fatalf("expected MaxRetries = 0, got %d", spec.MaxRetries)
	}

	d, active, err := spec.Loop.Interval(context.Background(), nil)
	if err != nil || !active || d != 15*time.Minute {
		t.Fatalf("unexpected interval result: d=%v, active=%v, err=%v", d, active, err)
	}
}

func TestNewLoopNannyTask(t *testing.T) {
	r := New()
	nanny := NewLoopNannyTask(r)
	if nanny == nil {
		t.Fatal("expected non-nil LoopNannyTask")
	}
}

type dummyLooperTask struct{}

func (d dummyLooperTask) Execute(_ context.Context, _ taskctx.Context, _ any) (map[string]any, error) {
	return nil, nil
}

func (d dummyLooperTask) LoopInterval(_ context.Context, _ taskctx.Context) (time.Duration, bool, error) {
	return 15 * time.Minute, true, nil
}
