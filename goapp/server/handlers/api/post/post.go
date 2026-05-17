package post

import (
	"net/http"
	"strings"
	"time"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/models"
	"github.com/zetaoss/zengine/goapp/server/middleware"
	"github.com/zetaoss/zengine/goapp/server/paginator"
	"github.com/zetaoss/zengine/goapp/server/serverctx"

	"gorm.io/gorm"
)

const perPage = 15

var allowedCats = map[string]struct{}{
	"질문": {},
	"잡담": {},
	"인사": {},
	"기타": {},
}

func Recent(c *serverctx.Context) {
	rows := make([]models.ForumPost, 0, 6)
	q := baseQuery(c.DB).Order("posts.id DESC")
	if err := q.Limit(6).Find(&rows).Error; err != nil {
		c.InternalError()
		return
	}
	fillTagNames(rows)
	c.JSON(rows)
}

func Index(c *serverctx.Context) {
	rows := make([]models.ForumPost, 0, perPage)
	payload, err := paginator.Paginate(c.R, baseQuery(c.DB).Order("posts.id DESC"), perPage, &rows)
	if err != nil {
		c.InternalError()
		return
	}
	fillTagNames(rows)
	c.JSON(payload)
}

func Show(c *serverctx.Context) {
	id, ok := c.PathInt("post")
	if !ok {
		c.NotFound()
		return
	}
	var row models.ForumPost
	if err := baseQuery(c.DB).Where("posts.id = ?", id).Take(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.NotFound()
			return
		}
		c.InternalError()
		return
	}
	_ = c.DB.Table("posts").Where("id = ?", id).UpdateColumn("hit", gorm.Expr("hit + 1")).Error
	row.Hit++
	row.TagNames = splitTags(row.TagsStr)
	c.JSON(row)
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
	var body struct {
		Cat   string `json:"cat"`
		Title string `json:"title"`
		Body  string `json:"body"`
	}
	if !c.Decode(&body) {
		return
	}
	if !validCat(body.Cat) || strings.TrimSpace(body.Title) == "" || len(body.Title) > 100 || strings.TrimSpace(body.Body) == "" || len(body.Body) > 5000 {
		c.JSONError(http.StatusUnprocessableEntity, "Invalid payload")
		return
	}
	insert := app.H{
		"cat":        body.Cat,
		"title":      body.Title,
		"body":       body.Body,
		"tags_str":   "",
		"channel_id": 1,
		"user_id":    user.ID,
		"user_name":  user.Name,
		"hit":        0,
		"is_notice":  0,
		"created_at": time.Now(),
		"updated_at": time.Now(),
	}
	if err := c.DB.Table("posts").Create(insert).Error; err != nil {
		c.InternalError()
		return
	}
	var row models.ForumPost
	if err := baseQuery(c.DB).Order("posts.id DESC").Take(&row).Error; err != nil {
		c.InternalError()
		return
	}
	row.TagNames = splitTags(row.TagsStr)
	c.JSON(row)
}

func Update(c *serverctx.Context) {
	user, ok := middleware.UserFromRequest(c.R)
	if !ok || user.ID < 1 {
		c.JSONError(http.StatusUnauthorized, "Unauthenticated")
		return
	}
	id, ok := c.PathInt("post")
	if !ok {
		c.NotFound()
		return
	}
	var target struct {
		UserID int `gorm:"column:user_id"`
	}
	if err := c.DB.Table("posts").Select("user_id").Where("id = ?", id).Take(&target).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.NotFound()
			return
		}
		c.InternalError()
		return
	}
	if target.UserID != user.ID {
		c.JSONError(http.StatusForbidden, "Forbidden")
		return
	}
	var body struct {
		Cat   string `json:"cat"`
		Title string `json:"title"`
		Body  string `json:"body"`
	}
	if !c.Decode(&body) {
		return
	}
	if !validCat(body.Cat) || strings.TrimSpace(body.Title) == "" || len(body.Title) > 100 || strings.TrimSpace(body.Body) == "" || len(body.Body) > 5000 {
		c.JSONError(http.StatusUnprocessableEntity, "Invalid payload")
		return
	}
	if err := c.DB.Table("posts").Where("id = ?", id).Updates(app.H{
		"cat":        body.Cat,
		"title":      body.Title,
		"body":       body.Body,
		"tags_str":   "",
		"channel_id": 1,
		"updated_at": time.Now(),
	}).Error; err != nil {
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
	id, ok := c.PathInt("post")
	if !ok {
		c.NotFound()
		return
	}
	var target struct {
		UserID int `gorm:"column:user_id"`
	}
	if err := c.DB.Table("posts").Select("user_id").Where("id = ?", id).Take(&target).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.NotFound()
			return
		}
		c.InternalError()
		return
	}
	if target.UserID != user.ID && !user.IsSysop() {
		c.JSONError(http.StatusForbidden, "Forbidden")
		return
	}
	if err := c.DB.Table("posts").Where("id = ?", id).Delete(nil).Error; err != nil {
		c.InternalError()
		return
	}
	c.JSON(map[string]bool{"ok": true})
}

func baseQuery(db *gorm.DB) *gorm.DB {
	return db.Table("posts").
		Select("posts.*, COALESCE((SELECT COUNT(*) FROM replies WHERE replies.post_id = posts.id), 0) AS replies_count")
}

func validCat(cat string) bool {
	_, ok := allowedCats[cat]
	return ok
}

func fillTagNames(rows []models.ForumPost) {
	for i := range rows {
		rows[i].TagNames = splitTags(rows[i].TagsStr)
	}
}

func splitTags(s string) []string {
	if s == "" {
		return []string{}
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		t := strings.TrimSpace(p)
		if t != "" {
			out = append(out, t)
		}
	}
	return out
}
