package dbmodels

import (
	"reflect"

	"github.com/fullstaq-labs/sqedule/server/dbmodels/approvalrulesetbindingmode"
	"gorm.io/gorm"
)

// ApplicationApprovalRulesetBindingPrimaryKey ...
type ApplicationApprovalRulesetBindingPrimaryKey struct {
	ApplicationID     string `gorm:"type:citext; primaryKey; not null"`
	ApprovalRulesetID string `gorm:"type:citext; primaryKey; not null"`
}

// ApplicationApprovalRulesetBinding ...
type ApplicationApprovalRulesetBinding struct {
	BaseModel
	ApplicationApprovalRulesetBindingPrimaryKey
	ReviewableBase
	Application        Application                                    `gorm:"foreignKey:OrganizationID,ApplicationID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	ApprovalRuleset    ApprovalRuleset                                `gorm:"foreignKey:OrganizationID,ApprovalRulesetID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	LatestMajorVersion *ApplicationApprovalRulesetBindingMajorVersion `gorm:"-"`
	LatestMinorVersion *ApplicationApprovalRulesetBindingMinorVersion `gorm:"-"`
}

// ApplicationApprovalRulesetBindingMajorVersion ...
type ApplicationApprovalRulesetBindingMajorVersion struct {
	BaseModel
	ApplicationID     string `gorm:"type:citext; not null; index:app_approval_ruleset_binding_major_version_idx,unique"`
	ApprovalRulesetID string `gorm:"type:citext; not null; index:app_approval_ruleset_binding_major_version_idx,unique"`
	ReviewableVersionBase

	ApplicationApprovalRulesetBinding ApplicationApprovalRulesetBinding `gorm:"foreignKey:OrganizationID,ApplicationID,ApprovalRulesetID; references:OrganizationID,ApplicationID,ApprovalRulesetID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
}

// ApplicationApprovalRulesetBindingMinorVersion ...
type ApplicationApprovalRulesetBindingMinorVersion struct {
	BaseModel
	ApplicationApprovalRulesetBindingMajorVersionID uint64 `gorm:"primaryKey; not null"`
	ReviewableAdjustmentBase
	Enabled bool `gorm:"not null; default:true"`

	Mode approvalrulesetbindingmode.Mode `gorm:"type:approval_ruleset_binding_mode; not null"`

	ApplicationApprovalRulesetBindingMajorVersion ApplicationApprovalRulesetBindingMajorVersion `gorm:"foreignKey:OrganizationID,ApplicationApprovalRulesetBindingMajorVersionID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
}

// FindAllApplicationApprovalRulesetBindings ...
func FindAllApplicationApprovalRulesetBindings(db *gorm.DB, organizationID string, applicationID string) ([]ApplicationApprovalRulesetBinding, error) {
	var result []ApplicationApprovalRulesetBinding
	tx := db.Where("organization_id = ?", organizationID)
	if len(applicationID) > 0 {
		tx = tx.Where("application_id = ?", applicationID)
	}
	tx = tx.Find(&result)
	return result, tx.Error
}

func FindAllApplicationApprovalRulesetBindingsWithApprovalRuleset(db *gorm.DB, organizationID string, rulesetID string) ([]ApplicationApprovalRulesetBinding, error) {
	var result []ApplicationApprovalRulesetBinding
	tx := db.Where("organization_id = ? AND approval_ruleset_id = ?", organizationID, rulesetID)
	tx = tx.Find(&result)
	return result, tx.Error
}

// MakeApplicationApprovalRulesetBindingsPointerArray ...
func MakeApplicationApprovalRulesetBindingsPointerArray(bindings []ApplicationApprovalRulesetBinding) []*ApplicationApprovalRulesetBinding {
	result := make([]*ApplicationApprovalRulesetBinding, 0, len(bindings))
	for i := range bindings {
		result = append(result, &bindings[i])
	}
	return result
}

// LoadApplicationApprovalRulesetBindingsLatestVersions ...
func LoadApplicationApprovalRulesetBindingsLatestVersions(db *gorm.DB, organizationID string, bindings []*ApplicationApprovalRulesetBinding) error {
	reviewables := make([]IReviewable, 0, len(bindings))
	for _, binding := range bindings {
		reviewables = append(reviewables, binding)
	}

	return LoadReviewablesLatestVersions(
		db,
		reflect.TypeOf(ApplicationApprovalRulesetBinding{}.ApplicationApprovalRulesetBindingPrimaryKey),
		[]string{"application_id", "approval_ruleset_id"},
		reflect.TypeOf([]interface{}{}),
		reflect.TypeOf(ApplicationApprovalRulesetBindingMajorVersion{}),
		reflect.TypeOf(ApplicationApprovalRulesetBindingMajorVersion{}.ID),
		"application_approval_ruleset_binding_major_version_id",
		reflect.TypeOf(ApplicationApprovalRulesetBindingMinorVersion{}),
		organizationID,
		reviewables,
	)
}
