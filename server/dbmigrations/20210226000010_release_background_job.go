package dbmigrations

import (
	"time"

	"github.com/fullstaq-labs/sqedule/server/dbutils/gormigrate"
	"gorm.io/gorm"
)

func init() {
	registerDbMigration(&migration20210226000010)
}

var migration20210226000010 = gormigrate.Migration{
	ID: "20210226000010 Release background job",
	Migrate: func(tx *gorm.DB) error {
		type Organization struct {
			ID string `gorm:"type:citext; primaryKey; not null"`
		}

		type BaseModel struct {
			OrganizationID string       `gorm:"type:citext; primaryKey; not null"`
			Organization   Organization `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
		}

		type Release struct {
			BaseModel
			ApplicationID string `gorm:"type:citext; primaryKey; not null"`
			ID            uint64 `gorm:"primaryKey; not null"`
		}

		type ReleaseBackgroundJob struct {
			BaseModel
			ApplicationID string    `gorm:"type:citext; primaryKey; not null"`
			ReleaseID     uint64    `gorm:"primaryKey; not null"`
			Release       Release   `gorm:"foreignKey:OrganizationID,ApplicationID,ReleaseID; references:OrganizationID,ApplicationID,ID; constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
			LockSubID     uint32    `gorm:"type:int; autoIncrement; unique; not null; check:(lock_sub_id > 0)"`
			CreatedAt     time.Time `gorm:"not null"`
		}

		return tx.AutoMigrate(&ReleaseBackgroundJob{})
	},
	Rollback: func(tx *gorm.DB) error {
		return tx.Migrator().DropTable("release_background_jobs")
	},
}
