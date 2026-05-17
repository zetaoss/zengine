package pagereaction

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/server/middleware"
	"github.com/zetaoss/zengine/goapp/server/serverctx"

	"gorm.io/gorm"
)

var emojis = []string{"👍", "😆", "😢", "😮", "❤️", "❤"}
var codes = []int{128077, 128518, 128546, 128558, 10084, 10084}

func Show(c *serverctx.Context) {
	pageID, err := strconv.Atoi(c.R.PathValue("page"))
	if err != nil || pageID < 1 {
		http.NotFound(c.W, c.R)
		return
	}

	emojiCount := map[string]int{}
	var pr struct {
		EmojiCount string `gorm:"column:emoji_count"`
	}
	if err := c.DB.Table("page_reactions").Select("emoji_count").Where("page_id = ?", pageID).Take(&pr).Error; err == nil {
		_ = json.Unmarshal([]byte(pr.EmojiCount), &emojiCount)
	}

	userEmojis := []string{}
	if user, ok := middleware.UserFromRequest(c.R); ok && user.ID > 0 {
		rows := make([]struct {
			EmojiCode int `gorm:"column:emoji_code"`
		}, 0, 8)
		_ = c.DB.Table("page_reaction_users").Select("emoji_code").Where("page_id = ? AND user_id = ?", pageID, user.ID).Find(&rows).Error
		for _, r := range rows {
			if em := code2emoji(r.EmojiCode); em != "" {
				userEmojis = append(userEmojis, em)
			}
		}
	}

	c.JSON([]app.H{{"emojiCount": emojiCount, "userEmojis": userEmojis}})
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
		PageID int    `json:"pageid"`
		Emoji  string `json:"emoji"`
		Enable *bool  `json:"enable"`
	}
	if err := json.NewDecoder(c.R.Body).Decode(&body); err != nil || body.PageID < 1 || body.Enable == nil {
		c.JSONError(http.StatusUnprocessableEntity, "Invalid payload")
		return
	}
	emojiCode := emoji2code(body.Emoji)
	if emojiCode <= 0 {
		c.JSONError(http.StatusUnprocessableEntity, "Invalid emoji")
		return
	}

	row := app.H{"page_id": body.PageID, "user_id": user.ID, "emoji_code": emojiCode}
	if *body.Enable {
		_ = c.DB.Table("page_reaction_users").Where(row).FirstOrCreate(&app.H{}).Error
	} else {
		_ = c.DB.Table("page_reaction_users").Where(row).Delete(nil).Error
	}

	rows := make([]struct {
		EmojiCode int `gorm:"column:emoji_code"`
		Cnt       int `gorm:"column:cnt"`
	}, 0, 8)
	_ = c.DB.Table("page_reaction_users").
		Select("emoji_code, COUNT(*) as cnt").
		Where("page_id = ?", body.PageID).
		Group("emoji_code").
		Find(&rows).Error

	total := 0
	ec := map[string]int{}
	for _, r := range rows {
		em := code2emoji(r.EmojiCode)
		if em == "" {
			continue
		}
		ec[em] = r.Cnt
		total += r.Cnt
	}
	ecRaw, _ := json.Marshal(ec)

	upsert := app.H{"page_id": body.PageID, "cnt": total, "emoji_count": string(ecRaw)}
	var pr struct {
		PageID int `gorm:"column:page_id"`
	}
	err := c.DB.Table("page_reactions").Select("page_id").Where("page_id = ?", body.PageID).Take(&pr).Error
	if err == gorm.ErrRecordNotFound {
		_ = c.DB.Table("page_reactions").Create(upsert).Error
	} else {
		_ = c.DB.Table("page_reactions").Where("page_id = ?", body.PageID).Updates(upsert).Error
	}

	c.JSON(map[string]bool{"ok": true})
}

func emoji2code(emoji string) int {
	for i, e := range emojis {
		if e == emoji {
			return codes[i]
		}
	}
	return 0
}

func code2emoji(code int) string {
	for i, c := range codes {
		if c == code {
			return emojis[i]
		}
	}
	return ""
}
