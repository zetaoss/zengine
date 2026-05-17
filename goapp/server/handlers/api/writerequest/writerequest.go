package writerequest

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/models"
	"github.com/zetaoss/zengine/goapp/server/middleware"
	"github.com/zetaoss/zengine/goapp/server/paginator"
	"github.com/zetaoss/zengine/goapp/server/serverctx"

	"gorm.io/gorm"
)

func Count(c *serverctx.Context) {
	type countRow struct {
		Done int `json:"done" gorm:"column:done"`
		Todo int `json:"todo" gorm:"column:todo"`
	}
	var out countRow
	err := c.DB.Table("write_requests").
		Select("COUNT(CASE WHEN writed_at IS NOT NULL THEN 1 ELSE NULL END) AS done, COUNT(CASE WHEN writed_at IS NULL THEN 1 ELSE NULL END) AS todo").
		Take(&out).Error
	if err != nil {
		http.Error(c.W, "internal server error", http.StatusInternalServerError)
		return
	}
	c.JSON(out)
}

func IndexTodo(c *serverctx.Context) {
	listPage(c, "todo")
}

func IndexTodoTop(c *serverctx.Context) {
	listPage(c, "todo-top")
}

func IndexDone(c *serverctx.Context) {
	listPage(c, "done")
}

func Store(c *serverctx.Context) {
	user, ok := middleware.UserFromRequest(c.R)
	if !ok || user.ID < 1 {
		c.JSONError(http.StatusUnauthorized, "Unauthenticated")
		return
	}
	var body struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(c.R.Body).Decode(&body); err != nil {
		c.JSONError(http.StatusUnprocessableEntity, "Invalid payload")
		return
	}
	title := strings.TrimSpace(body.Title)
	if title == "" || len(title) > 255 {
		c.JSONError(http.StatusUnprocessableEntity, "The title field is invalid")
		return
	}
	newRow := app.H{
		"user_id":    user.ID,
		"user_name":  user.Name,
		"title":      title,
		"created_at": time.Now(),
		"updated_at": time.Now(),
	}
	if err := c.DB.Table("write_requests").Create(newRow).Error; err != nil {
		http.Error(c.W, "internal server error", http.StatusInternalServerError)
		return
	}
	var created struct {
		ID int `gorm:"column:id"`
	}
	if err := c.DB.Table("write_requests").Select("id").Where("user_id = ? AND title = ?", user.ID, title).Order("id DESC").Take(&created).Error; err != nil {
		http.Error(c.W, "internal server error", http.StatusInternalServerError)
		return
	}
	c.JSON(app.H{"ok": true, "id": created.ID})
}

func Recommend(c *serverctx.Context) {
	id, ok := parseID(c.R)
	if !ok {
		http.NotFound(c.W, c.R)
		return
	}
	if err := c.DB.Table("write_requests").Where("id = ?", id).UpdateColumn("rate", gorm.Expr("rate + 1")).Error; err != nil {
		http.Error(c.W, "internal server error", http.StatusInternalServerError)
		return
	}
	var out struct {
		Rate int `gorm:"column:rate"`
	}
	if err := c.DB.Table("write_requests").Select("rate").Where("id = ?", id).Take(&out).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.NotFound(c.W, c.R)
			return
		}
		http.Error(c.W, "internal server error", http.StatusInternalServerError)
		return
	}
	c.JSON(app.H{"ok": true, "rate": out.Rate})
}

func Destroy(c *serverctx.Context) {
	user, ok := middleware.UserFromRequest(c.R)
	if !ok || user.ID < 1 {
		c.JSONError(http.StatusUnauthorized, "Unauthenticated")
		return
	}
	id, ok := parseID(c.R)
	if !ok {
		http.NotFound(c.W, c.R)
		return
	}
	var target struct {
		UserID int `gorm:"column:user_id"`
	}
	if err := c.DB.Table("write_requests").Select("user_id").Where("id = ?", id).Take(&target).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.NotFound(c.W, c.R)
			return
		}
		http.Error(c.W, "internal server error", http.StatusInternalServerError)
		return
	}
	if user.ID != target.UserID && !user.IsSysop() {
		c.JSONError(http.StatusForbidden, "Forbidden")
		return
	}
	if err := c.DB.Table("write_requests").Where("id = ?", id).Delete(nil).Error; err != nil {
		http.Error(c.W, "internal server error", http.StatusInternalServerError)
		return
	}
	c.JSON(app.H{"ok": true})
}

func listPage(c *serverctx.Context, mode string) {
	const perPage = 25

	base := c.DB.Table("write_requests as w").
		Select("w.*, w.user_name as user_name, (SELECT COALESCE(n.hit, 0) FROM not_matches n WHERE n.title = w.title LIMIT 1) as hit")

	countQ := c.DB.Table("write_requests as w")
	switch mode {
	case "done":
		base = base.Where("w.writed_at IS NOT NULL")
		countQ = countQ.Where("w.writed_at IS NOT NULL")
		base = base.Order("w.writed_at DESC").Order("w.id DESC")
	case "todo-top":
		base = base.Where("w.writed_at IS NULL")
		countQ = countQ.Where("w.writed_at IS NULL")
		base = base.Order("w.rate DESC, hit DESC, w.ref DESC, w.id DESC")
	default:
		base = base.Where("w.writed_at IS NULL")
		countQ = countQ.Where("w.writed_at IS NULL")
		base = base.Order("w.id DESC")
	}

	rows := make([]models.WriteRequest, 0, perPage)
	payload, err := paginator.PaginateWith(c.R, countQ, base, perPage, &rows)
	if err != nil {
		http.Error(c.W, "internal server error", http.StatusInternalServerError)
		return
	}
	c.JSON(payload)
}

func parseID(r *http.Request) (int, bool) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		return 0, false
	}
	return id, true
}
