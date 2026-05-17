package internalprofile

import (
	"net/http"
	"strconv"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/server/serverctx"

	"gorm.io/gorm"
)

func Show(c *serverctx.Context) {
	userID, err := strconv.Atoi(c.R.PathValue("userId"))
	if err != nil || userID < 1 {
		c.JSONError(http.StatusNotFound, "Not Found")
		return
	}
	var row struct {
		UserID   int     `gorm:"column:user_id"`
		UserName string  `gorm:"column:user_name"`
		T        *int    `gorm:"column:t"`
		GHash    *string `gorm:"column:ghash"`
	}
	err = c.DB.Table("zetawiki.user AS A").
		Joins("LEFT JOIN zetawiki.user_profiles AS B ON A.user_id = B.user_id").
		Select("A.user_id, A.user_name, B.t, B.ghash").
		Where("A.user_id = ?", userID).
		Take(&row).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSONError(http.StatusNotFound, "Not Found")
			return
		}
		http.Error(c.W, "internal server error", http.StatusInternalServerError)
		return
	}
	t := 1
	if row.T != nil {
		t = *row.T
	}
	gh := ""
	if row.GHash != nil {
		gh = *row.GHash
	}
	c.JSON(app.H{
		"name":  row.UserName,
		"t":     t,
		"ghash": gh,
	})
}
