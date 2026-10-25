package aiedit

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/app/config"
	"github.com/zetaoss/zengine/goapp/models"
	"github.com/zetaoss/zengine/goapp/prompts"
	"github.com/zetaoss/zengine/goapp/server/serverctx"
	"github.com/zetaoss/zengine/goapp/tasks/aiedit"

	"gorm.io/gorm"
)

const (
	perPage              = 25
	defaultPageTaskLimit = 10
	maxPageTaskLimit     = 100
)

// storePayload is an AI edit request: the page, the request type, an AI edit prompt template (prompt_id, or
// none for the chain's own messages), the page as it stands in the editor (wikitext, for update) and the run
// form's values (user_inputs). prompt_title and llm_input come from skins loaded before server rendering.
type storePayload struct {
	PageID      int            `json:"page_id"`
	Title       string         `json:"title"`
	RequestType string         `json:"request_type"`
	PromptID    int            `json:"prompt_id"`
	PromptTitle string         `json:"prompt_title"`
	Wikitext    string         `json:"wikitext"`
	UserInputs  map[string]any `json:"user_inputs"`
	LLMInput    string         `json:"llm_input"`
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
	q = q.Where("phase = ? AND TRIM(llm_output) <> ''", models.AIEditPhaseCompleted)

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
	req, prompt, ok := request(c, &body)
	if !ok {
		return
	}
	var r aiedit.Rendered
	if body.LLMInput != "" && body.UserInputs == nil && prompt == nil {
		r = aiedit.Rendered{Chain: aiedit.ChainOf(req.RequestType), Messages: []prompts.Message{{Role: "user", Content: strings.TrimSpace(body.LLMInput)}}}
	} else if r, ok = render(c, req); !ok {
		return
	}
	messagesJSON, _ := json.Marshal(r.Messages)
	var inputsJSON []byte
	if r.Vars != nil {
		inputsJSON, _ = json.Marshal(r.Vars)
	}
	insert := struct {
		ID            int                `gorm:"column:id"`
		UserID        int                `gorm:"column:user_id"`
		UserName      string             `gorm:"column:user_name"`
		PageID        int                `gorm:"column:page_id"`
		PageTitle     string             `gorm:"column:page_title"`
		RequestType   string             `gorm:"column:request_type"`
		PromptID      int                `gorm:"column:prompt_id"`
		PromptVersion string             `gorm:"column:prompt_version"`
		ChainVersion  string             `gorm:"column:chain_version"`
		Inputs        string             `gorm:"column:inputs"`
		LLMMessages   string             `gorm:"column:llm_messages"`
		LLMOutput     string             `gorm:"column:llm_output"`
		LLMInput      string             `gorm:"column:llm_input"`
		Phase         models.AIEditPhase `gorm:"column:phase"`
		CreatedAt     time.Time          `gorm:"column:created_at"`
		UpdatedAt     time.Time          `gorm:"column:updated_at"`
	}{
		UserID:        user.ID,
		UserName:      user.Name,
		PageID:        body.PageID,
		PageTitle:     req.Title,
		RequestType:   req.RequestType,
		PromptID:      promptID(prompt),
		PromptVersion: promptVersion(prompt),
		ChainVersion:  r.Chain.Metadata.Version,
		Inputs:        string(inputsJSON),
		LLMMessages:   string(messagesJSON),
		LLMOutput:     "",
		LLMInput:      aiedit.MessagesText(r.Messages),
		Phase:         models.AIEditPhasePending,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	if err := c.DB.Table("aiedit_tasks").Create(&insert).Error; err != nil {
		c.InternalError()
		return
	}
	if insert.ID > 0 {
		switch promptTitle := strings.TrimSpace(body.PromptTitle); {
		case prompt != nil:
			c.DB.Model(&models.AIEditPrompt{}).Where("id = ?", prompt.ID).UpdateColumns(map[string]any{"use_count": gorm.Expr("use_count + 1"), "used_at": time.Now()})
		case promptTitle != "":
			c.DB.Model(&models.AIEditPrompt{}).Where("title = ?", promptTitle).UpdateColumns(map[string]any{"use_count": gorm.Expr("use_count + 1"), "used_at": time.Now()})
		}
		if _, err := aiedit.Enqueue(c.R.Context(), c.AppContext, insert.ID); err != nil {
			c.InternalError()
			return
		}
	}
	c.JSON(app.H{"ok": true, "id": insert.ID, "created": true})
}

// Preview renders a request as Store would and returns the messages without storing anything.
func Preview(c *serverctx.Context) {
	if user, ok := c.User(); !ok || user.ID < 1 {
		c.JSONError(http.StatusUnauthorized, "Unauthenticated")
		return
	}
	var body storePayload
	if !c.Decode(&body) {
		return
	}
	req, _, ok := request(c, &body)
	if !ok {
		return
	}
	r, ok := render(c, req)
	if !ok {
		return
	}
	c.JSON(app.H{"chain": r.Chain.Name, "version": r.Chain.Metadata.Version, "messages": r.Messages})
}

// request checks the payload and loads its template, answering 422 (or 500) itself when it returns false.
func request(c *serverctx.Context, body *storePayload) (aiedit.Request, *models.AIEditPrompt, bool) {
	requestType, ok := models.NormalizeAIEditRequestType(body.RequestType)
	if !ok {
		c.JSONError(http.StatusUnprocessableEntity, "Invalid payload")
		return aiedit.Request{}, nil, false
	}
	title := strings.TrimSpace(body.Title)
	if title == "" && body.PageID > 0 {
		title = resolvePageTitle(c.DB, c.Cfg, body.PageID)
	}
	if title == "" {
		c.JSONError(http.StatusUnprocessableEntity, "Invalid payload")
		return aiedit.Request{}, nil, false
	}
	var prompt *models.AIEditPrompt
	if body.PromptID > 0 {
		var row models.AIEditPrompt
		if err := c.DB.Where("id = ?", body.PromptID).Take(&row).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSONError(http.StatusUnprocessableEntity, "Unknown prompt")
			} else {
				c.InternalError()
			}
			return aiedit.Request{}, nil, false
		}
		if row.RequestType != requestType {
			c.JSONError(http.StatusUnprocessableEntity, "The prompt is for another request type")
			return aiedit.Request{}, nil, false
		}
		prompt = &row
	}
	return aiedit.Request{RequestType: requestType, Title: title, Wikitext: body.Wikitext, Prompt: prompt, UserInputs: body.UserInputs}, prompt, true
}

func render(c *serverctx.Context, req aiedit.Request) (aiedit.Rendered, bool) {
	r, err := aiedit.Render(c.R.Context(), c.Cfg, req)
	if err != nil {
		if aiedit.IsInputError(err) {
			c.JSONError(http.StatusUnprocessableEntity, err.Error())
		} else {
			c.InternalError()
		}
		return aiedit.Rendered{}, false
	}
	return r, true
}

func promptID(p *models.AIEditPrompt) int {
	if p == nil {
		return 0
	}
	return p.ID
}

func promptVersion(p *models.AIEditPrompt) string {
	if p == nil {
		return ""
	}
	return p.Version
}

func retryAt(task models.AIEdit) *string {
	return aiedit.CalculateRetryAt(task)
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
