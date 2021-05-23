package dbmigrations

import (
	"database/sql"
	"time"

	"github.com/fullstaq-labs/sqedule/server/dbutils/gormigrate"
	"gorm.io/gorm"
)

func init() {
	registerDbMigration(&migration20201021000080)
}

var migration20201021000080 = gormigrate.Migration{
	ID: "20201021000080 Approval ruleset",
	Migrate: func(tx *gorm.DB) error {
		type Organization struct {
			ID string `gorm:"type:citext; primaryKey; not null"`
		}

		type BaseModel struct {
			OrganizationID string       `gorm:"type:citext; primaryKey; not null"`
			Organization   Organization `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
		}

		type ReviewableBase struct {
			CreatedAt time.Time `gorm:"not null"`
		}

		type ReviewableVersionBase struct {
			ID            uint64    `gorm:"primaryKey; autoIncrement; not null"`
			VersionNumber *uint32   `gorm:"type:int; check:(version_number > 0)"`
			CreatedAt     time.Time `gorm:"not null"`
			UpdatedAt     time.Time `gorm:"not null"`
		}

		type ReviewableAdjustmentBase struct {
			VersionNumber  uint32 `gorm:"type:int; primaryKey; not null; check:(version_number > 0)"`
			ReviewState    string `gorm:"type:review_state; not null"`
			ReviewComments sql.NullString
			CreatedAt      time.Time `gorm:"not null"`
		}

		type ApprovalRuleset struct {
			BaseModel
			ID string `gorm:"type:citext; primaryKey; not null"`
			ReviewableBase
		}

		type ApprovalRulesetMajorVersion struct {
			BaseModel
			ReviewableVersionBase
			ApprovalRulesetID string          `gorm:"type:citext; not null"`
			ApprovalRuleset   ApprovalRuleset `gorm:"foreignKey:OrganizationID,ApprovalRulesetID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
		}

		type ApprovalRulesetMinorVersion struct {
			BaseModel
			ApprovalRulesetMajorVersionID uint64 `gorm:"primaryKey; not null"`
			ReviewableAdjustmentBase
			Enabled bool `gorm:"not null; default:true"`

			DisplayName        string `gorm:"not null"`
			Description        string `gorm:"not null"`
			GloballyApplicable bool   `gorm:"not null; default:false"`

			ApprovalRulesetMajorVersion ApprovalRulesetMajorVersion `gorm:"foreignKey:OrganizationID,ApprovalRulesetMajorVersionID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
		}

		err := tx.AutoMigrate(&ApprovalRuleset{}, &ApprovalRulesetMajorVersion{},
			&ApprovalRulesetMinorVersion{})
		if err != nil {
			return err
		}

		err = tx.Exec("CREATE UNIQUE INDEX approval_ruleset_major_version_idx" +
			" ON approval_ruleset_major_versions (organization_id, approval_ruleset_id, version_number DESC)" +
			" WHERE (version_number IS NOT NULL)").Error
		if err != nil {
			return err
		}

		err = tx.Exec("CREATE INDEX approval_ruleset_minor_versions_globally_applicable_idx" +
			" ON approval_ruleset_minor_versions (organization_id, globally_applicable)" +
			" WHERE globally_applicable").Error
		if err != nil {
			return err
		}

		// Work around bug in Gorm: MinorVersion.VersionNumber shouldn't be autoincrement.
		err = tx.Exec("ALTER TABLE approval_ruleset_minor_versions ALTER COLUMN version_number DROP DEFAULT").Error
		if err != nil {
			return err
		}
		err = tx.Exec("DROP SEQUENCE approval_ruleset_minor_versions_version_number_seq").Error
		if err != nil {
			return err
		}

		return nil
	},
	Rollback: func(tx *gorm.DB) error {
		return tx.Migrator().DropTable("approval_ruleset_minor_versions",
			"approval_ruleset_major_versions", "approval_rulesets")
	},
}
