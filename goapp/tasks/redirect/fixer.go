package redirect

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/app/taskctx"
)

type FixerTask struct{}

func NewFixerTask() *FixerTask { return &FixerTask{} }

type doubleRedirectTarget struct {
	FromNamespace int    `gorm:"column:from_namespace"`
	FromTitle     string `gorm:"column:from_title"`
	ToNamespace   int    `gorm:"column:to_namespace"`
	ToTitle       string `gorm:"column:to_title"`
}

func (t *FixerTask) Execute(ctx context.Context, taskCtx taskctx.Context, _ any) (app.H, error) {
	db, err := taskCtx.GetDB()
	if err != nil {
		return nil, err
	}

	var targets []doubleRedirectTarget
	err = db.WithContext(ctx).Raw(`
		SELECT pa.page_namespace AS from_namespace, pa.page_title AS from_title,
		       pb.page_namespace AS to_namespace, pb.page_title AS to_title
		FROM redirect
		JOIN page pa ON rd_from = pa.page_id
		JOIN page pb ON rd_namespace = pb.page_namespace AND rd_title = pb.page_title
		WHERE rd_interwiki = '' AND pb.page_is_redirect = 1
	`).Scan(&targets).Error
	if err != nil {
		return nil, fmt.Errorf("query double redirects: %w", err)
	}

	if len(targets) == 0 {
		return app.H{"found": 0, "fixed": 0, "status": "no double redirects"}, nil
	}

	slog.Info("[redirect-fixer] found double redirects to fix", "count", len(targets))

	mwPath := os.Getenv("MW_INSTALL_PATH")
	if mwPath == "" {
		mwPath = "/app/w"
	}
	scriptPath := filepath.Join(mwPath, "maintenance", "run.php")
	if _, err := os.Stat(scriptPath); err != nil {
		return nil, fmt.Errorf("maintenance script not found at %s: %w", scriptPath, err)
	}

	execCtx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(execCtx, "php", scriptPath, "fixDoubleRedirects")
	out, err := cmd.CombinedOutput()
	outputStr := strings.TrimSpace(string(out))
	if err != nil {
		slog.Error("[redirect-fixer] fixDoubleRedirects script failed", "err", err, "output", outputStr)
		return nil, fmt.Errorf("fixDoubleRedirects execution failed: %w (output: %s)", err, outputStr)
	}

	slog.Info("[redirect-fixer] fixDoubleRedirects completed", "count", len(targets), "output", outputStr)
	return app.H{
		"found":  len(targets),
		"fixed":  len(targets),
		"output": outputStr,
	}, nil
}
