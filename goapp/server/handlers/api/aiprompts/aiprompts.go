package aiprompts

import (
	"encoding/json"
	"errors"
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

// promptView is a template with the messages it runs with (aiedit.TemplateMessages).
type promptView struct {
	models.AIEditPrompt
	Messages []prompts.Message `json:"messages"`
}

func newPromptView(p models.AIEditPrompt) promptView {
	return promptView{AIEditPrompt: p, Messages: aiedit.TemplateMessages(p)}
}

func Index(c *serverctx.Context) {
	var rows []models.AIEditPrompt

	if err := c.DB.Order("id DESC").Find(&rows).Error; err != nil {
		c.InternalError()
		return
	}
	views := make([]promptView, len(rows))
	for i, r := range rows {
		views[i] = newPromptView(r)
	}
	c.JSON(views)
}

func Show(c *serverctx.Context) {
	id, ok := c.PathInt("id")
	if !ok {
		c.NotFound()
		return
	}
	var row models.AIEditPrompt

	if err := c.DB.Where("id = ?", id).Take(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.NotFound()
			return
		}
		c.InternalError()
		return
	}

	c.JSON(newPromptView(row))
}

func Exists(c *serverctx.Context) {
	title := strings.TrimSpace(c.R.URL.Query().Get("title"))
	excludeID, _ := strconv.Atoi(c.R.URL.Query().Get("exclude_id"))

	if title == "" {
		c.JSON(app.H{"exists": false})
		return
	}

	var count int64
	q := c.DB.Model(&models.AIEditPrompt{}).Where("title = ?", title)
	if excludeID > 0 {
		q = q.Where("id != ?", excludeID)
	}
	q.Count(&count)

	c.JSON(app.H{"exists": count > 0})
}

func Store(c *serverctx.Context) {
	var body struct {
		ID          int    `json:"id"`
		Title       string `json:"title"`
		Description string `json:"description"`
		RequestType string `json:"request_type"`
		Content     string `json:"content"`
		// Model is left as it is when omitted; "" runs the template on its chain's model.
		Model *string `json:"model"`
		// Messages, when given, are the template's own messages in place of the chain step's; they are checked
		// against the chain (prompts.Chain.ValidateMessages) and content becomes their text.
		Messages *[]prompts.Message `json:"messages"`
		// Bump is how an edit that changes the messages moves the version: "minor" (the default) or "major".
		Bump string `json:"bump"`
	}
	if !c.Decode(&body) {
		return
	}
	if body.Model != nil {
		m := strings.TrimSpace(*body.Model)
		if m != "" {
			var err error
			if m, err = config.ValidateModel(m); err != nil {
				c.JSONError(http.StatusUnprocessableEntity, err.Error())
				return
			}
		}
		body.Model = &m
	}
	body.Title = strings.TrimSpace(body.Title)
	if body.Title == "" {
		c.JSONError(http.StatusUnprocessableEntity, "제목은 필수입니다.")
		return
	}
	requestType, ok := models.NormalizeAIEditRequestType(body.RequestType)
	if !ok {
		c.JSONError(http.StatusUnprocessableEntity, "유효하지 않은 요청 유형입니다.")
		return
	}
	body.RequestType = requestType
	var messagesJSON string
	if body.Messages != nil {
		messages := *body.Messages
		if err := aiedit.ChainOf(requestType).ValidateMessages(messages); err != nil {
			c.JSONError(http.StatusUnprocessableEntity, "메시지: "+err.Error())
			return
		}
		b, _ := json.Marshal(messages)
		messagesJSON = string(b)
		body.Content = aiedit.MessagesText(messages)
	}

	user, _ := c.User()

	now := time.Now().Format("2006-01-02 15:04:05")
	row := models.AIEditPrompt{
		ID:          body.ID,
		Title:       body.Title,
		Description: strings.TrimSpace(body.Description),
		RequestType: body.RequestType,
		Content:     body.Content,
		UpdatedAt:   now,
	}
	columns := []string{"title", "description", "request_type", "content", "version", "updated_at"}
	if body.Messages != nil {
		row.Messages = messagesJSON
		columns = append(columns, "messages")
	}
	if body.Model != nil {
		row.Model = *body.Model
		columns = append(columns, "model")
	}

	if row.ID > 0 {
		var existing models.AIEditPrompt
		if err := c.DB.Where("id = ?", row.ID).First(&existing).Error; err != nil {
			c.InternalError()
			return
		}
		if existing.UserID != user.ID {
			c.JSONError(http.StatusForbidden, "본인만 편집할 수 있습니다.")
			return
		}
		row.Version = existing.Version
		changed := row.Content != existing.Content
		if body.Messages != nil {
			before, _ := json.Marshal(aiedit.TemplateMessages(existing))
			changed = messagesJSON != string(before)
		}
		if changed {
			row.Version = prompts.NextVersion(existing.Version, body.Bump == "major")
		}

		var count int64
		c.DB.Model(&models.AIEditPrompt{}).Where("title = ? AND id != ?", row.Title, row.ID).Count(&count)
		if count > 0 {
			c.JSONError(http.StatusConflict, "이미 사용 중인 제목입니다.")
			return
		}

		if err := c.DB.Model(&models.AIEditPrompt{}).Where("id = ?", row.ID).Select(columns).Updates(&row).Error; err != nil {
			c.InternalError()
			return
		}
	} else {
		row.UserID = user.ID
		row.UserName = user.Name
		row.CreatedAt = now
		row.Version = "v1.0"

		var count int64
		c.DB.Model(&models.AIEditPrompt{}).Where("title = ?", row.Title).Count(&count)
		if count > 0 {
			c.JSONError(http.StatusConflict, "이미 사용 중인 제목입니다.")
			return
		}

		if err := c.DB.Create(&row).Error; err != nil {
			c.InternalError()
			return
		}
	}
	c.JSON(newPromptView(row))
}

func Destroy(c *serverctx.Context) {
	id, ok := c.PathInt("id")
	if !ok {
		c.NotFound()
		return
	}
	if err := c.DB.Where("id = ?", id).Delete(&models.AIEditPrompt{}).Error; err != nil {
		c.InternalError()
		return
	}
	c.JSON(app.H{"ok": true})
}
