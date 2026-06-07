// Package db provides a GORM database connection helper.
package db

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Open opens a GORM *DB using the provided DSN and verifies connectivity.
func Open(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("db.Open: %w", err)
	}

	sql, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("db.Open get sql.DB: %w", err)
	}
	if err := sql.Ping(); err != nil {
		return nil, fmt.Errorf("db.Open ping: %w", err)
	}
	return db, nil
}
