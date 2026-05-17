package commonreport

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/jobs/commonreportjob"
	"github.com/zetaoss/zengine/goapp/models"
	"github.com/zetaoss/zengine/goapp/server/middleware"
	"github.com/zetaoss/zengine/goapp/server/paginator"
	"github.com/zetaoss/zengine/goapp/server/serverctx"

	"gorm.io/gorm"
)

const perPage = 15

func Index(c *serverctx.Context) {
	reports := make([]models.CommonReport, 0, perPage)
	payload, err := paginator.Paginate(c.R, c.DB.Table("common_reports").Order("id DESC"), perPage, &reports)
	if err != nil {
		c.InternalError()
		return
	}

	if err := attachItems(c.DB, reports); err != nil {
		c.InternalError()
		return
	}
	c.JSON(payload)
}

func Show(c *serverctx.Context) {
	id, ok := c.PathInt("id")
	if !ok {
		c.NotFound()
		return
	}
	var report models.CommonReport
	if err := c.DB.Table("common_reports").Where("id = ?", id).Take(&report).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.NotFound()
			return
		}
		c.InternalError()
		return
	}
	items := make([]models.CommonReportItem, 0, 32)
	if err := c.DB.Table("common_report_items").Where("report_id = ?", id).Order("total DESC").Find(&items).Error; err != nil {
		c.InternalError()
		return
	}
	report.Items = items
	report.Total = sumTotal(items)
	c.JSON(report)
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
		Names []string `json:"names"`
	}
	if !c.Decode(&body) {
		return
	}
	names := normalizeNames(body.Names)
	if len(names) < 2 {
		c.JSONError(http.StatusUnprocessableEntity, "비교 대상을 2개 이상 입력해 주세요.")
		return
	}

	report, err := createReport(c.DB, names, user.ID, user.Name)
	if err != nil {
		c.InternalError()
		return
	}
	if _, err := commonreportjob.Enqueue(context.Background(), c.AppContext, report.ID); err != nil {
		slog.Error("common-report enqueue failed", "report_id", report.ID, "err", err)
	}
	c.JSON(report)
}

func Clone(c *serverctx.Context) {
	user, ok := middleware.UserFromRequest(c.R)
	if !ok || user.ID < 1 {
		c.JSONError(http.StatusUnauthorized, "Unauthenticated")
		return
	}
	if user.IsBlocked() {
		c.JSONError(http.StatusForbidden, "Forbidden")
		return
	}
	id, ok := c.PathInt("id")
	if !ok {
		c.NotFound()
		return
	}

	original, err := loadReportWithItems(c.DB, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.NotFound()
			return
		}
		c.InternalError()
		return
	}

	names := make([]string, 0, len(original.Items))
	for _, it := range original.Items {
		names = append(names, it.Name)
	}
	report, err := createReport(c.DB, names, user.ID, user.Name)
	if err != nil {
		c.InternalError()
		return
	}
	if _, err := commonreportjob.Enqueue(context.Background(), c.AppContext, report.ID); err != nil {
		slog.Error("common-report enqueue failed", "report_id", report.ID, "err", err)
	}
	c.JSON(report)
}

func Rerun(c *serverctx.Context) {
	user, ok := middleware.UserFromRequest(c.R)
	if !ok || user.ID < 1 {
		c.JSONError(http.StatusUnauthorized, "Unauthenticated")
		return
	}
	id, ok := c.PathInt("id")
	if !ok {
		c.NotFound()
		return
	}
	var report models.CommonReport
	if err := c.DB.Table("common_reports").Where("id = ?", id).Take(&report).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.NotFound()
			return
		}
		c.InternalError()
		return
	}
	if user.ID != report.UserID && !user.IsSysop() {
		c.JSONError(http.StatusForbidden, "Forbidden")
		return
	}

	if err := c.DB.Table("common_reports").Where("id = ?", id).Update("phase", models.CommonReportPhasePending).Error; err != nil {
		c.InternalError()
		return
	}
	if _, err := commonreportjob.Enqueue(context.Background(), c.AppContext, id); err != nil {
		slog.Error("common-report enqueue failed", "report_id", id, "err", err)
	}
	c.JSON(map[string]bool{"ok": true})
}

