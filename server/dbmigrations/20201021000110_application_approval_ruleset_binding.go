package dbmigrations

import (
	"database/sql"
	"time"

	"github.com/fullstaq-labs/sqedule/server/dbutils/gormigrate"
	"gorm.io/gorm"
)

func init() {
	registerDbMigration(&migration20201021000110)
}

var migration20201021000110 = gormigrate.Migration{
	ID: "20201021000110 Application approval ruleset binding",
	Migrate: func(tx *gorm.DB) error {
		type Organization struct {
			ID string `gorm:"type:citext; primaryKey; not null"`
		}

		type BaseModel struct {
			OrganizationID string       `gorm:"type:citext; primaryKey; not null"`
			Organization   Organization `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
		}

		type Application struct {
			BaseModel
			ID string `gorm:"type:citext; primaryKey; not null"`
		}

		type ApprovalRuleset struct {
			BaseModel
			ID string `gorm:"type:citext; primaryKey; not null"`
		}

		type ApplicationApprovalRulesetBindingPrimaryKey struct {
			ApplicationID     string `gorm:"type:citext; primaryKey; not null"`
			ApprovalRulesetID string `gorm:"type:citext; primaryKey; not null"`
		}

		type ApplicationApprovalRulesetBinding struct {
			BaseModel
			ApplicationApprovalRulesetBindingPrimaryKey
			Application     Application     `gorm:"foreignKey:OrganizationID,ApplicationID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
			ApprovalRuleset ApprovalRuleset `gorm:"foreignKey:OrganizationID,ApprovalRulesetID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
			CreatedAt       time.Time       `gorm:"not null"`
		}

		type ApplicationApprovalRulesetBindingMajorVersion struct {
			OrganizationID    string       `gorm:"type:citext; primaryKey; not null; index:app_approval_ruleset_binding_major_version_idx,unique"`
			Organization      Organization `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
			ID                uint64       `gorm:"primaryKey; autoIncrement; not null"`
			ApplicationID     string       `gorm:"type:citext; not null; index:app_approval_ruleset_binding_major_version_idx,unique"`
			ApprovalRulesetID string       `gorm:"type:citext; not null; index:app_approval_ruleset_binding_major_version_idx,unique"`
			VersionNumber     *uint32      `gorm:"type:int; index:app_approval_ruleset_binding_major_version_idx,sort:desc,where:version_number IS NOT NULL,unique; check:(version_number > 0)"`
			CreatedAt         time.Time    `gorm:"not null"`
			UpdatedAt         time.Time    `gorm:"not null"`

			ApplicationApprovalRulesetBinding ApplicationApprovalRulesetBinding `gorm:"foreignKey:OrganizationID,ApplicationID,ApprovalRulesetID; references:OrganizationID,ApplicationID,ApprovalRulesetID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
		}

		type ApplicationApprovalRulesetBindingMinorVersion struct {
			BaseModel
			ApplicationApprovalRulesetBindingMajorVersionID uint64 `gorm:"primaryKey; not null"`
			VersionNumber                                   uint32 `gorm:"type:int; primaryKey; not null; check:(version_number > 0)"`
			ReviewState                                     string `gorm:"type:review_state; not null"`
			ReviewComments                                  sql.NullString
			CreatedAt                                       time.Time `gorm:"not null"`
			Enabled                                         bool      `gorm:"not null; default:true"`

			Mode string `gorm:"type:approval_ruleset_binding_mode; not null"`

			ApplicationApprovalRulesetBindingMajorVersion ApplicationApprovalRulesetBindingMajorVersion `gorm:"foreignKey:OrganizationID,ApplicationApprovalRulesetBindingMajorVersionID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
		}

		err := tx.Exec("CREATE TYPE approval_ruleset_binding_mode AS ENUM " +
			"('permissive', 'enforcing')").Error
		if err != nil {
			return err
		}

		err = tx.AutoMigrate(&ApplicationApprovalRulesetBinding{})
		if err != nil {
			return err
		}

		err = tx.AutoMigrate(&ApplicationApprovalRulesetBindingMajorVersion{})
		if err != nil {
			return err
		}

		err = tx.AutoMigrate(&ApplicationApprovalRulesetBindingMinorVersion{})
		if err != nil {
			return err
		}

		return nil
	},
	Rollback: func(tx *gorm.DB) error {
		err := tx.Migrator().DropTable("application_approval_ruleset_binding_minor_versions",
			"application_approval_ruleset_binding_major_versions", "application_approval_ruleset_bindings")
		if err != nil {
			return err
		}

		return tx.Exec("DROP TYPE approval_ruleset_binding_mode").Error
	},
}
