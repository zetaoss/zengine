package reply

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/models"
	"github.com/zetaoss/zengine/goapp/server/serverctx"

	"gorm.io/gorm"
)

func Index(c *serverctx.Context) {
	postID, ok := c.PathInt("postId")
	if !ok {
		c.NotFound()
		return
	}
	rows := make([]models.ForumReply, 0, 128)
	if err := c.DB.Table("replies").Where("post_id = ?", postID).Order("created_at DESC").Find(&rows).Error; err != nil {
		c.InternalError()
		return
	}
	c.JSON(rows)
}

func Store(c *serverctx.Context) {
	user, ok := c.User()
	if !ok || user.ID < 1 {
		c.JSONError(http.StatusUnauthorized, "Unauthenticated")
		return
	}
	if user.IsBlocked() {
		c.JSONError(http.StatusForbidden, "Forbidden")
		return
	}
	postID, ok := c.PathInt("postId")
	if !ok {
		c.NotFound()
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
		c.InternalError()
		return
	}
	c.JSON(row)
}

func Update(c *serverctx.Context) {
	user, ok := c.User()
	if !ok || user.ID < 1 {
		c.JSONError(http.StatusUnauthorized, "Unauthenticated")
		return
	}
	postID, ok := c.PathInt("postId")
	if !ok {
		c.NotFound()
		return
	}
	replyID, ok := c.PathInt("id")
	if !ok {
		c.NotFound()
		return
	}

	var target struct {
		PostID int `gorm:"column:post_id"`
		UserID int `gorm:"column:user_id"`
	}
	if err := c.DB.Table("replies").Select("post_id, user_id").Where("id = ?", replyID).Take(&target).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.NotFound()
			return
		}
		c.InternalError()
		return
	}
	if target.PostID != postID {
		c.NotFound()
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
		c.InternalError()
		return
	}
	c.JSON(map[string]bool{"ok": true})
}

func Destroy(c *serverctx.Context) {
	user, ok := c.User()
	if !ok || user.ID < 1 {
		c.JSONError(http.StatusUnauthorized, "Unauthenticated")
		return
	}
	postID, ok := c.PathInt("postId")
	if !ok {
		c.NotFound()
		return
	}
	replyID, ok := c.PathInt("id")
	if !ok {
		c.NotFound()
		return
	}
	var target struct {
		PostID int `gorm:"column:post_id"`
		UserID int `gorm:"column:user_id"`
	}
	if err := c.DB.Table("replies").Select("post_id, user_id").Where("id = ?", replyID).Take(&target).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.NotFound()
			return
		}
		c.InternalError()
		return
	}
	if target.PostID != postID {
		c.NotFound()
		return
	}
	if target.UserID != user.ID && !user.IsSysop() {
		c.JSONError(http.StatusForbidden, "Forbidden")
		return
	}
	if err := c.DB.Table("replies").Where("id = ?", replyID).Delete(nil).Error; err != nil {
		c.InternalError()
		return
	}
	c.JSON(map[string]bool{"ok": true})
}