func Destroy(c *serverctx.Context) {
	user, ok := middleware.UserFromRequest(c.R)
	if !ok || user.ID < 1 {
		c.JSONError(http.StatusUnauthorized, "Unauthenticated")
		return
	}
	id, ok := c.PathInt("id")
	if !ok {
		c.NotFound()
		return
	}
	var report models.CommonReport
	if err := c.DB.Table("common_reports").Where("id = ?", id).Take(&report).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSONError(http.StatusNotFound, "해당 리포트가 없습니다.")
			return
		}
		c.InternalError()
		return
	}
	if user.ID != report.UserID && !user.IsSysop() {
		c.JSONError(http.StatusForbidden, "Forbidden")
		return
	}

	if err := c.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("common_report_items").Where("report_id = ?", id).Delete(nil).Error; err != nil {
			return err
		}
		return tx.Table("common_reports").Where("id = ?", id).Delete(nil).Error
	}); err != nil {
		c.InternalError()
		return
	}
	c.JSON(map[string]bool{"ok": true})
}

func createReport(db *gorm.DB, names []string, userID int, userName string) (models.CommonReport, error) {
	var report models.CommonReport
	err := db.Transaction(func(tx *gorm.DB) error {
		insert := app.H{
			"user_id":    userID,
			"user_name":  userName,
			"phase":      models.CommonReportPhasePending,
			"created_at": time.Now(),
			"updated_at": time.Now(),
		}
		if err := tx.Table("common_reports").Create(insert).Error; err != nil {
			return err
		}
		if err := tx.Table("common_reports").Order("id DESC").Take(&report).Error; err != nil {
			return err
		}
		rows := make([]app.H, 0, len(names))
		for _, name := range names {
			rows = append(rows, app.H{
				"report_id":     report.ID,
				"name":          name,
				"total":         0,
				"daum_blog":     0,
				"daum_book":     0,
				"naver_blog":    0,
				"naver_book":    0,
				"naver_news":    0,
				"google_search": 0,
				"created_at":    time.Now(),
				"updated_at":    time.Now(),
			})
		}
		return tx.Table("common_report_items").Create(&rows).Error
	})
	return report, err
}

func attachItems(db *gorm.DB, reports []models.CommonReport) error {
	if len(reports) == 0 {
		return nil
	}
	ids := make([]int, 0, len(reports))
	for _, r := range reports {
		ids = append(ids, r.ID)
	}
	items := make([]models.CommonReportItem, 0, len(reports)*4)
	if err := db.Table("common_report_items").Where("report_id IN ?", ids).Order("total DESC").Find(&items).Error; err != nil {
		return err
	}
	group := map[int][]models.CommonReportItem{}
	for _, it := range items {
		group[it.ReportID] = append(group[it.ReportID], it)
	}
	for i := range reports {
		reports[i].Items = group[reports[i].ID]
		reports[i].Total = sumTotal(reports[i].Items)
	}
	return nil
}

func loadReportWithItems(db *gorm.DB, id int) (models.CommonReport, error) {
	var report models.CommonReport
	if err := db.Table("common_reports").Where("id = ?", id).Take(&report).Error; err != nil {
		return models.CommonReport{}, err
	}
	items := make([]models.CommonReportItem, 0, 32)
	if err := db.Table("common_report_items").Where("report_id = ?", id).Order("total DESC").Find(&items).Error; err != nil {
		return models.CommonReport{}, err
	}
	report.Items = items
	report.Total = sumTotal(items)
	return report, nil
}

func sumTotal(items []models.CommonReportItem) int {
	total := 0
	for _, it := range items {
		total += it.Total
	}
	return total
}

func normalizeNames(names []string) []string {
	out := make([]string, 0, len(names))
	seen := map[string]struct{}{}
	for _, n := range names {
		s := strings.TrimSpace(n)
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}
