package dbmigrations

import (
	"github.com/fullstaq-labs/sqedule/dbutils/gormigrate"
	"gorm.io/gorm"
)

var migration20201021000010 = gormigrate.Migration{
	ID: "20201021000010 Organization member roles",
	Migrate: func(tx *gorm.DB) error {
		return tx.Exec("CREATE TYPE organization_member_role AS ENUM ('owner', 'admin', 'change_manager', 'technician', 'viewer')").Error
	},
	Rollback: func(tx *gorm.DB) error {
		return tx.Exec("DROP TYPE organization_member_roles").Error
	},
}

func init() {
	registerDbMigration(&migration20201021000010)
}
