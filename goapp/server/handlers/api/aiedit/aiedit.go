package aiedit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/app/config"
	"github.com/zetaoss/zengine/goapp/jobs/aieditjob"
	"github.com/zetaoss/zengine/goapp/models"
	"github.com/zetaoss/zengine/goapp/server/serverctx"

	"gorm.io/gorm"
)

const (
	perPage              = 25
	defaultPageTaskLimit = 10
	maxPageTaskLimit     = 100
)

type storePayload struct {
	PageID      int    `json:"page_id"`
	Title       string `json:"title"`
	RequestType string `json:"request_type"`
	PromptTitle string `json:"prompt_title"`
	LLMInput    string `json:"llm_input"`
}

type taskResponse struct {
	models.AIEdit
	Title string `json:"title"`
}

func MyIndex(c *serverctx.Context) {
	user, ok := c.User()
	if !ok || user.ID < 1 {
		c.JSONError(http.StatusUnauthorized, "Unauthenticated")
		return
	}

	pageID, _ := strconv.Atoi(c.R.URL.Query().Get("page_id"))
	pageTitle := strings.TrimSpace(c.R.URL.Query().Get("page_title"))

	if pageID < 1 && pageTitle == "" {
		c.JSONError(http.StatusUnprocessableEntity, "page_id or page_title is required")
		return
	}

	limit := defaultPageTaskLimit
	if rawLimit := c.R.URL.Query().Get("limit"); rawLimit != "" {
		parsedLimit, err := strconv.Atoi(rawLimit)
		if err != nil || parsedLimit < 1 || parsedLimit > maxPageTaskLimit {
			c.JSONError(http.StatusUnprocessableEntity, "Invalid limit")
			return
		}
		limit = parsedLimit
	}

	rows := make([]models.AIEdit, 0, limit)
	q := c.DB.Table("aiedit_tasks").
		Select("id, user_id, user_name, page_id, page_title, request_type, phase, llm_output, llm_model, revid, error_count, last_error, created_at, updated_at").
		Where("user_id = ?", user.ID)

	switch {
	case pageID > 0 && pageTitle != "":
		q = q.Where("page_id = ? OR page_title = ?", pageID, pageTitle)
	case pageID > 0:
		q = q.Where("page_id = ?", pageID)
	default:
		q = q.Where("page_title = ?", pageTitle)
	}

	if err := q.Order("id DESC").Limit(limit).Find(&rows).Error; err != nil {
		c.InternalError()
		return
	}
	c.JSON(taskResponses(rows))
}

func Show(c *serverctx.Context) {
	id, ok := c.PathInt("id")
	if !ok {
		c.NotFound()
		return
	}
	var row models.AIEdit
	if err := c.DB.Table("aiedit_tasks").Where("id = ?", id).Take(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.NotFound()
			return
		}
		c.InternalError()
		return
	}
	row.RetryAt = retryAt(row)
	c.JSON(newTaskResponse(row))
}

