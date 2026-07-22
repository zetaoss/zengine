package writerequest

import (
	"context"
	"time"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/app/taskctx"
)

type PrunerTask struct{}

func NewPrunerTask() *PrunerTask { return &PrunerTask{} }

func (j *PrunerTask) Execute(ctx context.Context, taskCtx taskctx.Context, _ any) (app.H, error) {
	db, err := taskCtx.GetDB()
	if err != nil {
		return nil, err
	}

	res := db.WithContext(ctx).Table("not_matches").
		Where("hit <= ? AND updated_at <= ?", 1, time.Now().AddDate(-1, 0, 0)).
		Delete(nil)
	if res.Error != nil {
		return nil, res.Error
	}
	return app.H{"deleted": res.RowsAffected}, nil
}
