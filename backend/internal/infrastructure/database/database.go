package database

import (
	"fmt"

	"github.com/AbelHaro/url-shortener/backend/internal/config"
	"github.com/AbelHaro/url-shortener/backend/internal/domain"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Centralized entity registration for migrations
var entities = []interface{}{
	&domain.User{},
	&domain.RefreshToken{},
	&domain.URL{},
	&domain.URLStatistics{},
	&domain.IDsRange{},
}

func NewDB(cfg *config.AppConfig) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.AutoMigrate(entities...); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}

/*
NewDBFromDSN creates a new database connection using a DSN string. It is primarily intended for testing purposes, allowing tests to connect to a database instance with a dynamically generated DSN (e.g., from a test container). It performs the same auto-migration as NewDB to ensure the schema is up-to-date.
*/
func NewDBFromDSN(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.AutoMigrate(entities...); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}
