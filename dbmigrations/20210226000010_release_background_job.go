package dbmigrations

import (
	"time"

	"github.com/fullstaq-labs/sqedule/dbutils/gormigrate"
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

		type ApprovalRuleset struct {
			BaseModel
			ID string `gorm:"type:citext; primaryKey; not null"`
		}

		type ApprovalRulesetMajorVersion struct {
			OrganizationID string       `gorm:"type:citext; primaryKey; not null; index:approval_ruleset_major_version_idx,unique"`
			Organization   Organization `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
			ID             uint64       `gorm:"primaryKey; autoIncrement; not null"`
		}

		type ApprovalRulesetMinorVersion struct {
			BaseModel
			ApprovalRulesetMajorVersionID uint64 `gorm:"primaryKey; not null"`
			VersionNumber                 uint32 `gorm:"type:int; primaryKey; not null; check:(version_number > 0)"`
		}

		type ReleaseBackgroundJob struct {
			BaseModel
			ApplicationID string    `gorm:"type:citext; primaryKey; not null"`
			ReleaseID     uint64    `gorm:"primaryKey; not null"`
			Release       Release   `gorm:"foreignKey:OrganizationID,ApplicationID,ReleaseID; references:OrganizationID,ApplicationID,ID; constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
			LockID        uint32    `gorm:"type:int; autoIncrement; unique; not null; check:(lock_id > 0)"`
			CreatedAt     time.Time `gorm:"not null"`
		}

		return tx.AutoMigrate(&ReleaseBackgroundJob{})
	},
	Rollback: func(tx *gorm.DB) error {
		return tx.Migrator().DropTable("release_background_jobs")
	},
}
