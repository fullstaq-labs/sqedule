package dbmodels

import "github.com/fullstaq-labs/sqedule/dbmodels/approvalrulesetbindingmode"

// ApprovalRulesetBinding ...
type ApprovalRulesetBinding struct {
	BaseModel

	ApplicationID string      `gorm:"type:citext; primaryKey; not null"`
	Application   Application `gorm:"foreignKey:OrganizationID,ApplicationID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`

	ApprovalRulesetID string          `gorm:"type:citext; primaryKey; not null"`
	ApprovalRuleset   ApprovalRuleset `gorm:"foreignKey:OrganizationID,ApprovalRulesetID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`

	Mode approvalrulesetbindingmode.Mode `gorm:"type:approval_ruleset_binding_mode; not null"`
}
