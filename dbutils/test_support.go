package dbutils

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// SetupTestDatabase ...
func SetupTestDatabase() (*gorm.DB, error) {
	logger := gormlogger.Default.LogMode(gormlogger.Warn)
	db, err := EstablishDatabaseConnection(
		// TODO: make database parameters configurable
		"postgresql",
		"dbname=sqedule_test",
		&gorm.Config{
			Logger: logger,
		})
	if err != nil {
		return nil, fmt.Errorf("Error establishing database connection: %w", err)
	}

	err = ClearDatabase(context.Background(), db)
	if err != nil {
		return nil, fmt.Errorf("Error clearing database: %w", err)
	}

	return db, nil
}
