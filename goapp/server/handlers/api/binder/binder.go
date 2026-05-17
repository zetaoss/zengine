package binder

import (
	"encoding/json"
	"net/http"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/models"
	"github.com/zetaoss/zengine/goapp/server/middleware"
	"github.com/zetaoss/zengine/goapp/server/serverctx"

	"gorm.io/gorm"
)

const redirectMaxHops = 9

func Index(c *serverctx.Context) {
	rows := make([]models.Binder, 0, 1024)
	err := c.DB.Table("ldb.binders AS b").
		Select("b.id, COALESCE(p.page_title, '') AS title, b.docs, b.links, b.title_doc, b.enabled, b.created_at").
		Joins("LEFT JOIN zetawiki.page AS p ON b.id = p.page_id").
		Order("b.enabled DESC").
		Order("b.docs DESC").
		Order("p.page_title").
		Order("b.id").
		Find(&rows).Error
	if err != nil {
		http.Error(c.W, "internal server error", http.StatusInternalServerError)
		return
	}
	c.JSON(rows)
}

func Update(c *serverctx.Context) {
	user, ok := middleware.UserFromRequest(c.R)
	if !ok || user.ID < 1 {
		c.JSONError(http.StatusUnauthorized, "Unauthenticated")
		return
	}
	if !user.IsSysop() {
		c.JSONError(http.StatusForbidden, "Forbidden")
		return
	}

	binderID, ok := c.PathInt("binder")
	if !ok {
		c.NotFound()
		return
	}

	var body struct {
		Enabled *bool `json:"enabled"`
	}
	if err := json.NewDecoder(c.R.Body).Decode(&body); err != nil || body.Enabled == nil {
		c.JSONError(http.StatusUnprocessableEntity, "Invalid payload")
		return
	}

	var exists int64
	if err := c.DB.Table("ldb.binders").Where("id = ?", binderID).Count(&exists).Error; err != nil {
		http.Error(c.W, "internal server error", http.StatusInternalServerError)
		return
	}
	if exists == 0 {
		c.JSONError(http.StatusNotFound, "Binder not found")
		return
	}

	realBinderID := resolveRealBinderID(c.DB, binderID)
	updatedBinderID := binderID
	var deletedID *int
	var replacementTitle *string
	updatedEnabled := *body.Enabled

	if err := c.DB.Transaction(func(tx *gorm.DB) error {
		if realBinderID > 0 && realBinderID != binderID {
			deletedID = &binderID
			updatedBinderID = realBinderID
			replacementTitle = getPageTitle(tx, realBinderID)

			if binderExists(tx, realBinderID) {
				if err := deleteBinder(tx, binderID); err != nil {
					return err
				}
				updatedEnabled = getBinderEnabled(tx, realBinderID)
				return nil
			}
			return transferBinder(tx, binderID, realBinderID, *body.Enabled)
		}
		return tx.Table("ldb.binders").
			Where("id = ?", updatedBinderID).
			Update("enabled", boolToInt(*body.Enabled)).Error
	}); err != nil {
		http.Error(c.W, "internal server error", http.StatusInternalServerError)
		return
	}

	c.JSON(app.H{
		"ok":                true,
		"id":                updatedBinderID,
		"enabled":           updatedEnabled,
		"deleted_id":        deletedID,
		"replacement_title": replacementTitle,
	})
}

func resolveRealBinderID(db *gorm.DB, binderID int) int {
	currentID := binderID
	for i := 0; i < redirectMaxHops; i++ {
		var page struct {
			PageID int `gorm:"column:page_id"`
		}
		if err := db.Table("zetawiki.page").Select("page_id").Where("page_id = ?", currentID).Take(&page).Error; err != nil {
			return 0
		}

		var rd struct {
			Namespace int    `gorm:"column:rd_namespace"`
			Title     string `gorm:"column:rd_title"`
		}
		if err := db.Table("zetawiki.redirect").
			Select("rd_namespace, rd_title").
			Where("rd_from = ?", currentID).
			Take(&rd).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return currentID
			}
			return 0
		}

		var target struct {
			PageID int `gorm:"column:page_id"`
		}
		if err := db.Table("zetawiki.page").
			Select("page_id").
			Where("page_namespace = ? AND page_title = ?", rd.Namespace, rd.Title).
			Take(&target).Error; err != nil {
			return 0
		}
		currentID = target.PageID
	}
	return 0
}

func deleteBinder(db *gorm.DB, binderID int) error {
	if err := db.Table("ldb.binder_pages").Where("binder_id = ?", binderID).Delete(nil).Error; err != nil {
		return err
	}
	return db.Table("ldb.binders").Where("id = ?", binderID).Delete(nil).Error
}

func transferBinder(db *gorm.DB, fromBinderID int, toBinderID int, enabled bool) error {
	if err := db.Table("ldb.binder_pages").Where("binder_id = ?", fromBinderID).Update("binder_id", toBinderID).Error; err != nil {
		return err
	}
	return db.Table("ldb.binders").Where("id = ?", fromBinderID).Updates(app.H{
		"id":      toBinderID,
		"enabled": boolToInt(enabled),
	}).Error
}

func binderExists(db *gorm.DB, binderID int) bool {
	var n int64
	if err := db.Table("ldb.binders").Where("id = ?", binderID).Count(&n).Error; err != nil {
		return false
	}
	return n > 0
}

func getBinderEnabled(db *gorm.DB, binderID int) bool {
	var out struct {
		Enabled int `gorm:"column:enabled"`
	}
	if err := db.Table("ldb.binders").Select("enabled").Where("id = ?", binderID).Take(&out).Error; err != nil {
		return false
	}
	return out.Enabled == 1
}

func getPageTitle(db *gorm.DB, pageID int) *string {
	var out struct {
		Title string `gorm:"column:page_title"`
	}
	if err := db.Table("zetawiki.page").Select("page_title").Where("page_id = ?", pageID).Take(&out).Error; err != nil {
		return nil
	}
	return &out.Title
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