func Store(c *serverctx.Context) {
	user, ok := c.User()
	if !ok || user.ID < 1 {
		c.JSONError(http.StatusUnauthorized, "Unauthenticated")
		return
	}
	var body storePayload
	if !c.Decode(&body) {
		return
	}
	if body.RequestType != "create" && body.RequestType != "edit" {
		c.JSONError(http.StatusUnprocessableEntity, "Invalid payload")
		return
	}
	title := strings.TrimSpace(body.Title)
	if title == "" && body.PageID > 0 {
		title = resolvePageTitle(c.DB, c.Cfg, body.PageID)
	}
	if title == "" {
		c.JSONError(http.StatusUnprocessableEntity, "Invalid payload")
		return
	}
	llmInput := strings.TrimSpace(body.LLMInput)
	if llmInput == "" {
		promptTitle := normalizePromptTitle("", body.RequestType)
		template, err := fetchWikiPageText(c.R.Context(), c.Cfg, promptTitle)
		if err != nil {
			c.InternalError()
			return
		}
		existing := ""
		if body.RequestType == "edit" {
			txt, _ := fetchWikiPageText(c.R.Context(), c.Cfg, title)
			existing = txt
		}
		llmInput = renderAIEditInput(template, map[string]string{
			"제목":   title,
			"기존문서": existing,
			"추가요청": "",
			"참고자료": "",
			"분류":   "",
		})
	}
	insert := struct {
		ID          int                `gorm:"column:id"`
		UserID      int                `gorm:"column:user_id"`
		UserName    string             `gorm:"column:user_name"`
		PageID      int                `gorm:"column:page_id"`
		PageTitle   string             `gorm:"column:page_title"`
		RequestType string             `gorm:"column:request_type"`
		LLMOutput   string             `gorm:"column:llm_output"`
		LLMInput    string             `gorm:"column:llm_input"`
		Phase       models.AIEditPhase `gorm:"column:phase"`
		CreatedAt   time.Time          `gorm:"column:created_at"`
		UpdatedAt   time.Time          `gorm:"column:updated_at"`
	}{
		UserID:      user.ID,
		UserName:    user.Name,
		PageID:      body.PageID,
		PageTitle:   title,
		RequestType: body.RequestType,
		LLMOutput:   "",
		LLMInput:    llmInput,
		Phase:       models.AIEditPhasePending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := c.DB.Table("aiedit_tasks").Create(&insert).Error; err != nil {
		c.InternalError()
		return
	}
	if insert.ID > 0 {
		if promptTitle := strings.TrimSpace(body.PromptTitle); promptTitle != "" {
			c.DB.Model(&models.AIEditPrompt{}).Where("title = ?", promptTitle).UpdateColumn("use_count", gorm.Expr("use_count + 1"))
		}
		if _, err := aieditjob.Enqueue(c.R.Context(), c.AppContext, insert.ID); err != nil {
			c.InternalError()
			return
		}
	}
	c.JSON(app.H{"ok": true, "id": insert.ID, "created": true})
}

func normalizePromptTitle(title string, requestType string) string {
	title = strings.TrimSpace(title)
	if title == "" {
		if strings.TrimSpace(requestType) == "edit" {
			return "틀:프롬프트 편집"
		}
		return "틀:프롬프트 생성"
	}
	if !strings.HasPrefix(title, "틀:") {
		return "틀:" + title
	}
	return title
}

func renderAIEditInput(template string, values map[string]string) string {
	out := template
	for key, value := range values {
		out = strings.ReplaceAll(out, "{"+key+"}", value)
	}
	return out
}

func fetchWikiPageText(ctx context.Context, cfg *config.Config, title string) (string, error) {
	apiServer := strings.TrimRight(cfg.App.APIServer, "/")
	if apiServer == "" {
		return "", fmt.Errorf("APP_URL is required")
	}
	u := apiServer + "/w/api.php?action=query&format=json&formatversion=2&prop=revisions&rvprop=content&rvslots=main&titles=" + url.QueryEscape(title)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	resp, err := (&http.Client{Timeout: 20 * time.Second}).Do(req)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	var payload struct {
		Query struct {
			Pages []struct {
				Missing   bool `json:"missing"`
				Revisions []struct {
					Slots map[string]struct {
						Content string `json:"content"`
					} `json:"slots"`
				} `json:"revisions"`
			} `json:"pages"`
		} `json:"query"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}
	if len(payload.Query.Pages) == 0 || payload.Query.Pages[0].Missing {
		return "", fmt.Errorf("wiki page not found: %s", title)
	}
	if len(payload.Query.Pages[0].Revisions) == 0 {
		return "", nil
	}
	return payload.Query.Pages[0].Revisions[0].Slots["main"].Content, nil
}

func retryAt(task models.AIEdit) *string {
	return aieditjob.CalculateRetryAt(task)
}

func newTaskResponse(task models.AIEdit) taskResponse {
	task.RetryAt = retryAt(task)
	return taskResponse{
		AIEdit: task,
		Title:  task.PageTitle,
	}
}

func taskResponses(tasks []models.AIEdit) []taskResponse {
	responses := make([]taskResponse, 0, len(tasks))
	for _, task := range tasks {
		responses = append(responses, newTaskResponse(task))
	}
	return responses
}

func resolvePageTitle(db *gorm.DB, cfg *config.Config, pageID int) string {
	apiServer := strings.TrimRight(cfg.App.APIServer, "/")
	if apiServer != "" {
		u := fmt.Sprintf("%s/w/api.php?action=query&format=json&formatversion=2&pageids=%d", apiServer, pageID)
		if resp, err := http.Get(u); err == nil && resp != nil {
			defer func() {
				_ = resp.Body.Close()
			}()
			var payload app.H
			if json.NewDecoder(resp.Body).Decode(&payload) == nil {
				query, _ := payload["query"].(app.H)
				pages, _ := query["pages"].([]any)
				if len(pages) > 0 {
					p0, _ := pages[0].(app.H)
					title, _ := p0["title"].(string)
					if strings.TrimSpace(title) != "" {
						return strings.ReplaceAll(strings.TrimSpace(title), "_", " ")
					}
				}
			}
		}
	}
	var row struct {
		PageTitle string `gorm:"column:page_title"`
	}
	if err := db.Table("zetawiki.page").Select("page_title").Where("page_id = ?", pageID).Take(&row).Error; err != nil {
		return ""
	}
	return strings.ReplaceAll(strings.TrimSpace(row.PageTitle), "_", " ")
}
