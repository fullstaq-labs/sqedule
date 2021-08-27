package dbutils

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

func listTableNames(db *gorm.DB) ([]string, error) {
	return QueryStringList(db, "SELECT table_name FROM information_schema.tables"+
		" WHERE table_schema = current_schema() AND table_type = 'BASE TABLE'"+
		" ORDER BY table_name")
}

func listUserDefinedTypes(db *gorm.DB) ([]string, error) {
	return QueryStringList(db, "SELECT t.typname AS type"+
		" FROM pg_type t"+
		" LEFT JOIN pg_catalog.pg_namespace n ON n.oid = t.typnamespace"+
		" WHERE (t.typrelid = 0 OR (SELECT c.relkind = 'c' FROM pg_catalog.pg_class c WHERE c.oid = t.typrelid))"+
		" AND n.nspname = current_schema()"+
		" AND NOT EXISTS(SELECT 1 FROM pg_catalog.pg_type el WHERE el.oid = t.typelem AND el.typarray = t.oid)"+
		" AND t.typname != 'citext'"+
		" ORDER BY type")
}

func listExtensions(db *gorm.DB) ([]string, error) {
	return QueryStringList(db, "SELECT extname FROM pg_extension WHERE extname != 'plpgsql' ORDER BY extname")
}

// ClearDatabaseSlow clears all tables and sequences in the current database, using TRUNCATE.
// This is usually slower than ClearDatabaseFast and is not concurrency-safe, but it reclaims
// disk space and resets sequence numbers.
func ClearDatabaseSlow(context context.Context, db *gorm.DB, tableNames []string) error {
	return db.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", strings.Join(tableNames, ", "))).Error
}

// ClearDatabaseFast clears all tables and sequences in the current database, using DELETE.
// This is usually faster than ClearDatabaseSlow and is concurrency-safe, but does not reclaim
// disk space and does not reset sequence numbers.
func ClearDatabaseFast(context context.Context, db *gorm.DB, tableNames []string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// Should delete nearly all data because of 'ON DELETE CASCADE'
		err := tx.Exec("DELETE FROM organizations").Error
		if err != nil {
			return err
		}

		for _, tableName := range tableNames {
			err = tx.Exec(fmt.Sprintf("DELETE FROM %s", tableName)).Error
			if err != nil {
				return err
			}
		}

		return nil
	})
}

// ResetDatabase drops all tables and user-defined types in the current database.
func ResetDatabase(context context.Context, db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		tableNames, err := listTableNames(tx)
		if err != nil {
			return fmt.Errorf("error listing table names: %w", err)
		}
		db.Logger.Info(context, "List of tables: %v", tableNames)

		typeNames, err := listUserDefinedTypes(tx)
		if err != nil {
			return fmt.Errorf("error listing user-defined types: %w", err)
		}
		db.Logger.Info(context, "List of user-defined types: %v", typeNames)

		extensionNames, err := listExtensions(tx)
		if err != nil {
			return fmt.Errorf("error listing extensions: %w", err)
		}
		db.Logger.Info(context, "List of extensions: %v", typeNames)

		for _, tableName := range tableNames {
			if result := tx.Exec(`DROP TABLE IF EXISTS "` + tableName + `" CASCADE`); result.Error != nil {
				return fmt.Errorf("error dropping table %s: %w", tableName, result.Error)
			}
		}
		for _, typeName := range typeNames {
			if result := tx.Exec(`DROP TYPE IF EXISTS "` + typeName + `" CASCADE`); result.Error != nil {
				return fmt.Errorf("error dropping user-defined type %s: %w", typeName, result.Error)
			}
		}
		for _, extensionName := range extensionNames {
			if result := tx.Exec(`DROP EXTENSION IF EXISTS "` + extensionName + `" CASCADE`); result.Error != nil {
				return fmt.Errorf("error dropping extension %s: %w", extensionName, result.Error)
			}
		}

		return nil
	})
}
