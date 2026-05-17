package runbox

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/jobs/runboxjob"
	"github.com/zetaoss/zengine/goapp/server/serverctx"

	"gorm.io/gorm"
)

func Show(c *serverctx.Context) {
	hash := strings.TrimSpace(c.R.PathValue("hash"))
	if hash == "" {
		http.NotFound(c.W, c.R)
		return
	}

	var row app.H
	if err := c.DB.Table("runboxes").
		Select("hash, phase, user_id, page_id, type, outs, cpu, mem, time").
		Where("hash = ?", hash).
		Take(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(map[string]string{"phase": "none"})
			return
		}
		http.Error(c.W, "internal server error", http.StatusInternalServerError)
		return
	}
	c.JSON(row)
}

func Store(c *serverctx.Context) {
	var body struct {
		Hash    string `json:"hash"`
		UserID  int    `json:"user_id"`
		PageID  int    `json:"page_id"`
		Type    string `json:"type"`
		Payload app.H  `json:"payload"`
	}
	if err := json.NewDecoder(c.R.Body).Decode(&body); err != nil {
		c.JSONError(http.StatusUnprocessableEntity, "Invalid payload")
		return
	}
	if body.Hash == "" || body.PageID < 1 || (body.Type != "lang" && body.Type != "notebook") || body.Payload == nil {
		c.JSONError(http.StatusUnprocessableEntity, "Invalid payload")
		return
	}

	insert := app.H{
		"hash":       body.Hash,
		"phase":      "pending",
		"user_id":    0,
		"page_id":    body.PageID,
		"type":       body.Type,
		"payload":    toJSON(body.Payload),
		"created_at": time.Now(),
		"updated_at": time.Now(),
	}
	_ = c.DB.Table("runboxes").Where("hash = ?", body.Hash).Delete(nil).Error
	if err := c.DB.Table("runboxes").Create(insert).Error; err != nil {
		http.Error(c.W, "internal server error", http.StatusInternalServerError)
		return
	}

	if _, err := runboxjob.Enqueue(context.Background(), c.AppContext, body.Hash); err != nil {
		http.Error(c.W, "internal server error", http.StatusInternalServerError)
		return
	}
	c.JSONStatus(http.StatusAccepted, map[string]string{"phase": "pending"})
}

func Rerun(c *serverctx.Context) {
	hash := strings.TrimSpace(c.R.PathValue("hash"))
	if hash == "" {
		http.NotFound(c.W, c.R)
		return
	}
	var row struct {
		Hash string `gorm:"column:hash"`
	}
	if err := c.DB.Table("runboxes").Select("hash").Where("hash = ?", hash).Take(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.NotFound(c.W, c.R)
			return
		}
		http.Error(c.W, "internal server error", http.StatusInternalServerError)
		return
	}
	_ = c.DB.Table("runboxes").Where("hash = ?", hash).Updates(app.H{"phase": "pending", "updated_at": time.Now()}).Error
	if _, err := runboxjob.Enqueue(context.Background(), c.AppContext, hash); err != nil {
		http.Error(c.W, "internal server error", http.StatusInternalServerError)
		return
	}
	c.JSON(map[string]bool{"ok": true})
}

func toJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
