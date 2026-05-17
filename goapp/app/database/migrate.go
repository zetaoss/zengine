package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/zetaoss/zengine/goapp/app/config"
)

type migrationFile struct {
	Version string
	Name    string
	Path    string
}

func RunMigrate(cfg *config.Config) error {
	fmt.Fprintln(os.Stderr, "[worker:migrate] start")
	db, err := Open(cfg)
	if err != nil {
		return err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	defer func() {
		_ = sqlDB.Close()
	}()

	if err := ensureSchemaMigrations(sqlDB); err != nil {
		return err
	}

	files, err := findUpMigrations()
	if err != nil {
		return err
	}
	if len(files) == 0 {
		fmt.Println("no migrations found")
		return nil
	}

	applied := 0
	for _, mf := range files {
		ok, err := isApplied(sqlDB, mf.Version)
		if err != nil {
			return err
		}
		if ok {
			continue
		}
		raw, err := os.ReadFile(mf.Path)
		if err != nil {
			return err
		}
		tx, err := sqlDB.Begin()
		if err != nil {
			return err
		}
		for _, stmt := range splitSQLStatements(string(raw)) {
			if _, err := tx.Exec(stmt); err != nil {
				_ = tx.Rollback()
				return fmt.Errorf("apply migration %s: %w", mf.Name, err)
			}
		}
		if _, err := tx.Exec(`INSERT INTO schema_migrations(version, name, applied_at) VALUES(?, ?, UTC_TIMESTAMP())`, mf.Version, mf.Name); err != nil {
			_ = tx.Rollback()
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
		applied++
		fmt.Printf("applied migration: %s\n", mf.Name)
	}
	fmt.Printf("migrate done: applied=%d\n", applied)
	return nil
}

func ensureSchemaMigrations(db *sql.DB) error {
	_, err := db.Exec(`
CREATE TABLE IF NOT EXISTS schema_migrations (
  version VARCHAR(32) PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
)`)
	return err
}

func isApplied(db *sql.DB, version string) (bool, error) {
	var n int
	err := db.QueryRow(`SELECT COUNT(1) FROM schema_migrations WHERE version = ?`, version).Scan(&n)
	return n > 0, err
}

func findUpMigrations() ([]migrationFile, error) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		return nil, fmt.Errorf("unable to resolve migrate.go path")
	}
	base := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "migrations"))
	entries, err := os.ReadDir(base)
	if err != nil {
		return nil, err
	}
	out := make([]migrationFile, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".up.sql") {
			continue
		}
		parts := strings.SplitN(name, "_", 2)
		if len(parts) < 2 {
			continue
		}
		out = append(out, migrationFile{Version: parts[0], Name: name, Path: filepath.Join(base, name)})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Version == out[j].Version {
			return out[i].Name < out[j].Name
		}
		return out[i].Version < out[j].Version
	})
	return out, nil
}

func splitSQLStatements(raw string) []string {
	parts := strings.Split(raw, ";")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		s := strings.TrimSpace(p)
		if s == "" {
			continue
		}
		out = append(out, s)
	}
	return out
}
