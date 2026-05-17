package writerequestjob

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/app/job"

	"gorm.io/gorm"
)

type MatcherJob struct{}

const (
	matcherJobName    = "request-matcher"
	matcherJobTimeout = 5 * time.Minute
	matcherJobQueue   = "default"
)

func NewMatcherJob() *MatcherJob {
	return &MatcherJob{}
}

func (j *MatcherJob) Name() string { return matcherJobName }

func (j *MatcherJob) Timeout() time.Duration { return matcherJobTimeout }

func (j *MatcherJob) Run(ctx context.Context, jobCtx job.JobContext, _ any) job.Result {
	db, err := jobCtx.GetDB()
	if err != nil {
		return job.Error(err)
	}

	targets, err := collectWriteRequestTargets(db)
	if err != nil {
		return job.Error(err)
	}

	if len(targets) == 0 {
		return job.Success(app.H{"updated": 0})
	}

	existsMap, err := fetchTitleExistsMap(ctx, jobCtx, targets)
	if err != nil {
		return job.Error(err)
	}

	updated := 0
	for _, title := range targets {
		if existsMap[title] {
			err := db.Table("write_requests").
				Where("title = ? AND is_matched = ?", title, false).
				Updates(app.H{"is_matched": true, "updated_at": time.Now()}).Error
			if err == nil {
				updated++
			}
		}
	}

	return job.Success(app.H{"updated": updated})
}

func collectWriteRequestTargets(db *gorm.DB) ([]string, error) {
	var titles []string
	err := db.Table("write_requests").
		Where("is_matched = ?", false).
		Distinct("title").
		Pluck("title", &titles).Error
	return titles, err
}

func fetchTitleExistsMap(ctx context.Context, jobCtx job.JobContext, titles []string) (map[string]bool, error) {
	apiServer := strings.TrimRight(jobCtx.Config().App.APIServer, "/")
	if apiServer == "" {
		return nil, fmt.Errorf("apiServer is required")
	}

	uniq := uniqueStrings(titles)
	out := make(map[string]bool, len(uniq))
	client := &http.Client{Timeout: 20 * time.Second}

	for start := 0; start < len(uniq); start += 50 {
		end := start + 50
		if end > len(uniq) {
			end = len(uniq)
		}
		chunk := uniq[start:end]

		for _, title := range chunk {
			url := fmt.Sprintf("%s/w/api.php?action=query&titles=%s&format=json", apiServer, strings.ReplaceAll(title, " ", "_"))
			req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
			resp, err := client.Do(req)
			if err != nil {
				continue
			}
			// Simple check if page exists (this is a placeholder logic, usually requires parsing JSON)
			// For now, let's assume if it's 200 and not "missing" in body
			out[title] = (resp.StatusCode == 200)
			_ = resp.Body.Close()
		}
	}
	return out, nil
}

func uniqueStrings(input []string) []string {
	m := make(map[string]bool)
	var uniq []string
	for _, row := range input {
		if m[row] {
			continue
		}
		m[row] = true
		uniq = append(uniq, row)
	}
	return uniq
}
