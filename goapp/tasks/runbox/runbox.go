package runbox

import (
	"bytes"
	"context"
	cryptorand "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
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
	leaseID, err := newLeaseID()
	if err != nil {
		return nil, fmt.Errorf("create runbox lease: %w", err)
	}
	leaseMarker := toJSON(app.H{"lease_id": leaseID})
	claim := db.WithContext(ctx).Table("runboxes").Where(
		"hash = ? AND phase IN ?",
		hash,
		[]string{"pending", "failed"},
	).Updates(app.H{"phase": "running", "outs": leaseMarker, "updated_at": time.Now()})
	if claim.Error != nil {
		return nil, claim.Error
	}
	if claim.RowsAffected == 0 {
		// A duplicate delivery may observe a job already running or completed.
		// It must not execute the same hash concurrently.
		return app.H{"hash": hash, "phase": "skipped"}, nil
	}

	// fail marks the job failed with a reason fit for the page and logs the error itself.
	fail := func(err error, status int, bobMessage string) {
		slog.Error("[runbox-job] failed", "hash", hash, "status", status, "error", err)
		_ = markFailed(ctx, db, hash, leaseMarker, pageError(err, status, bobMessage))
	}

	bob := taskCtx.Config().API.BobEndpoint
	if bob == "" {
		err := fmt.Errorf("BOB_ENDPOINT is required")
		fail(err, 0, "")
		return nil, err
	}

	ep := bob + "/runbox/" + row.Type
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ep, bytes.NewBufferString(row.Payload))
	if err != nil {
		fail(err, 0, "")
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	slog.Info("[runbox-job] request",
		"hash", hash,
		"type", row.Type,
		"url", ep,
		"payload_bytes", len(row.Payload),
	)

	resp, err := (&http.Client{Timeout: 60 * time.Second}).Do(req)
	if err != nil || resp == nil {
		if err == nil {
			err = fmt.Errorf("empty response from runbox")
		}
		fail(err, 0, "")
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	bodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		err := fmt.Errorf("read runbox response: %w", readErr)
		fail(err, resp.StatusCode, "")
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
		message, _ := data["error"].(string)
		if strings.TrimSpace(message) != "" {
			err = fmt.Errorf("%s", message)
		}
		fail(err, resp.StatusCode, message)
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

	result := db.WithContext(ctx).Table("runboxes").Where("hash = ? AND phase = ? AND outs = ?", hash, "running", leaseMarker).Updates(app.H{
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

// Failure reasons stored in outs.error and shown on the page. They hold no internal details.
const (
	ReasonUnavailable = "runbox unavailable"
	ReasonTimedOut    = "runbox timed out"
)

// pageError is the failure reason shown on the page. Only bob's message for a bad request (4xx) is shown
// as is; other errors may hold internal addresses (bob, its Docker host), so they stay in the worker log.
func pageError(err error, status int, bobMessage string) string {
	if status >= http.StatusBadRequest && status < http.StatusInternalServerError && strings.TrimSpace(bobMessage) != "" {
		return bobMessage
	}
	var ne net.Error
	if errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &ne) && ne.Timeout()) {
		return ReasonTimedOut
	}
	return ReasonUnavailable
}

func markFailed(ctx context.Context, db *gorm.DB, hash, leaseMarker, reason string) error {
	result := db.WithContext(ctx).Table("runboxes").Where("hash = ? AND phase = ? AND outs = ?", hash, "running", leaseMarker).Updates(app.H{
		"phase":      "failed",
		"outs":       toJSON(app.H{"error": reason}),
		"updated_at": time.Now(),
	})
	if result.Error != nil {
		slog.Error("[runbox-job] failed to mark job failed", "hash", hash, "error", result.Error)
	}
	return result.Error
}

func newLeaseID() (string, error) {
	b := make([]byte, 16)
	if _, err := cryptorand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func toJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
