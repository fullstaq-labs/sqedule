package dbutils

import (
	"context"
	"fmt"
	"sync"

	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type TestDatabaseContext struct {
	Db          *gorm.DB
	TableNames  []string
	ClearedOnce uint32
	SetupError  error
}

var (
	testDatabaseCtx   TestDatabaseContext
	testDatabaseSetup sync.Once
)

// SetupTestDatabase, when called for the first time, connects to the test database and clears it
// with `ClearDatabaseSlow`. On subsequent calls, it returns the initial database connection instead
// of making a new one, and clears the database with `ClearDatabaseFast`.
func SetupTestDatabase() (*gorm.DB, error) {
	testDatabaseSetup.Do(func() {
		if testDatabaseCtx.SetupError != nil {
			return
		}

		testDatabaseCtx.Db, testDatabaseCtx.SetupError = connectToTestDatabase()
		if testDatabaseCtx.SetupError != nil {
			testDatabaseCtx.SetupError = fmt.Errorf("Error establishing database connection: %w", testDatabaseCtx.SetupError)
			return
		}

		testDatabaseCtx.TableNames, testDatabaseCtx.SetupError = listTableNames(testDatabaseCtx.Db)
		if testDatabaseCtx.SetupError != nil {
			testDatabaseCtx.SetupError = fmt.Errorf("Error listing table names: %w", testDatabaseCtx.SetupError)
			return
		}
		testDatabaseCtx.Db.Logger.Info(context.Background(), "List of tables: %v", testDatabaseCtx.TableNames)
	})
	if testDatabaseCtx.SetupError != nil {
		return nil, testDatabaseCtx.SetupError
	}

	err := ClearDatabase(context.Background(), testDatabaseCtx.Db, testDatabaseCtx.TableNames)
	if err != nil {
		return nil, fmt.Errorf("Error clearing database: %w", err)
	}

	return testDatabaseCtx.Db, err
}

func connectToTestDatabase() (*gorm.DB, error) {
	var db *gorm.DB
	var err error

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

	return db, nil
}
