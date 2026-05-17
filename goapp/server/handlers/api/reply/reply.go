package reply

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/models"
	"github.com/zetaoss/zengine/goapp/server/middleware"
	"github.com/zetaoss/zengine/goapp/server/serverctx"

	"gorm.io/gorm"
)

func Index(c *serverctx.Context) {
	postID, ok := parseID(c.R, "post")
	if !ok {
		http.NotFound(c.W, c.R)
		return
	}
	rows := make([]models.ForumReply, 0, 128)
	if err := c.DB.Table("replies").Where("post_id = ?", postID).Order("created_at DESC").Find(&rows).Error; err != nil {
		http.Error(c.W, "internal server error", http.StatusInternalServerError)
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
	if user.IsBlocked() {
		c.JSONError(http.StatusForbidden, "Forbidden")
		return
	}
	postID, ok := parseID(c.R, "post")
	if !ok {
		http.NotFound(c.W, c.R)
		return
	}
	var body struct {
		Body string `json:"body"`
	}
	if err := json.NewDecoder(c.R.Body).Decode(&body); err != nil {
		c.JSONError(http.StatusUnprocessableEntity, "Invalid payload")
		return
	}
	body.Body = strings.TrimSpace(body.Body)
	if body.Body == "" || len(body.Body) > 5000 {
		c.JSONError(http.StatusUnprocessableEntity, "Invalid payload")
		return
	}
	row := models.ForumReply{
		PostID:    postID,
		UserID:    user.ID,
		UserName:  user.Name,
		Body:      body.Body,
		CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
	}
	if err := c.DB.Table("replies").Create(&row).Error; err != nil {
		http.Error(c.W, "internal server error", http.StatusInternalServerError)
		return
	}
	c.JSON(row)
}

func Update(c *serverctx.Context) {
	user, ok := middleware.UserFromRequest(c.R)
	if !ok || user.ID < 1 {
		c.JSONError(http.StatusUnauthorized, "Unauthenticated")
		return
	}
	postID, ok := parseID(c.R, "post")
	if !ok {
		http.NotFound(c.W, c.R)
		return
	}
	replyID, ok := parseID(c.R, "reply")
	if !ok {
		http.NotFound(c.W, c.R)
		return
	}

	var target struct {
		PostID int `gorm:"column:post_id"`
		UserID int `gorm:"column:user_id"`
	}
	if err := c.DB.Table("replies").Select("post_id, user_id").Where("id = ?", replyID).Take(&target).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.NotFound(c.W, c.R)
			return
		}
		http.Error(c.W, "internal server error", http.StatusInternalServerError)
		return
	}
	if target.PostID != postID {
		http.NotFound(c.W, c.R)
		return
	}
	if target.UserID != user.ID {
		c.JSONError(http.StatusForbidden, "Forbidden")
		return
	}

	var body struct {
		Body string `json:"body"`
	}
	if err := json.NewDecoder(c.R.Body).Decode(&body); err != nil {
		c.JSONError(http.StatusUnprocessableEntity, "Invalid payload")
		return
	}
	body.Body = strings.TrimSpace(body.Body)
	if body.Body == "" || len(body.Body) > 5000 {
		c.JSONError(http.StatusUnprocessableEntity, "Invalid payload")
		return
	}

	if err := c.DB.Table("replies").Where("id = ?", replyID).Updates(app.H{
		"body":       body.Body,
		"updated_at": time.Now(),
	}).Error; err != nil {
		http.Error(c.W, "internal server error", http.StatusInternalServerError)
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
	postID, ok := parseID(c.R, "post")
	if !ok {
		http.NotFound(c.W, c.R)
		return
	}
	replyID, ok := parseID(c.R, "reply")
	if !ok {
		http.NotFound(c.W, c.R)
		return
	}
	var target struct {
		PostID int `gorm:"column:post_id"`
		UserID int `gorm:"column:user_id"`
	}
	if err := c.DB.Table("replies").Select("post_id, user_id").Where("id = ?", replyID).Take(&target).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.NotFound(c.W, c.R)
			return
		}
		http.Error(c.W, "internal server error", http.StatusInternalServerError)
		return
	}
	if target.PostID != postID {
		http.NotFound(c.W, c.R)
		return
	}
	if target.UserID != user.ID && !user.IsSysop() {
		c.JSONError(http.StatusForbidden, "Forbidden")
		return
	}
	if err := c.DB.Table("replies").Where("id = ?", replyID).Delete(nil).Error; err != nil {
		http.Error(c.W, "internal server error", http.StatusInternalServerError)
		return
	}
	c.JSON(map[string]bool{"ok": true})
}

func parseID(r *http.Request, key string) (int, bool) {
	id, err := strconv.Atoi(r.PathValue(key))
	if err != nil || id < 1 {
		return 0, false
	}
	return id, true
}
