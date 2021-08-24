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

		type ReviewableBase struct {
			CreatedAt time.Time `gorm:"not null"`
			UpdatedAt time.Time `gorm:"not null"`
		}

		type ReviewableVersionBase struct {
			ID            uint64       `gorm:"primaryKey; autoIncrement; not null"`
			VersionNumber *uint32      `gorm:"type:int; check:(version_number > 0)"`
			CreatedAt     time.Time    `gorm:"not null"`
			ApprovedAt    sql.NullTime `gorm:"check:((approved_at IS NULL) = (version_number IS NULL))"`
		}

		type ReviewableAdjustmentBase struct {
			AdjustmentNumber uint32 `gorm:"type:int; primaryKey; not null; check:(adjustment_number > 0)"`
			ProposalState    string `gorm:"type:proposal_state; not null"`
			ReviewComments   sql.NullString
			CreatedAt        time.Time `gorm:"not null"`
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
			ReviewableBase
			Application     Application     `gorm:"foreignKey:OrganizationID,ApplicationID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
			ApprovalRuleset ApprovalRuleset `gorm:"foreignKey:OrganizationID,ApprovalRulesetID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
		}

		type ApplicationApprovalRulesetBindingVersion struct {
			BaseModel
			ApplicationID     string `gorm:"type:citext; not null"`
			ApprovalRulesetID string `gorm:"type:citext; not null"`
			ReviewableVersionBase

			ApplicationApprovalRulesetBinding ApplicationApprovalRulesetBinding `gorm:"foreignKey:OrganizationID,ApplicationID,ApprovalRulesetID; references:OrganizationID,ApplicationID,ApprovalRulesetID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
		}

		type ApplicationApprovalRulesetBindingAdjustment struct {
			BaseModel
			ApplicationApprovalRulesetBindingVersionID uint64 `gorm:"primaryKey; not null"`
			ReviewableAdjustmentBase
			Enabled *bool `gorm:"not null; default:true"`

			Mode string `gorm:"type:approval_ruleset_binding_mode; not null"`

			ApplicationApprovalRulesetBindingVersion ApplicationApprovalRulesetBindingVersion `gorm:"foreignKey:OrganizationID,ApplicationApprovalRulesetBindingVersionID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
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
		err = tx.AutoMigrate(&ApplicationApprovalRulesetBindingVersion{})
		if err != nil {
			return err
		}
		err = tx.AutoMigrate(&ApplicationApprovalRulesetBindingAdjustment{})
		if err != nil {
			return err
		}

		err = tx.Exec("CREATE UNIQUE INDEX app_approval_ruleset_binding_version_idx" +
			" ON application_approval_ruleset_binding_versions (organization_id, application_id, approval_ruleset_id, version_number DESC)" +
			" WHERE (version_number IS NOT NULL)").Error
		if err != nil {
			return err
		}

		// Work around bug in Gorm: Adjustment.VersionNumber shouldn't be autoincrement.
		err = tx.Exec("ALTER TABLE application_approval_ruleset_binding_adjustments ALTER COLUMN adjustment_number DROP DEFAULT").Error
		if err != nil {
			return err
		}
		err = tx.Exec("DROP SEQUENCE application_approval_ruleset_binding_adju_adjustment_number_seq").Error
		if err != nil {
			return err
		}

		return nil
	},
	Rollback: func(tx *gorm.DB) error {
		err := tx.Migrator().DropTable("application_approval_ruleset_binding_adjustments",
			"application_approval_ruleset_binding_versions", "application_approval_ruleset_bindings")
		if err != nil {
			return err
		}

		return tx.Exec("DROP TYPE approval_ruleset_binding_mode").Error
	},
}
