package pagereaction

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/server/serverctx"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var emojis = []string{"👍", "❤️", "❤", "🎉", "🚀", "👀", "😄", "😆"}
var codes = []int{128077, 10084, 10084, 127881, 128640, 128064, 128516, 128516}

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
	if user, ok := c.User(); ok && user.ID > 0 {
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

	c.JSON(app.H{"emojiCount": emojiCount, "userEmojis": userEmojis})
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

	total := 0
	ec := map[string]int{}

	err := c.DB.Transaction(func(tx *gorm.DB) error {
		if *body.Enable {
			row := app.H{"page_id": body.PageID, "user_id": user.ID, "emoji_code": emojiCode}
			if err := tx.Table("page_reaction_users").Clauses(clause.OnConflict{DoNothing: true}).Create(row).Error; err != nil {
				return err
			}
		} else {
			if err := tx.Table("page_reaction_users").Where("page_id = ? AND user_id = ? AND emoji_code = ?", body.PageID, user.ID, emojiCode).Delete(nil).Error; err != nil {
				return err
			}
		}

		rows := make([]struct {
			EmojiCode int `gorm:"column:emoji_code"`
			Cnt       int `gorm:"column:cnt"`
		}, 0, 8)
		if err := tx.Table("page_reaction_users").
			Select("emoji_code, COUNT(*) as cnt").
			Where("page_id = ?", body.PageID).
			Group("emoji_code").
			Find(&rows).Error; err != nil {
			return err
		}

		for _, r := range rows {
			em := code2emoji(r.EmojiCode)
			if em == "" {
				continue
			}
			ec[em] = r.Cnt
			total += r.Cnt
		}
		ecRaw, err := json.Marshal(ec)
		if err != nil {
			return err
		}

		upsert := app.H{"page_id": body.PageID, "cnt": total, "emoji_count": string(ecRaw)}
		return tx.Table("page_reactions").Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "page_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"cnt", "emoji_count"}),
		}).Create(upsert).Error
	})

	if err != nil {
		c.JSONError(http.StatusInternalServerError, "Failed to save reaction")
		return
	}

	c.JSON(app.H{"ok": true, "emojiCount": ec, "total": total})
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
