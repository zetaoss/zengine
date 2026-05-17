package editbot

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/app/config"
	"github.com/zetaoss/zengine/goapp/jobs/editbotjob"
	"github.com/zetaoss/zengine/goapp/models"
	"github.com/zetaoss/zengine/goapp/server/middleware"
	"github.com/zetaoss/zengine/goapp/server/paginator"
	"github.com/zetaoss/zengine/goapp/server/serverctx"

	"gorm.io/gorm"
)

const perPage = 25

type writeRequestPromptPayload struct {
	RequestType string `json:"request_type"`
	LLMInput    string `json:"llm_input"`
}

func Index(c *serverctx.Context) {
	rows := make([]models.EditBot, 0, perPage)
	q := c.DB.Table("edit_tasks").
		Select("id, user_id, user_name, title, request_type, phase, llm_model, revid, error_count, last_error, created_at, updated_at").
		Order("id DESC")
	payload, err := paginator.Paginate(c.R, q, perPage, &rows)
	if err != nil {
		c.InternalError()
		return
	}
	for i := range rows {
		rows[i].RetryAt = retryAt(rows[i])
	}
	c.JSON(payload)
}

func Show(c *serverctx.Context) {
	id, ok := c.PathInt("task")
	if !ok {
		c.NotFound()
		return
	}
	var row models.EditBot
	if err := c.DB.Table("edit_tasks").Where("id = ?", id).Take(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.NotFound()
			return
		}
		c.InternalError()
		return
	}
	row.RetryAt = retryAt(row)
	c.JSON(row)
}

func StoreFromWriteRequest(c *serverctx.Context) {
	user, ok := middleware.UserFromRequest(c.R)
	if !ok || user.ID < 1 {
		c.JSONError(http.StatusUnauthorized, "Unauthenticated")
		return
	}
	writeRequestID, ok := c.PathInt("writeRequest")
	if !ok {
		c.NotFound()
		return
	}
	title, ok := findWriteRequestTitle(c.DB, writeRequestID)
	if !ok {
		c.NotFound()
		return
	}
	var body writeRequestPromptPayload
	if !c.Decode(&body) {
		return
	}
	body.LLMInput = strings.TrimSpace(body.LLMInput)
	if body.LLMInput == "" {
		c.JSONError(http.StatusUnprocessableEntity, "LLM 입력을 먼저 렌더링해 주세요.")
		return
	}
	insert := struct {
		ID          int                 `gorm:"column:id"`
		UserID      int                 `gorm:"column:user_id"`
		UserName    string              `gorm:"column:user_name"`
		Title       string              `gorm:"column:title"`
		RequestType string              `gorm:"column:request_type"`
		LLMInput    string              `gorm:"column:llm_input"`
		Phase       models.EditBotPhase `gorm:"column:phase"`
		CreatedAt   time.Time           `gorm:"column:created_at"`
		UpdatedAt   time.Time           `gorm:"column:updated_at"`
	}{
		UserID:      user.ID,
		UserName:    user.Name,
		Title:       title,
		RequestType: body.RequestType,
		LLMInput:    body.LLMInput,
		Phase:       models.EditBotPhasePending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := c.DB.Table("edit_tasks").Create(&insert).Error; err != nil {
		c.InternalError()
		return
	}
	if insert.ID > 0 {
		if _, err := editbotjob.Enqueue(c.R.Context(), c.AppContext, insert.ID); err != nil {
			c.InternalError()
			return
		}
	}
	c.JSON(app.H{"ok": true, "id": insert.ID})
}

func StoreFromPage(c *serverctx.Context) {
	user, ok := middleware.UserFromRequest(c.R)
	if !ok || user.ID < 1 {
		c.JSONError(http.StatusUnauthorized, "Unauthenticated")
		return
	}
	var body struct {
		PageID      int    `json:"page_id"`
		RequestType string `json:"request_type"`
		LLMInput    string `json:"llm_input"`
	}
	if !c.Decode(&body) {
		return
	}
	if body.PageID < 1 || (body.RequestType != "create" && body.RequestType != "edit") {
		c.JSONError(http.StatusUnprocessableEntity, "Invalid payload")
		return
	}
	title := resolvePageTitle(c.DB, c.Cfg, body.PageID)
	if title == "" {
		c.JSONError(http.StatusNotFound, "문서를 찾을 수 없습니다.")
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
		llmInput = renderEditBotInput(template, map[string]string{
			"제목":   title,
			"기존문서": existing,
			"추가요청": "",
			"참고자료": "",
			"분류":   "",
		})
	}
	insert := struct {
		ID          int                 `gorm:"column:id"`
		UserID      int                 `gorm:"column:user_id"`
		UserName    string              `gorm:"column:user_name"`
		Title       string              `gorm:"column:title"`
		RequestType string              `gorm:"column:request_type"`
		LLMOutput   string              `gorm:"column:llm_output"`
		LLMInput    string              `gorm:"column:llm_input"`
		Phase       models.EditBotPhase `gorm:"column:phase"`
		CreatedAt   time.Time           `gorm:"column:created_at"`
		UpdatedAt   time.Time           `gorm:"column:updated_at"`
	}{
		UserID:      user.ID,
		UserName:    user.Name,
		Title:       title,
		RequestType: body.RequestType,
		LLMOutput:   "",
		LLMInput:    llmInput,
		Phase:       models.EditBotPhasePending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := c.DB.Table("edit_tasks").Create(&insert).Error; err != nil {
		c.InternalError()
		return
	}
	if insert.ID > 0 {
		if _, err := editbotjob.Enqueue(c.R.Context(), c.AppContext, insert.ID); err != nil {
			c.InternalError()
			return
		}
	}
	c.JSON(app.H{"ok": true, "id": insert.ID, "created": true})
}

func Destroy(c *serverctx.Context) {
	id, ok := c.PathInt("task")
	if !ok {
		c.NotFound()
		return
	}
	if err := c.DB.Table("edit_tasks").Where("id = ?", id).Delete(nil).Error; err != nil {
		c.InternalError()
		return
	}
	c.JSON(map[string]bool{"ok": true})
}

func findWriteRequestTitle(db *gorm.DB, id int) (string, bool) {
	var wr struct {
		Title string `gorm:"column:title"`
	}
	if err := db.Table("write_requests").Select("title").Where("id = ?", id).Take(&wr).Error; err != nil {
		return "", false
	}
	return wr.Title, strings.TrimSpace(wr.Title) != ""
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

func renderEditBotInput(template string, values map[string]string) string {
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

func retryAt(task models.EditBot) *string {
	return editbotjob.CalculateRetryAt(task)
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
