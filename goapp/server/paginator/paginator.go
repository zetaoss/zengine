package paginator

import (
	"net/http"
	"strconv"
	"github.com/zetaoss/zengine/goapp/app"

	"gorm.io/gorm"
)

func pageFromRequest(r *http.Request) int {
	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	return page
}

func lastPage(total int64, perPage int) int {
	if perPage <= 0 {
		return 1
	}
	lastPage := int((total + int64(perPage) - 1) / int64(perPage))
	if lastPage < 1 {
		return 1
	}
	return lastPage
}

func offset(page int, perPage int) int {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		return 0
	}
	return (page - 1) * perPage
}

func payload(page int, lastPage int, data any) app.H {
	return app.H{
		"current_page": page,
		"last_page":    lastPage,
		"data":         data,
	}
}

func Paginate(r *http.Request, q *gorm.DB, perPage int, out any) (app.H, error) {
	return PaginateWith(r, q, q, perPage, out)
}

func PaginateWith(r *http.Request, countQ *gorm.DB, dataQ *gorm.DB, perPage int, out any) (app.H, error) {
	page := pageFromRequest(r)

	// Clone queries to avoid cross-contamination
	countQ = countQ.Session(&gorm.Session{})
	dataQ = dataQ.Session(&gorm.Session{})

	var total int64
	if err := countQ.Count(&total).Error; err != nil {
		return nil, err
	}
	lastPage := lastPage(total, perPage)

	if err := dataQ.Limit(perPage).Offset(offset(page, perPage)).Find(out).Error; err != nil {
		return nil, err
	}
	return payload(page, lastPage, out), nil
}
