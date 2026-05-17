package runboxjob

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

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/app/job"
)

type payload struct {
	Hash string `json:"hash"`
}

type RunboxJob struct{}

const (
	runboxJobName    = "runbox"
	runboxJobTimeout = 5 * time.Minute
	runboxJobQueue   = "runbox"
)

func NewRunboxJob() *RunboxJob {
	return &RunboxJob{}
}

func Enqueue(ctx context.Context, jobCtx job.JobContext, hash string) (uint64, error) {
	return (&RunboxJob{}).Enqueue(ctx, jobCtx, hash)
}

func (j *RunboxJob) Name() string { return runboxJobName }

func (j *RunboxJob) Timeout() time.Duration { return runboxJobTimeout }

func (j *RunboxJob) Enqueue(ctx context.Context, jobCtx job.JobContext, hash string) (uint64, error) {
	return jobCtx.Enqueue(ctx, job.Request{
		JobName: runboxJobName,
		Queue:   runboxJobQueue,
		Input:   map[string]string{"hash": hash},
	})
}

func (j *RunboxJob) Run(ctx context.Context, jobCtx job.JobContext, p payload) job.Result {
	hash := strings.TrimSpace(p.Hash)
	if hash == "" {
		return job.Error(fmt.Errorf("runbox hash is required"))
	}

	db, err := jobCtx.GetDB()
	if err != nil {
		return job.Error(err)
	}

	var row struct {
		Type    string `gorm:"column:type"`
		Payload string `gorm:"column:payload"`
	}
	if err = db.WithContext(ctx).Table("runboxes").Select("type, payload").Where("hash = ?", hash).Take(&row).Error; err != nil {
		return job.Error(err)
	}
	_ = db.WithContext(ctx).Table("runboxes").Where("hash = ?", hash).Updates(app.H{"phase": "running", "updated_at": time.Now()}).Error

	ep := jobCtx.Config().API.RunboxEndpoint
	if ep == "" {
		_ = db.WithContext(ctx).Table("runboxes").Where("hash = ?", hash).Updates(app.H{"phase": "failed", "updated_at": time.Now()}).Error
		return job.Error(fmt.Errorf("RUNBOX_URL is required"))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ep+"/"+row.Type, bytes.NewBufferString(row.Payload))
	if err != nil {
		_ = db.WithContext(ctx).Table("runboxes").Where("hash = ?", hash).Updates(app.H{"phase": "failed", "updated_at": time.Now()}).Error
		return job.Error(err)
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
		_ = db.WithContext(ctx).Table("runboxes").Where("hash = ?", hash).Updates(app.H{"phase": "failed", "updated_at": time.Now()}).Error
		if err == nil {
			err = fmt.Errorf("empty response from runbox")
		}
		return job.Error(err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	bodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		_ = db.WithContext(ctx).Table("runboxes").Where("hash = ?", hash).Updates(app.H{"phase": "failed", "updated_at": time.Now()}).Error
		return job.Error(fmt.Errorf("read runbox response: %w", readErr))
	}
	slog.Info("[runbox-job] response",
		"hash", hash,
		"status", resp.StatusCode,
		"body", string(bodyBytes),
	)

	var data app.H
	if err := json.Unmarshal(bodyBytes, &data); err != nil || resp.StatusCode < 200 || resp.StatusCode >= 300 {
		_ = db.WithContext(ctx).Table("runboxes").Where("hash = ?", hash).Updates(app.H{"phase": "failed", "updated_at": time.Now()}).Error
		if err == nil {
			err = fmt.Errorf("runbox http status=%d", resp.StatusCode)
		}
		return job.Error(err)
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

	_ = db.WithContext(ctx).Table("runboxes").Where("hash = ?", hash).Updates(app.H{
		"cpu":        data["cpu"],
		"mem":        data["mem"],
		"time":       data["time"],
		"outs":       toJSON(outs),
		"phase":      "succeeded",
		"updated_at": time.Now(),
	}).Error

	return job.Success(app.H{"hash": hash, "phase": "succeeded"})
}

func toJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
