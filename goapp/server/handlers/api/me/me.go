package me

import (
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/server/middleware"
	"github.com/zetaoss/zengine/goapp/server/serverctx"

	"gorm.io/gorm"
)

var ghashRE = regexp.MustCompile("^[0-9a-f]{64}$")

func Me(c *serverctx.Context) {
	user, ok := middleware.UserFromRequest(c.R)
	if !ok || user.ID < 1 {
		c.JSON(app.H{"me": nil})
		return
	}
	c.JSON(app.H{
		"me": app.H{
			"id":     user.ID,
			"name":   user.Name,
			"groups": user.Groups,
		},
	})
}

func GetAvatar(c *serverctx.Context) {
	user, ok := middleware.UserFromRequest(c.R)
	if !ok || user.ID < 1 {
		c.JSONError(http.StatusUnauthorized, "Unauthenticated")
		return
	}
	var row struct {
		T     *int    `gorm:"column:t"`
		GHint *string `gorm:"column:ghint"`
	}
	_ = c.DB.Table("zetawiki.user_profiles").Where("user_id = ?", user.ID).Take(&row).Error
	t := 1
	if row.T != nil {
		t = *row.T
	}
	ghint := ""
	if row.GHint != nil {
		ghint = *row.GHint
	}
	c.JSON(app.H{"t": t, "ghint": ghint})
}

func VerifyGravatar(c *serverctx.Context) {
	ghash := strings.ToLower(strings.TrimSpace(c.R.URL.Query().Get("ghash")))
	if !ghashRE.MatchString(ghash) {
		c.JSONStatus(http.StatusUnprocessableEntity, app.H{"ok": false})
		return
	}
	if !gravatarExists(ghash) {
		c.JSONStatus(http.StatusNotFound, app.H{"ok": false})
		return
	}
	c.JSON(app.H{"ok": true, "ghash": ghash})
}

func UpdateAvatar(c *serverctx.Context) {
	user, ok := middleware.UserFromRequest(c.R)
	if !ok || user.ID < 1 {
		c.JSONError(http.StatusUnauthorized, "Unauthenticated")
		return
	}
	var body struct {
		T     int    `json:"t"`
		GHash string `json:"ghash"`
		GHint string `json:"ghint"`
	}
	if !c.Decode(&body) {
		return
	}
	if body.T < 1 || body.T > 3 {
		c.JSONError(http.StatusUnprocessableEntity, "Invalid t")
		return
	}
	body.GHash = strings.ToLower(strings.TrimSpace(body.GHash))
	body.GHint = strings.TrimSpace(body.GHint)

	var current struct {
		UserID int     `gorm:"column:user_id"`
		T      int     `gorm:"column:t"`
		GHash  *string `gorm:"column:ghash"`
		GHint  *string `gorm:"column:ghint"`
	}
	err := c.DB.Table("zetawiki.user_profiles").Where("user_id = ?", user.ID).Take(&current).Error
	notFound := err == gorm.ErrRecordNotFound
	if err != nil && !notFound {
		c.InternalError()
		return
	}

	saveHash := ""
	saveHint := ""
	if body.T == 3 {
		if body.GHash != "" {
			if !ghashRE.MatchString(body.GHash) {
				c.JSONError(http.StatusUnprocessableEntity, "Invalid ghash")
				return
			}
			if body.GHint == "" {
				c.JSONError(http.StatusUnprocessableEntity, "ghint required for gravatar")
				return
			}
			if !gravatarExists(body.GHash) {
				c.JSONError(http.StatusBadRequest, "Gravatar not found")
				return
			}
			saveHash = body.GHash
			saveHint = body.GHint
		} else {
			existing := ""
			if current.GHash != nil {
				existing = strings.TrimSpace(*current.GHash)
			}
			if existing == "" {
				c.JSONError(http.StatusUnprocessableEntity, "gravatar not configured")
				return
			}
			saveHash = existing
			if current.GHint != nil {
				saveHint = strings.TrimSpace(*current.GHint)
			}
		}
	}

	updates := app.H{"t": body.T}
	if body.T == 3 {
		updates["ghash"] = saveHash
		updates["ghint"] = saveHint
	}

	if notFound {
		updates["user_id"] = user.ID
		if err := c.DB.Table("zetawiki.user_profiles").Create(updates).Error; err != nil {
			c.InternalError()
			return
		}
	} else {
		if err := c.DB.Table("zetawiki.user_profiles").Where("user_id = ?", user.ID).Updates(updates).Error; err != nil {
			c.InternalError()
			return
		}
	}

	avatar := app.H{"id": user.ID, "name": user.Name, "t": body.T, "ghash": saveHash}
	c.JSON(app.H{"avatar": avatar})
}

func gravatarExists(ghash string) bool {
	req, err := http.NewRequest(http.MethodHead, "https://www.gravatar.com/avatar/"+ghash+"?d=404", nil)
	if err != nil {
		return false
	}
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	return resp.StatusCode == http.StatusOK
}
