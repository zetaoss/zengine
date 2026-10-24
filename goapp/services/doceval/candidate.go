package doceval

import (
	"context"
	"math/rand/v2"

	"gorm.io/gorm"
)

type CandidatePage struct {
	PageID     int    `gorm:"column:page_id" json:"page_id"`
	PageTitle  string `gorm:"column:page_title" json:"page_title"`
	PageLatest int    `gorm:"column:page_latest" json:"page_latest"`
}

// PickCandidate selects an article candidate according to strategy: "random", "oldest", or "newest".
// In all three strategies, unreviewed articles are picked first (non-duplicate).
// If all articles have been evaluated, it falls back to the least-recently evaluated articles.
func PickCandidate(ctx context.Context, db *gorm.DB, strategy string) (*CandidatePage, error) {
	strategy = NormalizeStrategy(strategy)

	dbCtx := db
	if ctx != nil {
		dbCtx = db.WithContext(ctx)
	}

	// 1. Try unreviewed articles first (deduplicated against doc_evaluations).
	var c CandidatePage
	var err error

	switch strategy {
	case StrategyOldest:
		query := `
			SELECT p.page_id, p.page_title, p.page_latest
			FROM zetawiki.page p
			WHERE p.page_namespace = 0
			  AND p.page_is_redirect = 0
			  AND NOT EXISTS (
			    SELECT 1 FROM doc_evaluations e WHERE e.page_id = p.page_id
			  )
			ORDER BY p.page_id ASC
			LIMIT 1
		`
		err = dbCtx.Raw(query).Scan(&c).Error

	case StrategyNewest:
		query := `
			SELECT p.page_id, p.page_title, p.page_latest
			FROM zetawiki.page p
			WHERE p.page_namespace = 0
			  AND p.page_is_redirect = 0
			  AND NOT EXISTS (
			    SELECT 1 FROM doc_evaluations e WHERE e.page_id = p.page_id
			  )
			ORDER BY p.page_id DESC
			LIMIT 1
		`
		err = dbCtx.Raw(query).Scan(&c).Error

	default: // StrategyRandom
		// Utilize page_random index for fast random pick on unreviewed articles.
		// Generate random threshold once so MySQL doesn't re-evaluate RAND() per row.
		randThreshold := rand.Float64()
		queryWithRand := `
			SELECT p.page_id, p.page_title, p.page_latest
			FROM zetawiki.page p
			WHERE p.page_namespace = 0
			  AND p.page_is_redirect = 0
			  AND p.page_random >= ?
			  AND NOT EXISTS (
			    SELECT 1 FROM doc_evaluations e WHERE e.page_id = p.page_id
			  )
			ORDER BY p.page_random ASC
			LIMIT 1
		`
		err = dbCtx.Raw(queryWithRand, randThreshold).Scan(&c).Error
		if err == nil && c.PageID == 0 {
			// Wrap around if random threshold was near 1.0
			wrapQuery := `
				SELECT p.page_id, p.page_title, p.page_latest
				FROM zetawiki.page p
				WHERE p.page_namespace = 0
				  AND p.page_is_redirect = 0
				  AND NOT EXISTS (
				    SELECT 1 FROM doc_evaluations e WHERE e.page_id = p.page_id
				  )
				ORDER BY p.page_random ASC
				LIMIT 1
			`
			err = dbCtx.Raw(wrapQuery).Scan(&c).Error
		}
	}

	if err != nil {
		return nil, err
	}
	if c.PageID != 0 {
		return &c, nil
	}

	// 2. Fallback when ALL articles have already been evaluated at least once:
	// pick the article evaluated longest ago, ordered by the chosen strategy.
	fallbackOrder := "(e.last_eval_at IS NOT NULL), e.last_eval_at ASC, RAND()"
	switch strategy {
	case StrategyOldest:
		fallbackOrder = "e.last_eval_at ASC, p.page_id ASC"
	case StrategyNewest:
		fallbackOrder = "e.last_eval_at ASC, p.page_id DESC"
	}

	fallbackQuery := `
		SELECT p.page_id, p.page_title, p.page_latest
		FROM zetawiki.page p
		LEFT JOIN (
			SELECT page_id, MAX(created_at) AS last_eval_at
			FROM doc_evaluations
			GROUP BY page_id
		) e ON p.page_id = e.page_id
		WHERE p.page_namespace = 0 AND p.page_is_redirect = 0
		ORDER BY ` + fallbackOrder + `
		LIMIT 1
	`
	if err := dbCtx.Raw(fallbackQuery).Scan(&c).Error; err != nil {
		return nil, err
	}
	if c.PageID == 0 {
		return nil, nil
	}
	return &c, nil
}
