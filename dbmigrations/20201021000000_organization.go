package dbmigrations

import (
	"github.com/fullstaq-labs/sqedule/dbutils/gormigrate"
	"gorm.io/gorm"
)

var migration20201021000000 = gormigrate.Migration{
	ID: "20201021000000 Organization",
	Migrate: func(tx *gorm.DB) error {
		type Organization struct {
			ID          string `gorm:"primaryKey; not null"`
			DisplayName string `gorm:"not null"`
		}

		return tx.AutoMigrate(&Organization{})
	},
	Rollback: func(tx *gorm.DB) error {
		return tx.Migrator().DropTable("organizations")
	},
}

func init() {
	registerDbMigration(&migration20201021000000)
}
