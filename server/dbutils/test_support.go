package dbutils

import (
	"context"
	"fmt"
	"sync"

	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var cachedDbconn *gorm.DB
var cachedDbconnMutex sync.Mutex

func SetupTestDatabase() (*gorm.DB, error) {
	db, err := getCachedOrNewDbconn()
	if err != nil {
		return nil, fmt.Errorf("Error establishing database connection: %w", err)
	}

	err = ClearDatabase(context.Background(), db)
	if err != nil {
		return nil, fmt.Errorf("Error clearing database: %w", err)
	}

	return db, nil
}

func getCachedOrNewDbconn() (*gorm.DB, error) {
	cachedDbconnMutex.Lock()
	defer cachedDbconnMutex.Unlock()

	var db *gorm.DB = cachedDbconn
	var err error

	if db == nil {
		logger := gormlogger.Default.LogMode(gormlogger.Warn)
		db, err = EstablishDatabaseConnection(
			// TODO: make database parameters configurable
			"postgresql",
			"dbname=sqedule_test",
			&gorm.Config{
				Logger: logger,
			})
		if err != nil {
			return nil, err
		}

		cachedDbconn = db
	}

	return db, nil
}
