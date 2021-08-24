package dbmigrations

import (
	"github.com/fullstaq-labs/sqedule/server/dbutils/gormigrate"
	"gorm.io/gorm"
)

func init() {
	registerDbMigration(&migration20201021000040)
}

var migration20201021000040 = gormigrate.Migration{
	ID: "20201021000040 Proposal state",
	Migrate: func(tx *gorm.DB) error {
		return tx.Exec("CREATE TYPE proposal_state AS ENUM " +
			"('draft', 'reviewing', 'approved', 'rejected', 'abandoned')").Error
	},
	Rollback: func(tx *gorm.DB) error {
		return tx.Exec("DROP TYPE proposal_state").Error
	},
}
