package dbmigrations

import (
	"github.com/fullstaq-labs/sqedule/server/dbutils/gormigrate"
	"gorm.io/gorm"
)

func init() {
	registerDbMigration(&migration20210304000010)
}

var migration20210304000010 = gormigrate.Migration{
	ID: "20210304000010 Release approval ruleset binding",
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

		type ReleaseApprovalRulesetBinding struct {
			BaseModel

			ApplicationID string  `gorm:"type:citext; primaryKey; not null"`
			ReleaseID     uint64  `gorm:"primaryKey; not null"`
			Release       Release `gorm:"foreignKey:OrganizationID,ApplicationID,ReleaseID; references:OrganizationID,ApplicationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`

			ApprovalRulesetID string          `gorm:"type:citext; primaryKey; not null"`
			ApprovalRuleset   ApprovalRuleset `gorm:"foreignKey:OrganizationID,ApprovalRulesetID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`

			ApprovalRulesetMajorVersionID uint64                      `gorm:"not null"`
			ApprovalRulesetMajorVersion   ApprovalRulesetMajorVersion `gorm:"foreignKey:OrganizationID,ApprovalRulesetMajorVersionID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`

			ApprovalRulesetMinorVersionNumber uint32                      `gorm:"type:int; not null"`
			ApprovalRulesetMinorVersion       ApprovalRulesetMinorVersion `gorm:"foreignKey:OrganizationID,ApprovalRulesetMajorVersionID,ApprovalRulesetMinorVersionNumber; references:OrganizationID,ApprovalRulesetMajorVersionID,VersionNumber; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`

			Mode string `gorm:"type:approval_ruleset_binding_mode; not null"`
		}

		return tx.AutoMigrate(&ReleaseApprovalRulesetBinding{})
	},
	Rollback: func(tx *gorm.DB) error {
		return tx.Migrator().DropTable("release_approval_ruleset_bindings")
	},
}
