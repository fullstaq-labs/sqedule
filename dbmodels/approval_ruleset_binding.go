package dbmodels

import (
	"github.com/fullstaq-labs/sqedule/dbmodels/approvalrulesetbindingmode"
	"gorm.io/gorm"
)

// ApprovalRulesetBinding ...
type ApprovalRulesetBinding struct {
	BaseModel

	ApplicationID string      `gorm:"type:citext; primaryKey; not null"`
	Application   Application `gorm:"foreignKey:OrganizationID,ApplicationID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`

	ApprovalRulesetID string          `gorm:"type:citext; primaryKey; not null"`
	ApprovalRuleset   ApprovalRuleset `gorm:"foreignKey:OrganizationID,ApprovalRulesetID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`

	Mode approvalrulesetbindingmode.Mode `gorm:"type:approval_ruleset_binding_mode; not null"`
}

// FindAllApprovalRulesetBindings ...
func FindAllApprovalRulesetBindings(db *gorm.DB, organizationID string, applicationID string) ([]ApprovalRulesetBinding, error) {
	var result []ApprovalRulesetBinding
	tx := db.Where("organization_id = ?", organizationID)
	if len(applicationID) > 0 {
		tx = tx.Where("application_id = ?", applicationID)
	}
	tx = tx.Find(&result)
	return result, tx.Error
}

// CollectApprovalRulesetBindingApplications ...
func CollectApprovalRulesetBindingApplications(bindings []ApprovalRulesetBinding) []*Application {
	result := make([]*Application, 0)
	for i := range bindings {
		binding := &bindings[i]
		result = append(result, &binding.Application)
	}
	return result
}

// CollectApprovalRulesetBindingRulesets ...
func CollectApprovalRulesetBindingRulesets(bindings []ApprovalRulesetBinding) []*ApprovalRuleset {
	result := make([]*ApprovalRuleset, 0)
	for i := range bindings {
		binding := &bindings[i]
		result = append(result, &binding.ApprovalRuleset)
	}
	return result
}
