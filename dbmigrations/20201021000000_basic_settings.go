package dbmigrations

import (
	"github.com/fullstaq-labs/sqedule/dbutils/gormigrate"
	"gorm.io/gorm"
)

func init() {
	registerDbMigration(&migration20201021000000)
}

var migration20201021000000 = gormigrate.Migration{
	ID: "20201021000000 Basic settings",
	Migrate: func(tx *gorm.DB) error {
		return tx.Exec("CREATE EXTENSION IF NOT EXISTS citext").Error
	},
	Rollback: func(tx *gorm.DB) error {
		return tx.Exec("DROP EXTENSION citext").Error
	},
}
