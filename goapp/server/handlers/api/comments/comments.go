package comments

import (
	"net/http"
	"strings"
	"time"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/server/middleware"
	"github.com/zetaoss/zengine/goapp/server/serverctx"

	"gorm.io/gorm"
)

type row struct {
	ID       int       `json:"id" gorm:"column:id"`
	PageID   int       `json:"page_id" gorm:"column:page_id"`
	UserID   int       `json:"user_id" gorm:"column:user_id"`
	UserName string    `json:"user_name" gorm:"column:user_name"`
	Created  time.Time `json:"created" gorm:"column:created"`
	Message  string    `json:"message" gorm:"column:message"`
}

type recentRow struct {
	PageTitle string    `json:"page_title" gorm:"column:page_title"`
	ID        int       `json:"id" gorm:"column:id"`
	PageID    int       `json:"page_id" gorm:"column:page_id"`
	UserID    int       `json:"user_id" gorm:"column:user_id"`
	UserName  string    `json:"user_name" gorm:"column:user_name"`
	Created   time.Time `json:"created" gorm:"column:created"`
	Message   string    `json:"message" gorm:"column:message"`
}

func Recent(c *serverctx.Context) {
	rows := make([]recentRow, 0, 10)
	err := c.DB.Table("zetawiki.page_comments").
		Select("page.page_title", "page_comments.id", "page_comments.page_id", "page_comments.user_id", "page_comments.user_name", "page_comments.created", "page_comments.message").
		Joins("JOIN zetawiki.page ON page_comments.page_id = page.page_id").
		Order("page_comments.created DESC").
		Limit(10).
		Find(&rows).Error
	if err != nil {
		c.InternalError()
		return
	}
	c.JSON(rows)
}

func List(c *serverctx.Context) {
	pageID, ok := c.PathInt("pageID")
	if !ok {
		c.NotFound()
		return
	}
	rows := make([]row, 0, 128)
	err := c.DB.Table("zetawiki.page_comments").
		Select("id", "user_id", "user_name", "created", "message").
		Where("page_id = ?", pageID).
		Order("id DESC").
		Find(&rows).Error
	if err != nil {
		c.InternalError()
		return
	}
	c.JSON(rows)
}

func Store(c *serverctx.Context) {
	user, ok := middleware.UserFromRequest(c.R)
	if !ok || user.ID < 1 {
		c.JSONError(http.StatusUnauthorized, "Unauthenticated")
		return
	}
	var body struct {
		PageID  int    `json:"pageid"`
		Message string `json:"message"`
	}
	if !c.Decode(&body) {
		return
	}
	message := strings.TrimSpace(body.Message)
	if body.PageID < 1 || message == "" || len(message) > 5000 {
		c.JSONError(http.StatusUnprocessableEntity, "Invalid payload")
		return
	}
	insert := app.H{
		"page_id":    body.PageID,
		"message":    message,
		"user_id":    user.ID,
		"user_name":  user.Name,
		"created":    time.Now(),
		"created_at": time.Now(),
		"updated_at": time.Now(),
	}
	if err := c.DB.Table("zetawiki.page_comments").Create(insert).Error; err != nil {
		c.InternalError()
		return
	}
	c.JSON(map[string]bool{"ok": true})
}

func Update(c *serverctx.Context) {
	user, ok := middleware.UserFromRequest(c.R)
	if !ok || user.ID < 1 {
		c.JSONError(http.StatusUnauthorized, "Unauthenticated")
		return
	}
	id, ok := c.PathInt("id")
	if !ok {
		c.NotFound()
		return
	}
	var target struct {
		UserID int `gorm:"column:user_id"`
	}
	if err := c.DB.Table("zetawiki.page_comments").Select("user_id").Where("id = ?", id).Take(&target).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.NotFound()
			return
		}
		c.InternalError()
		return
	}
	if user.ID != target.UserID {
		c.JSONError(http.StatusForbidden, "Forbidden")
		return
	}
	var body struct {
		Message string `json:"message"`
	}
	if !c.Decode(&body) {
		return
	}
	message := strings.TrimSpace(body.Message)
	if message == "" || len(message) > 5000 {
		c.JSONError(http.StatusUnprocessableEntity, "Invalid payload")
		return
	}
	if err := c.DB.Table("zetawiki.page_comments").Where("id = ?", id).Updates(app.H{"message": message, "updated_at": time.Now()}).Error; err != nil {
		c.InternalError()
		return
	}
	c.JSON(map[string]bool{"ok": true})
}

func Destroy(c *serverctx.Context) {
	user, ok := middleware.UserFromRequest(c.R)
	if !ok || user.ID < 1 {
		c.JSONError(http.StatusUnauthorized, "Unauthenticated")
		return
	}
	id, ok := c.PathInt("id")
	if !ok {
		c.NotFound()
		return
	}
	var target struct {
		UserID int `gorm:"column:user_id"`
	}
	if err := c.DB.Table("zetawiki.page_comments").Select("user_id").Where("id = ?", id).Take(&target).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.NotFound()
			return
		}
		c.InternalError()
		return
	}
	if user.ID != target.UserID && !user.IsSysop() {
		c.JSONError(http.StatusForbidden, "Forbidden")
		return
	}
	if err := c.DB.Table("zetawiki.page_comments").Where("id = ?", id).Delete(nil).Error; err != nil {
		c.InternalError()
		return
	}
	c.JSON(map[string]bool{"ok": true})
}
