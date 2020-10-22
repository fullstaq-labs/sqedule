package dbmigrations

import (
	"github.com/fullstaq-labs/sqedule/dbutils/gormigrate"
)

// DbMigrations contains all migration objects.
var DbMigrations []*gormigrate.Migration

func registerDbMigration(dbMigration *gormigrate.Migration) {
	DbMigrations = append(DbMigrations, dbMigration)
}
