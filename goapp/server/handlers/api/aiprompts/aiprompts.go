package aiprompts

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/models"
	"github.com/zetaoss/zengine/goapp/server/serverctx"
	"gorm.io/gorm"
)

func Index(c *serverctx.Context) {
	var rows []models.AIEditPrompt

	if err := c.DB.Order("id DESC").Find(&rows).Error; err != nil {
		c.InternalError()
		return
	}
	c.JSON(rows)
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

	c.JSON(row)
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
	body.RequestType = strings.TrimSpace(body.RequestType)
	if body.RequestType != "create" && body.RequestType != "edit" {
		c.JSONError(http.StatusUnprocessableEntity, "유효하지 않은 요청 유형입니다.")
		return
	}

	user, _ := c.User()

	now := time.Now().Format("2006-01-02 15:04:05")
	row := models.AIEditPrompt{
		ID:          body.ID,
		Title:       body.Title,
		RequestType: body.RequestType,
		Content:     body.Content,
		UpdatedAt:   now,
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

		var count int64
		c.DB.Model(&models.AIEditPrompt{}).Where("title = ? AND id != ?", row.Title, row.ID).Count(&count)
		if count > 0 {
			c.JSONError(http.StatusConflict, "이미 사용 중인 제목입니다.")
			return
		}

		if err := c.DB.Model(&models.AIEditPrompt{}).Where("id = ?", row.ID).Select("title", "request_type", "content", "updated_at").Updates(&row).Error; err != nil {
			c.InternalError()
			return
		}
	} else {
		row.UserID = user.ID
		row.UserName = user.Name
		row.CreatedAt = now

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
	c.JSON(row)
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
