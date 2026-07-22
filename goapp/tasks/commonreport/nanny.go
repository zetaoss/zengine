package commonreport

import (
	"context"
	"time"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/app/taskctx"
	"github.com/zetaoss/zengine/goapp/models"
)

const (
	nannyTaskType    = "common-report-nanny"
	nannyTaskTimeout = 1 * time.Minute
	batchSize        = 100
)

type NannyTask struct{}

func NewNannyTask() *NannyTask {
	return &NannyTask{}
}

func (j *NannyTask) Execute(ctx context.Context, taskCtx taskctx.Context, _ any) (app.H, error) {
	db, err := taskCtx.GetDB()
	if err != nil {
		return nil, err
	}

	now := time.Now()

	candidates := make([]models.CommonReport, 0, batchSize)
	if err := db.WithContext(ctx).Table("common_reports").
		Where("phase IN ?", []models.CommonReportPhase{models.CommonReportPhasePending, models.CommonReportPhaseRunning}).
		Order("id").
		Limit(batchSize).
		Find(&candidates).Error; err != nil {
		return nil, err
	}

	enqueued := 0
	for _, task := range candidates {
		if !isReadyToEnqueue(task, now) {
			continue
		}

		_, err := Enqueue(ctx, taskCtx, task.ID)
		if err != nil {
			return nil, err
		}
		enqueued++
	}

	return app.H{
		"enqueued":   enqueued,
		"candidates": len(candidates),
	}, nil
}

func isReadyToEnqueue(task models.CommonReport, now time.Time) bool {
	updatedAt, err := time.Parse(time.RFC3339, task.UpdatedAt)
	if err != nil {
		return true
	}

	threshold := 1 * time.Minute // Default buffer for pending
	if task.Phase == models.CommonReportPhaseRunning {
		threshold = commonReportTaskTimeout + 1*time.Minute
	}

	return now.After(updatedAt.Add(threshold))
}
