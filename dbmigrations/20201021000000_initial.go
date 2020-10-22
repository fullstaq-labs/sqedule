package dbmigrations

import (
	"errors"

	"github.com/fullstaq-labs/sqedule/dbutils/gormigrate"
	"gorm.io/gorm"
)

var migration20201021000000 = gormigrate.Migration{
	ID: "20201021000000",
	Migrate: func(tx *gorm.DB) error {
		return errors.New("oh no")
	},
	Rollback: func(tx *gorm.DB) error {
		return nil
	},
}

func init() {
	registerDbMigration(&migration20201021000000)
}
