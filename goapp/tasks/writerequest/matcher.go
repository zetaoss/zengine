package writerequest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/app/taskctx"

	"gorm.io/gorm"
)

type MatcherTask struct{}

const (
	matcherTaskType    = "request-matcher"
	matcherTaskTimeout = 5 * time.Minute
	matcherPageSize    = 50
)

type writeRequestTarget struct {
	ID       uint       `gorm:"column:id"`
	Title    string     `gorm:"column:title"`
	WritedAt *time.Time `gorm:"column:writed_at"`
}

func NewMatcherTask() *MatcherTask { return &MatcherTask{} }

func (j *MatcherTask) Execute(ctx context.Context, taskCtx taskctx.Context, _ any) (app.H, error) {
	db, err := taskCtx.GetDB()
	if err != nil {
		return nil, err
	}

	targets, err := collectWriteRequestTargets(db.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	if len(targets) == 0 {
		return app.H{"checked": 0, "to_done": 0, "to_todo": 0}, nil
	}

	titles := make([]string, 0, len(targets))
	for _, target := range targets {
		titles = append(titles, target.Title)
	}
	existsMap, err := fetchTitleExistsMap(ctx, taskCtx, titles)
	if err != nil {
		return nil, err
	}

	toDone := make([]uint, 0)
	toTodo := make([]uint, 0)
	checked := make([]uint, 0, len(targets))
	for _, target := range targets {
		checked = append(checked, target.ID)
		exists := resolveTitleExists(existsMap, target.Title)
		if exists && target.WritedAt == nil {
			toDone = append(toDone, target.ID)
		} else if !exists && target.WritedAt != nil {
			toTodo = append(toTodo, target.ID)
		}
	}

	now := time.Now()
	err = db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if len(toDone) > 0 {
			if err := tx.Table("write_requests").Where("id IN ?", toDone).Updates(app.H{
				"writer_id": 0, "writer_name": "Unknown", "writed_at": now, "updated_at": now,
			}).Error; err != nil {
				return err
			}
		}
		if len(toTodo) > 0 {
			if err := tx.Table("write_requests").Where("id IN ?", toTodo).Updates(app.H{
				"writed_at": nil, "updated_at": now,
			}).Error; err != nil {
				return err
			}
		}
		return tx.Table("write_requests").Where("id IN ?", checked).Update("updated_at", now).Error
	})
	if err != nil {
		return nil, err
	}

	return app.H{"checked": len(targets), "to_done": len(toDone), "to_todo": len(toTodo)}, nil
}

func collectWriteRequestTargets(db *gorm.DB) ([]writeRequestTarget, error) {
	queries := []*gorm.DB{
		db.Table("write_requests").Select("id, title, writed_at").Where("writed_at IS NULL").Order("id DESC").Limit(matcherPageSize),
		db.Table("write_requests AS w").Select("w.id, w.title, w.writed_at").Where("w.writed_at IS NULL").
			Order("w.rate DESC").Order("(SELECT COALESCE(n.hit, 0) FROM not_matches n WHERE n.title = w.title LIMIT 1) DESC").
			Order("w.ref DESC").Order("w.id DESC").Limit(matcherPageSize),
		db.Table("write_requests").Select("id, title, writed_at").Where("writed_at IS NOT NULL").Order("writed_at DESC").Order("id DESC").Limit(matcherPageSize),
		db.Table("write_requests").Select("id, title, writed_at").Order("updated_at").Order("id").Limit(matcherPageSize),
	}

	seen := make(map[uint]bool)
	targets := make([]writeRequestTarget, 0, matcherPageSize*len(queries))
	for _, query := range queries {
		var rows []writeRequestTarget
		if err := query.Find(&rows).Error; err != nil {
			return nil, err
		}
		for _, row := range rows {
			if !seen[row.ID] {
				seen[row.ID] = true
				targets = append(targets, row)
			}
		}
	}
	return targets, nil
}

func fetchTitleExistsMap(ctx context.Context, taskCtx taskctx.Context, titles []string) (map[string]bool, error) {
	apiServer := strings.TrimRight(taskCtx.Config().App.APIServer, "/")
	if apiServer == "" {
		return nil, fmt.Errorf("apiServer is required")
	}

	uniqueTitles := uniqueStrings(titles)
	out := make(map[string]bool, len(uniqueTitles))
	client := &http.Client{Timeout: 20 * time.Second}
	for start := 0; start < len(uniqueTitles); start += matcherPageSize {
		end := min(start+matcherPageSize, len(uniqueTitles))
		params := url.Values{"action": {"query"}, "format": {"json"}, "titles": {strings.Join(uniqueTitles[start:end], "|")}}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiServer+"/w/api.php", strings.NewReader(params.Encode()))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			_ = resp.Body.Close()
			return nil, fmt.Errorf("MediaWiki API request failed: HTTP %d", resp.StatusCode)
		}
		var payload struct {
			Query struct {
				Pages map[string]struct {
					Title   string          `json:"title"`
					Missing json.RawMessage `json:"missing"`
				} `json:"pages"`
			} `json:"query"`
		}
		decodeErr := json.NewDecoder(resp.Body).Decode(&payload)
		_ = resp.Body.Close()
		if decodeErr != nil {
			return nil, fmt.Errorf("decode MediaWiki API response: %w", decodeErr)
		}
		if payload.Query.Pages == nil {
			return nil, fmt.Errorf("MediaWiki API response is missing query.pages")
		}
		for _, page := range payload.Query.Pages {
			exists := page.Missing == nil
			for _, variant := range titleVariants(page.Title) {
				out[normalizeTitleKey(variant)] = exists
			}
		}
	}
	return out, nil
}

func resolveTitleExists(existsMap map[string]bool, title string) bool {
	for _, variant := range titleVariants(title) {
		if exists, ok := existsMap[normalizeTitleKey(variant)]; ok {
			return exists
		}
	}
	return false
}

func titleVariants(title string) []string {
	title = strings.TrimSpace(title)
	if title == "" {
		return nil
	}
	return uniqueStrings([]string{title, strings.ReplaceAll(title, "_", " "), strings.ReplaceAll(title, " ", "_")})
}

func normalizeTitleKey(title string) string { return strings.ToLower(strings.TrimSpace(title)) }

func uniqueStrings(input []string) []string {
	seen := make(map[string]bool, len(input))
	unique := make([]string, 0, len(input))
	for _, value := range input {
		if !seen[value] {
			seen[value] = true
			unique = append(unique, value)
		}
	}
	return unique
}
