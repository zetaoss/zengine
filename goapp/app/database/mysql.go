package database

import (
	"fmt"
	"time"

	"github.com/zetaoss/zengine/goapp/app/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func Open(cfg *config.Config) (*gorm.DB, error) {
	dsn := buildMySQLDSN(cfg)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	// Close idle connections before the server's wait_timeout (300s) does, so the pool does not keep
	// connections the server already dropped and MariaDB logs no "Aborted connection" warnings.
	sqlDB.SetConnMaxIdleTime(3 * time.Minute)
	if err := sqlDB.Ping(); err != nil {
		_ = sqlDB.Close()
		return nil, err
	}
	return db, nil
}

// dbName is goapp's own database, fixed like the port; MediaWiki's tables are named with their database (zetawiki.page).
const dbName = "ldb"

func buildMySQLDSN(cfg *config.Config) string {
	user := cfg.DB.Username
	pass := cfg.DB.Password
	host := cfg.DB.Host

	addr := fmt.Sprintf("tcp(%s:3306)", host)
	return fmt.Sprintf(
		"%s:%s@%s/%s?parseTime=true&loc=UTC&charset=utf8mb4&collation=utf8mb4_unicode_ci",
		user,
		pass,
		addr,
		dbName,
	)
}
