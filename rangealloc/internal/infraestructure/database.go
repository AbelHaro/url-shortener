package infraestructure

import (
	"fmt"
	"strings"

	"github.com/AbelHaro/url-shortener/rangealloc/config"
	"github.com/AbelHaro/url-shortener/rangealloc/internal/domain"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewDB(cfg *config.DBConfig) (*gorm.DB, error) {
	// First, ensure the database exists
	if err := ensureDatabaseExists(cfg); err != nil {
		return nil, fmt.Errorf("failed to ensure database exists: %w", err)
	}

	// Now connect to the actual database
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	entities := []interface{}{
		&domain.Range{},
	}

	if err := db.AutoMigrate(entities...); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}

// ensureDatabaseExists creates the database if it doesn't exist
func ensureDatabaseExists(cfg *config.DBConfig) error {
	// Connect to the default postgres database to execute the CREATE DATABASE command

	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to postgres database: %w", err)
	}

	// Execute CREATE DATABASE
	createDBSQL := fmt.Sprintf("CREATE DATABASE \"%s\"", cfg.DBName)
	if err := db.Exec(createDBSQL).Error; err != nil {
		// If the database already exists, that's fine - just log and continue
		// PostgreSQL error code 42P04 indicates "duplicate_database"
		if strings.Contains(err.Error(), "42P04") {
			return nil
		}
		return fmt.Errorf("failed to create database: %w", err)
	}

	return nil
}
