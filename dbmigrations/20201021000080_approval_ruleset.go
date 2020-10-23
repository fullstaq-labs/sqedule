package dbmigrations

import (
	"database/sql"
	"time"

	"github.com/fullstaq-labs/sqedule/dbutils/gormigrate"
	"gorm.io/gorm"
)

func init() {
	registerDbMigration(&migration20201021000080)
}

var migration20201021000080 = gormigrate.Migration{
	ID: "20201021000080 Approval ruleset",
	Migrate: func(tx *gorm.DB) error {
		type Organization struct {
			ID string `gorm:"type: citext; primaryKey; not null"`
		}

		type BaseModel struct {
			OrganizationID string       `gorm:"type: citext; primaryKey; not null"`
			Organization   Organization `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
		}

		type ApprovalRuleset struct {
			BaseModel
			ID        string    `gorm:"type: citext; primaryKey; not null"`
			CreatedAt time.Time `gorm:"not null"`
		}

		type ApprovalRulesetMajorVersion struct {
			OrganizationID    string       `gorm:"type: citext; primaryKey; not null; index:approval_ruleset_major_version_idx,unique"`
			Organization      Organization `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
			ID                uint64       `gorm:"primaryKey; autoIncrement; not null"`
			ApprovalRulesetID string       `gorm:"type: citext; index:approval_ruleset_major_version_idx,unique"`
			VersionNumber     *uint32      `gorm:"index:approval_ruleset_major_version_idx,unique"`
			CreatedAt         time.Time    `gorm:"not null"`
			UpdatedAt         time.Time    `gorm:"not null"`

			ApprovalRuleset ApprovalRuleset `gorm:"foreignKey:OrganizationID,ApprovalRulesetID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
		}

		type ApprovalRulesetMinorVersion struct {
			BaseModel
			ApprovalRulesetMajorVersionID uint64 `gorm:"primaryKey; not null"`
			VersionNumber                 uint32 `gorm:"primaryKey; not null"`
			ReviewState                   string `gorm:"type:review_state; not null"`
			ReviewComments                sql.NullString
			CreatedAt                     time.Time `gorm:"not null"`
			Enabled                       bool      `gorm:"not null; default:true"`

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

		return tx.Exec("CREATE INDEX approval_ruleset_minor_versions_globally_applicable_idx" +
			" ON approval_ruleset_minor_versions (organization_id, globally_applicable)" +
			" WHERE globally_applicable").Error
	},
	Rollback: func(tx *gorm.DB) error {
		return tx.Migrator().DropTable("approval_ruleset_minor_versions",
			"approval_ruleset_major_versions", "approval_rulesets")
	},
}
