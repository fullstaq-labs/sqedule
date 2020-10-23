package dbmigrations

import (
	"github.com/fullstaq-labs/sqedule/dbutils/gormigrate"
	"gorm.io/gorm"
)

func init() {
	registerDbMigration(&migration20201021000040)
}

var migration20201021000040 = gormigrate.Migration{
	ID: "20201021000040 Review state",
	Migrate: func(tx *gorm.DB) error {
		return tx.Exec("CREATE TYPE review_state AS ENUM " +
			"('draft', 'reviewing', 'approved', 'rejected', 'abandoned')").Error
	},
	Rollback: func(tx *gorm.DB) error {
		return tx.Exec("DROP TYPE review_state").Error
	},
}
