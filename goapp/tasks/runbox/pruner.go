package runbox

import (
	"context"
	"time"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/app/taskctx"
)

type PrunerTask struct{}

const staleRunboxTimeout = 3 * time.Minute

func NewPrunerTask() *PrunerTask {
	return &PrunerTask{}
}

func (j *PrunerTask) Execute(ctx context.Context, taskCtx taskctx.Context, _ any) (app.H, error) {
	db, err := taskCtx.GetDB()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	pendingBefore := now.Add(-staleRunboxTimeout)
	runningBefore := now.Add(-staleRunboxTimeout)
	res := db.WithContext(ctx).Table("runboxes").
		Where("(phase = ? AND updated_at < ?) OR (phase = ? AND updated_at < ?)",
			"pending", pendingBefore,
			"running", runningBefore,
		).
		Updates(app.H{
			"phase":      "failed",
			"updated_at": now,
		})
	if res.Error != nil {
		return nil, res.Error
	}

	return app.H{
		"failed_at":      now.UTC().Format(time.RFC3339),
		"updated":        res.RowsAffected,
		"pending_before": pendingBefore.UTC().Format(time.RFC3339),
		"running_before": runningBefore.UTC().Format(time.RFC3339),
	}, nil
}
