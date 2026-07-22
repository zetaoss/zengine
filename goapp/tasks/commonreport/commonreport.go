package commonreport

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/hibiken/asynq"
	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/app/taskctx"
	"github.com/zetaoss/zengine/goapp/models"

	"gorm.io/gorm"
)

type payload struct {
	ReportID int `json:"report_id"`
}

type CommonReportTask struct{}

const (
	commonReportTaskType    = "common-report"
	commonReportTaskTimeout = 5 * time.Minute
	commonReportTaskQueue   = "default"
	commonReportUniqueTTL   = 30 * time.Minute
)

func NewCommonReportTask() *CommonReportTask {
	return &CommonReportTask{}
}

func Enqueue(ctx context.Context, taskCtx taskctx.Context, reportID int) (*asynq.TaskInfo, error) {
	raw, err := json.Marshal(payload{ReportID: reportID})
	if err != nil {
		return nil, err
	}
	info, err := taskCtx.EnqueueTask(ctx, asynq.NewTask(commonReportTaskType, raw), asynq.Queue(commonReportTaskQueue), asynq.MaxRetry(3), asynq.Timeout(commonReportTaskTimeout), asynq.Unique(commonReportUniqueTTL))
	if errors.Is(err, asynq.ErrDuplicateTask) {
		return nil, nil
	}
	return info, err
}

func (j *CommonReportTask) Execute(ctx context.Context, taskCtx taskctx.Context, p payload) (app.H, error) {
	if p.ReportID < 1 {
		return nil, fmt.Errorf("common-report report_id is required")
	}

	db, err := taskCtx.GetDB()
	if err != nil {
		return nil, err
	}

	report, err := loadReportWithItems(db.WithContext(ctx), p.ReportID)
	if err != nil {
		slog.Error("common-report Task load failed", "report_id", p.ReportID, "err", err)
		return nil, err
	}
	slog.Info("common-report Task start", "report_id", p.ReportID, "items", len(report.Items))
	_ = db.WithContext(ctx).Table("common_reports").Where("id = ?", p.ReportID).Updates(app.H{"phase": models.CommonReportPhaseRunning, "updated_at": time.Now()}).Error

	if err := processReport(ctx, db.WithContext(ctx), taskCtx.Config().API.SearchEndpoint, report); err != nil {
		slog.Error("common-report Task failed", "report_id", p.ReportID, "err", err)
		_ = db.WithContext(ctx).Table("common_reports").Where("id = ?", p.ReportID).Updates(app.H{"phase": models.CommonReportPhaseFailed, "updated_at": time.Now()}).Error
		return nil, err
	}
	slog.Info("common-report Task succeeded", "report_id", p.ReportID)
	_ = db.WithContext(ctx).Table("common_reports").Where("id = ?", p.ReportID).Updates(app.H{"phase": models.CommonReportPhaseSucceeded, "updated_at": time.Now()}).Error

	return app.H{"report_id": p.ReportID, "phase": models.CommonReportPhaseSucceeded}, nil
}

func processReport(ctx context.Context, db *gorm.DB, endpoint string, report models.CommonReport) error {
	slog.Debug("[processReport]", "endpoint", endpoint)
	ep := strings.TrimSpace(endpoint)
	if ep == "" {
		return fmt.Errorf("SEARCH_ENDPOINT is required")
	}

	q := url.Values{}
	for _, item := range report.Items {
		q.Add("q", item.Name)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ep+"/search?"+q.Encode(), nil)
	if err != nil {
		return err
	}
	resp, err := (&http.Client{Timeout: 30 * time.Second}).Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("search api failed: %d", resp.StatusCode)
	}

	var response app.H
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return err
	}
	result, _ := response["result"].(map[string]any)
	engineRaw, _ := result["engines"].([]any)
	valueRaw, _ := result["values"].([]any)
	if len(engineRaw) == 0 {
		return fmt.Errorf("search payload has no engines")
	}

	engines := make([]string, 0, len(engineRaw))
	for _, e := range engineRaw {
		engines = append(engines, fmt.Sprint(e))
	}
	engineIndex := map[string]int{}
	for i, name := range engines {
		engineIndex[name] = i
	}

	return db.Transaction(func(tx *gorm.DB) error {
		for i, item := range report.Items {
			values := []any{}
			if i < len(valueRaw) {
				if arr, ok := valueRaw[i].([]any); ok {
					values = arr
				}
			}
			get := func(name string) int {
				idx, ok := engineIndex[name]
				if !ok || idx < 0 || idx >= len(values) {
					return 0
				}
				return toInt(values[idx])
			}

			daumBlog := get("daum_blog")
			daumBook := get("daum_book")
			naverBlog := get("naver_blog")
			naverBook := get("naver_book")
			naverNews := get("naver_news")
			googleSearch := get("google_search")
			total := daumBlog + daumBook + naverBlog + naverBook + naverNews + googleSearch

			if err := tx.Table("common_report_items").Where("id = ?", item.ID).Updates(app.H{
				"daum_blog":     daumBlog,
				"daum_book":     daumBook,
				"naver_blog":    naverBlog,
				"naver_book":    naverBook,
				"naver_news":    naverNews,
				"google_search": googleSearch,
				"total":         total,
				"updated_at":    time.Now(),
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func loadReportWithItems(db *gorm.DB, id int) (models.CommonReport, error) {
	var report models.CommonReport
	if err := db.Table("common_reports").Where("id = ?", id).Take(&report).Error; err != nil {
		return models.CommonReport{}, err
	}
	items := make([]models.CommonReportItem, 0, 32)
	if err := db.Table("common_report_items").Where("report_id = ?", id).Order("total DESC").Find(&items).Error; err != nil {
		return models.CommonReport{}, err
	}
	report.Items = items
	return report, nil
}

func toInt(v any) int {
	switch x := v.(type) {
	case int:
		return x
	case int64:
		return int(x)
	case float64:
		return int(x)
	case float32:
		return int(x)
	case string:
		n, _ := strconv.Atoi(strings.TrimSpace(x))
		return n
	default:
		return 0
	}
}
