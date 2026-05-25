package editbot

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
	"github.com/zetaoss/zengine/goapp/jobs/editbotjob"
	"github.com/zetaoss/zengine/goapp/models"
	"github.com/zetaoss/zengine/goapp/server/paginator"
	"github.com/zetaoss/zengine/goapp/server/serverctx"

	"gorm.io/gorm"
)

const perPage = 25

type writeRequestPromptPayload struct {
	RequestType string `json:"request_type"`
	PromptTitle string `json:"prompt_title"`
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
	id, ok := c.PathInt("id")
	if !ok {
		c.NotFound()
		return
	}
	var row models.EditBot
	if err := c.DB.Table("edit_tasks").Where("id = ?", id).Take(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
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
	user, ok := c.User()
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
		if promptTitle := strings.TrimSpace(body.PromptTitle); promptTitle != "" {
			c.DB.Model(&models.EditbotPrompt{}).Where("title = ?", promptTitle).UpdateColumn("use_count", gorm.Expr("use_count + 1"))
		}
		if _, err := editbotjob.Enqueue(c.R.Context(), c.AppContext, insert.ID); err != nil {
			c.InternalError()
			return
		}
	}

	// update write request status
	_ = c.DB.Table("write_requests").Where("id = ?", writeRequestID).Update("status", "done")

	c.JSON(app.H{"ok": true, "id": insert.ID})
}

func StoreFromPage(c *serverctx.Context) {
	user, ok := c.User()
	if !ok || user.ID < 1 {
		c.JSONError(http.StatusUnauthorized, "Unauthenticated")
		return
	}
	var body struct {
		PageID      int    `json:"page_id"`
		RequestType string `json:"request_type"`
		PromptTitle string `json:"prompt_title"`
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
		if promptTitle := strings.TrimSpace(body.PromptTitle); promptTitle != "" {
			c.DB.Model(&models.EditbotPrompt{}).Where("title = ?", promptTitle).UpdateColumn("use_count", gorm.Expr("use_count + 1"))
		}
		if _, err := editbotjob.Enqueue(c.R.Context(), c.AppContext, insert.ID); err != nil {
			c.InternalError()
			return
		}
	}
	c.JSON(app.H{"ok": true, "id": insert.ID, "created": true})
}

func Destroy(c *serverctx.Context) {
	id, ok := c.PathInt("id")
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

func PromptIndex(c *serverctx.Context) {
	user, _ := c.User()
	var rows []models.EditbotPrompt

	if err := c.DB.Order("id DESC").Find(&rows).Error; err != nil {
		c.InternalError()
		return
	}

	if user.ID < 1 {
		c.JSON(rows)
		return
	}

	// Fetch favorite prompt IDs for this user
	var favIDs []int
	c.DB.Table("editbot_prompt_favorites").Where("user_id = ?", user.ID).Pluck("prompt_id", &favIDs)
	favMap := make(map[int]bool, len(favIDs))
	for _, id := range favIDs {
		favMap[id] = true
	}

	// Create response with is_favorite field
	type promptResp struct {
		models.EditbotPrompt
		IsFavorite bool `json:"is_favorite"`
	}
	resp := make([]promptResp, len(rows))
	for i, r := range rows {
		resp[i] = promptResp{
			EditbotPrompt: r,
			IsFavorite:    favMap[r.ID],
		}
	}

	c.JSON(resp)
}

func PromptToggleFavorite(c *serverctx.Context) {
	id, ok := c.PathInt("id")
	if !ok {
		c.NotFound()
		return
	}
	user, ok := c.User()
	if !ok || user.ID < 1 {
		c.JSONError(http.StatusUnauthorized, "Unauthenticated")
		return
	}

	var count int64
	c.DB.Table("editbot_prompt_favorites").Where("user_id = ? AND prompt_id = ?", user.ID, id).Count(&count)

	if count > 0 {
		if err := c.DB.Table("editbot_prompt_favorites").Where("user_id = ? AND prompt_id = ?", user.ID, id).Delete(nil).Error; err != nil {
			c.InternalError()
			return
		}
		c.JSON(app.H{"is_favorite": false})
	} else {
		insert := map[string]any{
			"user_id":    user.ID,
			"prompt_id":   id,
			"created_at": time.Now().Format("2006-01-02 15:04:05"),
		}
		if err := c.DB.Table("editbot_prompt_favorites").Create(&insert).Error; err != nil {
			c.InternalError()
			return
		}
		c.JSON(app.H{"is_favorite": true})
	}
}

func PromptShow(c *serverctx.Context) {
	id, ok := c.PathInt("id")
	if !ok {
		c.NotFound()
		return
	}
	user, _ := c.User()
	var row models.EditbotPrompt

	if err := c.DB.Where("id = ?", id).Take(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.NotFound()
			return
		}
		c.InternalError()
		return
	}

	isFavorite := false
	if user.ID > 0 {
		var count int64
		c.DB.Table("editbot_prompt_favorites").Where("user_id = ? AND prompt_id = ?", user.ID, id).Count(&count)
		isFavorite = count > 0
	}

	c.JSON(struct {
		models.EditbotPrompt
		IsFavorite bool `json:"is_favorite"`
	}{
		EditbotPrompt: row,
		IsFavorite:    isFavorite,
	})
}

