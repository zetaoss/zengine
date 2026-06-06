package articletpl

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/models"
	"github.com/zetaoss/zengine/goapp/server/serverctx"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const articleTplKey = "article-tpl"

func GetArticleTpl(c *serverctx.Context) {
	enabled, err := findArticleTplEnabledIDs(c.DB)
	if err != nil {
		http.Error(c.W, "internal server error", http.StatusInternalServerError)
		return
	}
	c.JSON(app.H{"key": articleTplKey, "value": enabled})
}

type enabledArticleTpl struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

func GetEnabledArticleTpl(c *serverctx.Context) {
	enabledIDs, err := findArticleTplEnabledIDs(c.DB)
	if err != nil {
		http.Error(c.W, "internal server error", http.StatusInternalServerError)
		return
	}
	if len(enabledIDs) == 0 {
		c.JSON(app.H{"key": articleTplKey, "value": []enabledArticleTpl{}})
		return
	}

	items, err := fetchEnabledArticleTplDetails(c, enabledIDs)
	if err != nil {
		http.Error(c.W, "internal server error", http.StatusInternalServerError)
		return
	}
	c.JSON(app.H{"key": articleTplKey, "value": items})
}

func PutArticleTpl(c *serverctx.Context) {
	user, ok := c.User()
	if !ok || user.ID < 1 {
		c.JSONError(http.StatusUnauthorized, "Unauthenticated")
		return
	}
	if !user.IsSysop() {
		c.JSONError(http.StatusForbidden, "Forbidden")
		return
	}

	rawBody, err := io.ReadAll(c.R.Body)
	if err != nil {
		http.Error(c.W, "internal server error", http.StatusInternalServerError)
		return
	}

	enabled, ok, decodeErr := decodeEnabled(rawBody)
	if !ok {
		slog.Warn("article-tpl invalid payload", "err", decodeErr, "raw", string(rawBody))
		c.JSONError(http.StatusUnprocessableEntity, "Invalid payload")
		return
	}
	enabled = normalizeIDs(enabled)

	valueBytes, err := json.Marshal(enabled)
	if err != nil {
		http.Error(c.W, "internal server error", http.StatusInternalServerError)
		return
	}

	row := models.AppSetting{
		SettingKey: articleTplKey,
		ValueJSON:  string(valueBytes),
	}
	err = c.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "setting_key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value_json"}),
	}).Create(&row).Error
	if err != nil {
		if isTableMissingError(err) {
			c.JSONError(http.StatusServiceUnavailable, "app_settings table is not ready")
			return
		}
		http.Error(c.W, "internal server error", http.StatusInternalServerError)
		return
	}
	c.JSON(app.H{"ok": true, "key": articleTplKey, "value": enabled})
}

func decodeEnabled(raw []byte) ([]int, bool, string) {
	var arr []int
	errArr := json.Unmarshal(raw, &arr)
	if errArr == nil {
		return arr, true, ""
	}

	var obj struct {
		Enabled []int `json:"enabled"`
	}
	errObj := json.Unmarshal(raw, &obj)
	if errObj == nil {
		return obj.Enabled, true, ""
	}

	return nil, false, "array decode: " + errArr.Error() + "; object decode: " + errObj.Error()
}

func findArticleTplEnabledIDs(db *gorm.DB) ([]int, error) {
	var row models.AppSetting
	if err := db.Where("setting_key = ?", articleTplKey).Take(&row).Error; err != nil {
		if isTableMissingError(err) {
			return []int{}, nil
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []int{}, nil
		}
		return nil, err
	}

	var enabled []int
	if err := json.Unmarshal([]byte(row.ValueJSON), &enabled); err != nil {
		return []int{}, nil
	}
	return normalizeIDs(enabled), nil
}

func isTableMissingError(err error) bool {
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "doesn't exist") && strings.Contains(msg, "app_settings")
}

func normalizeIDs(ids []int) []int {
	seen := make(map[int]struct{}, len(ids))
	out := make([]int, 0, len(ids))
	for _, id := range ids {
		if id < 1 {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

func fetchEnabledArticleTplDetails(c *serverctx.Context, enabledIDs []int) ([]enabledArticleTpl, error) {
	apiServer := strings.TrimRight(c.Config().App.APIServer, "/")
	if apiServer == "" {
		return nil, fmt.Errorf("API_SERVER is required")
	}

	pageIDs := make([]string, 0, len(enabledIDs))
	for _, id := range enabledIDs {
		pageIDs = append(pageIDs, strconv.Itoa(id))
	}

	u, err := url.Parse(apiServer + "/w/api.php")
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("action", "query")
	q.Set("format", "json")
	q.Set("formatversion", "2")
	q.Set("prop", "revisions")
	q.Set("rvprop", "content")
	q.Set("rvslots", "main")
	q.Set("pageids", strings.Join(pageIDs, "|"))
	u.RawQuery = q.Encode()

	req, _ := http.NewRequestWithContext(c.R.Context(), http.MethodGet, u.String(), nil)
	resp, err := (&http.Client{Timeout: 20 * time.Second}).Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("mediawiki api failed: status=%d", resp.StatusCode)
	}

	payload, err := decodeMediaWikiPayload(resp)
	if err != nil {
		return nil, err
	}

	pageByID := make(map[int]enabledArticleTpl, len(payload.Query.Pages))
	for _, page := range payload.Query.Pages {
		if page.PageID < 1 || page.Missing {
			continue
		}
		if !strings.HasPrefix(page.Title, "틀:새문서틀") {
			continue
		}
		pageByID[page.PageID] = enabledArticleTpl{
			ID:      page.PageID,
			Title:   page.Title,
			Content: revisionContent(page.Revisions),
		}
	}

	out := make([]enabledArticleTpl, 0, len(enabledIDs))
	for _, id := range enabledIDs {
		if item, ok := pageByID[id]; ok {
			out = append(out, item)
		}
	}
	return out, nil
}

type mediaWikiPayload struct {
	Query struct {
		Pages []struct {
			PageID    int    `json:"pageid"`
			Title     string `json:"title"`
			Missing   bool   `json:"missing"`
			Revisions []struct {
				Content string `json:"content"`
				Slots   struct {
					Main struct {
						Content string `json:"content"`
					} `json:"main"`
				} `json:"slots"`
			} `json:"revisions"`
		} `json:"pages"`
	} `json:"query"`
}

func decodeMediaWikiPayload(resp *http.Response) (*mediaWikiPayload, error) {
	var payload mediaWikiPayload
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	return &payload, nil
}

func revisionContent(revisions []struct {
	Content string `json:"content"`
	Slots   struct {
		Main struct {
			Content string `json:"content"`
		} `json:"main"`
	} `json:"slots"`
}) string {
	if len(revisions) == 0 {
		return ""
	}
	if revisions[0].Slots.Main.Content != "" {
		return revisions[0].Slots.Main.Content
	}
	return revisions[0].Content
}
