package dbmigrations

import (
	"github.com/fullstaq-labs/sqedule/dbutils/gormigrate"
	"gorm.io/gorm"
)

func init() {
	registerDbMigration(&migration20201021000005)
}

var migration20201021000005 = gormigrate.Migration{
	ID: "20201021000005 Organization",
	Migrate: func(tx *gorm.DB) error {
		type Organization struct {
			ID          string `gorm:"type:citext; primaryKey; not null"`
			DisplayName string `gorm:"not null"`
		}

		return tx.AutoMigrate(&Organization{})
	},
	Rollback: func(tx *gorm.DB) error {
		return tx.Migrator().DropTable("organizations")
	},
}
