package user

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/zetaoss/zengine/goapp/server/serverctx"
)

func Show(c *serverctx.Context) {
	userName := strings.ReplaceAll(strings.TrimSpace(c.R.PathValue("userName")), "_", " ")
	if userName == "" {
		http.NotFound(c.W, c.R)
		return
	}
	var row struct {
		UserID           int    `json:"user_id" gorm:"column:user_id"`
		UserName         string `json:"user_name" gorm:"column:user_name"`
		UserRegistration string `json:"user_registration" gorm:"column:user_registration"`
		UserEditcount    int    `json:"user_editcount" gorm:"column:user_editcount"`
	}
	if err := c.DB.Table("zetawiki.user").
		Select("user_id, user_name, user_registration, user_editcount").
		Where("user_name = ?", userName).
		Take(&row).Error; err != nil {
		http.NotFound(c.W, c.R)
		return
	}
	c.JSON(row)
}

func Stats(c *serverctx.Context) {
	userID, err := strconv.Atoi(c.R.PathValue("userId"))
	if err != nil || userID < 1 {
		http.NotFound(c.W, c.R)
		return
	}
	rows := make([]struct {
		Rev int    `gorm:"column:rev"`
		DT  string `gorm:"column:dt"`
	}, 0, 4096)
	err = c.DB.Table("zetawiki.revision as r").
		Joins("JOIN zetawiki.actor as a ON r.rev_actor = a.actor_id").
		Select("COUNT(*) as rev, DATE(r.rev_timestamp) as dt").
		Where("a.actor_user = ?", userID).
		Group("dt").
		Order("dt").
		Find(&rows).Error
	if err != nil {
		http.Error(c.W, "internal server error", http.StatusInternalServerError)
		return
	}
	out := map[string]int{}
	for _, row := range rows {
		out[row.DT] = row.Rev
	}
	c.JSON(out)
}
