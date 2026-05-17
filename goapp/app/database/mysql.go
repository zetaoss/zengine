package database

import (
	"fmt"

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
	if err := sqlDB.Ping(); err != nil {
		_ = sqlDB.Close()
		return nil, err
	}
	return db, nil
}

func buildMySQLDSN(cfg *config.Config) string {
	user := cfg.DB.Username
	pass := cfg.DB.Password
	dbName := cfg.DB.Database
	host := cfg.DB.Host
	port := cfg.DB.Port
	if port == 0 {
		port = 3306
	}

	addr := fmt.Sprintf("tcp(%s:%d)", host, port)
	return fmt.Sprintf(
		"%s:%s@%s/%s?parseTime=true&loc=UTC&charset=utf8mb4&collation=utf8mb4_unicode_ci",
		user,
		pass,
		addr,
		dbName,
	)
}
