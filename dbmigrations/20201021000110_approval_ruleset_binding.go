package dbmigrations

import (
	"github.com/fullstaq-labs/sqedule/dbmodels/approvalrulesetbindingmode"
	"github.com/fullstaq-labs/sqedule/dbutils/gormigrate"
	"gorm.io/gorm"
)

func init() {
	registerDbMigration(&migration20201021000110)
}

var migration20201021000110 = gormigrate.Migration{
	ID: "20201021000110 Approval ruleset binding",
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

		// ApprovalRulesetBinding ...
		type ApprovalRulesetBinding struct {
			BaseModel

			ApplicationID string      `gorm:"type:citext; primaryKey; not null"`
			Application   Application `gorm:"foreignKey:OrganizationID,ApplicationID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`

			ApprovalRulesetID string          `gorm:"type:citext; primaryKey; not null"`
			ApprovalRuleset   ApprovalRuleset `gorm:"foreignKey:OrganizationID,ApprovalRulesetID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`

			Mode approvalrulesetbindingmode.Mode `gorm:"type:approval_ruleset_binding_mode; not null"`
		}

		err := tx.Exec("CREATE TYPE approval_ruleset_binding_mode AS ENUM " +
			"('permissive', 'enforcing')").Error
		if err != nil {
			return err
		}

		return tx.AutoMigrate(&ApprovalRulesetBinding{})
	},
	Rollback: func(tx *gorm.DB) error {
		err := tx.Migrator().DropTable("approval_ruleset_bindings")
		if err != nil {
			return err
		}

		return tx.Exec("DROP TYPE approval_ruleset_binding_mode").Error
	},
}
