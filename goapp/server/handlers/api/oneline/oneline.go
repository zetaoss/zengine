package oneline

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"gorm.io/gorm"
	"github.com/zetaoss/zengine/goapp/server/middleware"
	"github.com/zetaoss/zengine/goapp/server/paginator"
	"github.com/zetaoss/zengine/goapp/server/serverctx"
)

type row struct {
	ID       int    `json:"id" gorm:"column:id"`
	UserID   int    `json:"user_id" gorm:"column:user_id"`
	UserName string `json:"user_name" gorm:"column:user_name"`
	Created  string `json:"created" gorm:"column:created"`
	Message  string `json:"message" gorm:"column:message"`
}

func Recent(c *serverctx.Context) {
	rows := make([]row, 0, 15)
	if err := c.DB.Table("onelines").Order("id DESC").Limit(15).Find(&rows).Error; err != nil {
		http.Error(c.W, "internal server error", http.StatusInternalServerError)
		return
	}
	c.JSON(rows)
}

func Index(c *serverctx.Context) {
	const perPage = 30
	rows := make([]row, 0, perPage)
	payload, err := paginator.Paginate(c.R, c.DB.Table("onelines").Order("id DESC"), perPage, &rows)
	if err != nil {
		http.Error(c.W, "internal server error", http.StatusInternalServerError)
		return
	}
	c.JSON(payload)
}

func Store(c *serverctx.Context) {
	user, ok := middleware.UserFromRequest(c.R)
	if !ok {
		c.JSONError(http.StatusUnauthorized, "Unauthenticated")
		return
	}
	var body struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(c.R.Body).Decode(&body); err != nil {
		c.JSONError(http.StatusUnprocessableEntity, "Invalid payload")
		return
	}
	if len(body.Message) < 1 || len(body.Message) > 5000 {
		c.JSONError(http.StatusUnprocessableEntity, "The message field is invalid")
		return
	}

	newRow := row{UserID: user.ID, UserName: user.Name, Created: time.Now().Format("2006-01-02 15:04:05"), Message: body.Message}
	if err := c.DB.Table("onelines").Create(&newRow).Error; err != nil {
		http.Error(c.W, "internal server error", http.StatusInternalServerError)
		return
	}
	c.JSON(newRow)
}

func Destroy(c *serverctx.Context) {
	user, ok := middleware.UserFromRequest(c.R)
	if !ok {
		c.JSONError(http.StatusUnauthorized, "Unauthenticated")
		return
	}
	id, err := strconv.Atoi(c.R.PathValue("id"))
	if err != nil || id < 1 {
		http.NotFound(c.W, c.R)
		return
	}
	var target row
	if err := c.DB.Table("onelines").Where("id = ?", id).First(&target).Error; err != nil {
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
	if err := c.DB.Table("onelines").Where("id = ?", id).Delete(nil).Error; err != nil {
		http.Error(c.W, "internal server error", http.StatusInternalServerError)
		return
	}
	c.JSON(map[string]bool{"ok": true})
}
