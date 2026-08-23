package runbox

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/hibiken/asynq"
	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/app/taskctx"
	"gorm.io/gorm"
)

type payload struct {
	Hash string `json:"hash"`
}

type RunboxTask struct{}

const (
	runboxTaskType    = "runbox"
	runboxTaskTimeout = 2 * time.Minute
	runboxTaskQueue   = "runbox"
)

func NewRunboxTask() *RunboxTask {
	return &RunboxTask{}
}

func Enqueue(ctx context.Context, taskCtx taskctx.Context, hash string) (*asynq.TaskInfo, error) {
	raw, err := json.Marshal(payload{Hash: hash})
	if err != nil {
		return nil, err
	}
	return taskCtx.EnqueueTask(ctx, asynq.NewTask(runboxTaskType, raw), asynq.Queue(runboxTaskQueue), asynq.MaxRetry(0), asynq.Timeout(runboxTaskTimeout))
}

func (j *RunboxTask) Execute(ctx context.Context, taskCtx taskctx.Context, p payload) (app.H, error) {
	hash := strings.TrimSpace(p.Hash)
	if hash == "" {
		return nil, fmt.Errorf("runbox hash is required")
	}

	db, err := taskCtx.GetDB()
	if err != nil {
		return nil, err
	}

	var row struct {
		Type    string `gorm:"column:type"`
		Payload string `gorm:"column:payload"`
	}
	if err = db.WithContext(ctx).Table("runboxes").Select("type, payload").Where("hash = ?", hash).Take(&row).Error; err != nil {
		return nil, err
	}
	// The existing runboxes.updated_at column is second-precision in the
	// deployed database. Normalize the lease timestamp before both writes so
	// the conditional terminal update matches the claimed row.
	claimAt := time.Now().UTC().Truncate(time.Second)
	claim := db.WithContext(ctx).Table("runboxes").Where(
		"hash = ? AND phase IN ?",
		hash,
		[]string{"pending", "failed"},
	).Updates(app.H{"phase": "running", "updated_at": claimAt})
	if claim.Error != nil {
		return nil, claim.Error
	}
	if claim.RowsAffected == 0 {
		// A duplicate delivery may observe a job already running or completed.
		// It must not execute the same hash concurrently.
		return app.H{"hash": hash, "phase": "skipped"}, nil
	}

	ep := taskCtx.Config().API.RunboxEndpoint
	if ep == "" {
		err := fmt.Errorf("RUNBOX_URL is required")
		_ = markFailed(ctx, db, hash, claimAt, err.Error())
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ep+"/"+row.Type, bytes.NewBufferString(row.Payload))
	if err != nil {
		_ = markFailed(ctx, db, hash, claimAt, err.Error())
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	slog.Info("[runbox-job] request",
		"hash", hash,
		"type", row.Type,
		"url", ep+"/"+row.Type,
		"payload_bytes", len(row.Payload),
	)

	resp, err := (&http.Client{Timeout: 60 * time.Second}).Do(req)
	if err != nil || resp == nil {
		if err == nil {
			err = fmt.Errorf("empty response from runbox")
		}
		_ = markFailed(ctx, db, hash, claimAt, err.Error())
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	bodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		err := fmt.Errorf("read runbox response: %w", readErr)
		_ = markFailed(ctx, db, hash, claimAt, err.Error())
		return nil, err
	}
	slog.Info("[runbox-job] response",
		"hash", hash,
		"status", resp.StatusCode,
		"body", string(bodyBytes),
	)

	var data app.H
	decodeErr := json.Unmarshal(bodyBytes, &data)
	if decodeErr != nil || resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err := decodeErr
		if err == nil {
			err = fmt.Errorf("runbox http status=%d", resp.StatusCode)
		}
		if message, ok := data["error"].(string); ok && strings.TrimSpace(message) != "" {
			err = fmt.Errorf("%s", message)
		}
		_ = markFailed(ctx, db, hash, claimAt, err.Error())
		if resp.StatusCode >= http.StatusBadRequest && resp.StatusCode < http.StatusInternalServerError {
			return nil, fmt.Errorf("%w: %v", asynq.SkipRetry, err)
		}
		return nil, err
	}

	outs := app.H{"logs": data["logs"], "images": data["images"]}
	if v, ok := data["outputsList"]; ok {
		outs = app.H{"outputsList": v}
	}
	slog.Info("[runbox-job] parsed-outs",
		"hash", hash,
		"has_logs", data["logs"] != nil,
		"has_images", data["images"] != nil,
		"has_outputs_list", data["outputsList"] != nil,
	)

	result := db.WithContext(ctx).Table("runboxes").Where("hash = ? AND phase = ? AND updated_at = ?", hash, "running", claimAt).Updates(app.H{
		"cpu":        data["cpu"],
		"mem":        data["mem"],
		"time":       data["time"],
		"outs":       toJSON(outs),
		"phase":      "succeeded",
		"updated_at": time.Now(),
	})
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("runbox job was superseded: %s", hash)
	}

	return app.H{"hash": hash, "phase": "succeeded"}, nil
}

func markFailed(ctx context.Context, db *gorm.DB, hash string, claimAt time.Time, reason string) error {
	result := db.WithContext(ctx).Table("runboxes").Where("hash = ? AND phase = ? AND updated_at = ?", hash, "running", claimAt).Updates(app.H{
		"phase":      "failed",
		"outs":       toJSON(app.H{"error": reason}),
		"updated_at": time.Now(),
	})
	if result.Error != nil {
		slog.Error("[runbox-job] failed to mark job failed", "hash", hash, "error", result.Error)
	}
	return result.Error
}

func toJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
