package doceval

import (
	"context"
	"database/sql"
	"math/rand/v2"

	"github.com/zetaoss/zengine/goapp/prompts"
	"gorm.io/gorm"
)

type CandidatePage struct {
	PageID     int    `gorm:"column:page_id" json:"page_id"`
	PageTitle  string `gorm:"column:page_title" json:"page_title"`
	PageLatest int    `gorm:"column:page_latest" json:"page_latest"`
}

// majorSQL is the major part of a chain_version column, as prompts.Major gives it.
func majorSQL(col string) string {
	return "SUBSTRING_INDEX(" + col + ", '.', 1)"
}

// currentDefinitionSQL matches a definition d written for page p's latest revision and title by the current
// doc-define chain's major version, bound as @defineMajor (defineMajorArg).
var currentDefinitionSQL = `d.page_id = p.page_id AND d.revid = p.page_latest AND d.page_title = p.page_title AND ` + majorSQL("d.chain_version") + ` = @defineMajor`

// defineMajorArg binds @defineMajor in currentDefinitionSQL.
func defineMajorArg() sql.NamedArg {
	return sql.Named("defineMajor", prompts.Major(prompts.DocDefine.Metadata.Version))
}

// PickCandidate selects an article to evaluate by strategy: one with a current definition and no evaluation of
// its latest revision by the current doc-eval major version. An edit or a new major version makes a page a
// candidate again. It returns nil when every article is evaluated.
func PickCandidate(ctx context.Context, db *gorm.DB, strategy string) (*CandidatePage, error) {
	return pickPage(ctx, db, strategy, `
		EXISTS (SELECT 1 FROM doc_definitions d WHERE `+currentDefinitionSQL+`)
		AND NOT EXISTS (
		  SELECT 1 FROM doc_evaluations e
		  WHERE e.page_id = p.page_id AND e.revid = p.page_latest AND `+majorSQL("e.chain_version")+` = @evalMajor
		)`,
		defineMajorArg(),
		sql.Named("evalMajor", prompts.Major(prompts.DocEval.Metadata.Version)))
}

// PickUndefined selects an article to define by strategy: one without a current definition (none, or one of
// another revision, title or major version). It returns nil when every article is defined.
func PickUndefined(ctx context.Context, db *gorm.DB, strategy string) (*CandidatePage, error) {
	return pickPage(ctx, db, strategy, `
		NOT EXISTS (SELECT 1 FROM doc_definitions d WHERE `+currentDefinitionSQL+`)`,
		defineMajorArg())
}

// pickPage returns one non-redirect article matching cond, ordered by strategy: random (page_random from a
// random threshold, wrapping around), oldest (smallest page_id), or newest (largest page_id).
func pickPage(ctx context.Context, db *gorm.DB, strategy, cond string, args ...any) (*CandidatePage, error) {
	query := func(extra, order string) string {
		return `
			SELECT p.page_id, p.page_title, p.page_latest
			FROM zetawiki.page p
			WHERE p.page_namespace = 0
			  AND p.page_is_redirect = 0
			  AND ` + cond + extra + `
			ORDER BY ` + order + `
			LIMIT 1
		`
	}
	dbCtx := db.WithContext(ctx)
	var c CandidatePage
	var err error
	switch NormalizeStrategy(strategy) {
	case StrategyOldest:
		err = dbCtx.Raw(query("", "p.page_id ASC"), args...).Scan(&c).Error
	case StrategyNewest:
		err = dbCtx.Raw(query("", "p.page_id DESC"), args...).Scan(&c).Error
	default:
		// One threshold per pick, so MySQL does not evaluate RAND() per row and the page_random index is used.
		threshold := append(append([]any{}, args...), sql.Named("threshold", rand.Float64()))
		err = dbCtx.Raw(query(" AND p.page_random >= @threshold", "p.page_random ASC"), threshold...).Scan(&c).Error
		if err == nil && c.PageID == 0 {
			err = dbCtx.Raw(query("", "p.page_random ASC"), args...).Scan(&c).Error
		}
	}
	if err != nil {
		return nil, err
	}
	if c.PageID == 0 {
		return nil, nil
	}
	return &c, nil
}
