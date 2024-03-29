package dbmigrations

import (
	"sort"

	"github.com/fullstaq-labs/sqedule/server/dbutils/gormigrate"
)

var dbMigrations []*gormigrate.Migration

// DbMigrations returns the list of migration objects, sorted by ID.
func DbMigrations() []*gormigrate.Migration {
	sort.Slice(dbMigrations, func(i, j int) bool {
		return dbMigrations[i].ID < dbMigrations[j].ID
	})
	return dbMigrations
}

func registerDbMigration(dbMigration *gormigrate.Migration) {
	dbMigrations = append(dbMigrations, dbMigration)
}
