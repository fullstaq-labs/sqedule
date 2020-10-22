package dbutils

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

func listTableNames(db *gorm.DB) ([]string, error) {
	return QueryStringList(db, "SELECT table_name FROM information_schema.tables"+
		" WHERE table_schema = current_schema() AND table_type = 'BASE TABLE'")
}

func listUserDefinedTypes(db *gorm.DB) ([]string, error) {
	return QueryStringList(db, "SELECT t.typname AS type"+
		" FROM pg_type t"+
		" LEFT JOIN pg_catalog.pg_namespace n ON n.oid = t.typnamespace"+
		" WHERE (t.typrelid = 0 OR (SELECT c.relkind = 'c' FROM pg_catalog.pg_class c WHERE c.oid = t.typrelid))"+
		" AND n.nspname = current_schema()"+
		" AND NOT EXISTS(SELECT 1 FROM pg_catalog.pg_type el WHERE el.oid = t.typelem AND el.typarray = t.oid)")
}

// ResetDatabase drops all tables and user-defined types in the current database.
func ResetDatabase(context context.Context, db *gorm.DB) error {
	if err := db.Begin().Error; err != nil {
		return err
	}

	tableNames, err := listTableNames(db)
	if err != nil {
		db.Rollback()
		return fmt.Errorf("error listing table names: %w", err)
	}
	db.Logger.Info(context, "List of tables: %v", tableNames)

	typeNames, err := listUserDefinedTypes(db)
	if err != nil {
		db.Rollback()
		return fmt.Errorf("error listing user-defined types: %w", err)
	}
	db.Logger.Info(context, "List of user-defined types: %v", typeNames)

	for _, tableName := range tableNames {
		if result := db.Exec(`DROP TABLE IF EXISTS "` + tableName + `" CASCADE`); result.Error != nil {
			db.Rollback()
			return fmt.Errorf("error dropping table %s: %w", tableName, result.Error)
		}
	}
	for _, typeName := range typeNames {
		if result := db.Exec(`DROP TYPE IF EXISTS "` + typeName + `" CASCADE`); result.Error != nil {
			db.Rollback()
			return fmt.Errorf("error dropping user-defined type %s: %w", typeName, result.Error)
		}
	}

	return db.Commit().Error
}
