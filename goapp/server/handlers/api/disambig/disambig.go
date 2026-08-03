package disambig

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/zetaoss/zengine/goapp/server/serverctx"
)

type DisambigNode struct {
	Title       string `json:"title"`
	Text        string `json:"text"`
	Description string `json:"description"`
	Href        string `json:"href"`
	ID          int    `json:"id,omitempty"`
	New         int    `json:"new,omitempty"`
}

type DisambigCache struct {
	ID    int            `json:"id"`
	Text  string         `json:"text"`
	Nodes []DisambigNode `json:"nodes"`
}

type DisambigItem struct {
	ID        int            `json:"id" gorm:"column:id"`
	Title     string         `json:"title" gorm:"column:title"`
	Entries   int            `json:"entries" gorm:"column:entries"`
	CacheRaw  string         `json:"-" gorm:"column:cache"`
	Nodes     []DisambigNode `json:"nodes" gorm:"-"`
	CreatedAt time.Time      `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"column:updated_at"`
}

type DisambigResponse struct {
	Data       []DisambigItem `json:"data"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	Limit      int            `json:"limit"`
	TotalPages int            `json:"total_pages"`
}

func Index(c *serverctx.Context) {
	page, _ := strconv.Atoi(c.R.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(c.R.URL.Query().Get("limit"))
	if limit < 1 {
		limit = 25
	} else if limit > 100 {
		limit = 100
	}

	sort := strings.ToLower(c.R.URL.Query().Get("sort"))
	order := strings.ToUpper(c.R.URL.Query().Get("order"))
	if order != "DESC" {
		order = "ASC"
	}

	var orderCol string
	switch sort {
	case "id":
		orderCol = "d.id"
	case "entries":
		orderCol = "d.entries"
	case "created":
		orderCol = "d.created_at"
	case "updated":
		orderCol = "d.updated_at"
	default:
		orderCol = "p.page_title"
	}

	var total int64
	if err := c.DB.Table("ldb.disambigs AS d").Count(&total).Error; err != nil {
		http.Error(c.W, "internal server error", http.StatusInternalServerError)
		return
	}

	page, totalPages := paginationBounds(page, limit, total)

	offset := (page - 1) * limit
	rows := make([]DisambigItem, 0, limit)
	err := c.DB.Table("ldb.disambigs AS d").
		Select("d.id, COALESCE(p.page_title, '') AS title, d.entries, d.cache, d.created_at, d.updated_at").
		Joins("LEFT JOIN zetawiki.page AS p ON d.id = p.page_id").
		Order(orderCol + " " + order).
		Order("d.id ASC").
		Offset(offset).
		Limit(limit).
		Find(&rows).Error
	if err != nil {
		http.Error(c.W, "internal server error", http.StatusInternalServerError)
		return
	}

	for i := range rows {
		rows[i].Nodes = make([]DisambigNode, 0)
		if rows[i].CacheRaw != "" {
			var parsed DisambigCache
			if err := json.Unmarshal([]byte(rows[i].CacheRaw), &parsed); err == nil && parsed.Nodes != nil {
				rows[i].Nodes = parsed.Nodes
			}
		}
	}

	c.JSON(DisambigResponse{
		Data:       rows,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	})
}

func paginationBounds(page, limit int, total int64) (int, int) {
	totalPages := 1
	if total > 0 {
		totalPages = int((total + int64(limit) - 1) / int64(limit))
	}
	if page > totalPages {
		page = totalPages
	}
	return page, totalPages
}