func PromptExists(c *serverctx.Context) {
	title := strings.TrimSpace(c.R.URL.Query().Get("title"))
	excludeID, _ := strconv.Atoi(c.R.URL.Query().Get("exclude_id"))

	if title == "" {
		c.JSON(app.H{"exists": false})
		return
	}

	var count int64
	q := c.DB.Model(&models.EditbotPrompt{}).Where("title = ?", title)
	if excludeID > 0 {
		q = q.Where("id != ?", excludeID)
	}
	q.Count(&count)

	c.JSON(app.H{"exists": count > 0})
}

func PromptStore(c *serverctx.Context) {
	var body struct {
		ID          int    `json:"id"`
		Title       string `json:"title"`
		RequestType string `json:"request_type"`
		Content     string `json:"content"`
	}
	if !c.Decode(&body) {
		return
	}
	body.Title = strings.TrimSpace(body.Title)
	if body.Title == "" {
		c.JSONError(http.StatusUnprocessableEntity, "제목은 필수입니다.")
		return
	}

	user, _ := c.User()

	now := time.Now().Format("2006-01-02 15:04:05")
	row := models.EditbotPrompt{
		ID:          body.ID,
		Title:       body.Title,
		RequestType: body.RequestType,
		Content:     body.Content,
		UpdatedAt:   now,
	}

	if row.ID > 0 {
		var existing models.EditbotPrompt
		if err := c.DB.Where("id = ?", row.ID).First(&existing).Error; err != nil {
			c.InternalError()
			return
		}
		if existing.UserID != user.ID {
			c.JSONError(http.StatusForbidden, "본인만 편집할 수 있습니다.")
			return
		}

		var count int64
		c.DB.Model(&models.EditbotPrompt{}).Where("title = ? AND id != ?", row.Title, row.ID).Count(&count)
		if count > 0 {
			c.JSONError(http.StatusConflict, "이미 사용 중인 제목입니다.")
			return
		}

		if err := c.DB.Model(&models.EditbotPrompt{}).Where("id = ?", row.ID).Select("title", "request_type", "content", "updated_at").Updates(&row).Error; err != nil {
			c.InternalError()
			return
		}
	} else {
		row.UserID = user.ID
		row.UserName = user.Name
		row.CreatedAt = now

		var count int64
		c.DB.Model(&models.EditbotPrompt{}).Where("title = ?", row.Title).Count(&count)
		if count > 0 {
			c.JSONError(http.StatusConflict, "이미 사용 중인 제목입니다.")
			return
		}

		if err := c.DB.Create(&row).Error; err != nil {
			c.InternalError()
			return
		}
	}
	c.JSON(row)
}

func PromptDestroy(c *serverctx.Context) {
	id, ok := c.PathInt("id")
	if !ok {
		c.NotFound()
		return
	}
	if err := c.DB.Where("id = ?", id).Delete(&models.EditbotPrompt{}).Error; err != nil {
		c.InternalError()
		return
	}
	c.JSON(app.H{"ok": true})
}
