package dbmodels

import (
	"reflect"

	"github.com/fullstaq-labs/sqedule/server/dbmodels/approvalrulesetbindingmode"
	"gorm.io/gorm"
)

//
// ******** Types, constants & variables ********
//

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

	Version *ApplicationApprovalRulesetBindingVersion `gorm:"-"`
}

type ApplicationApprovalRulesetBindingVersion struct {
	BaseModel
	ApplicationID     string `gorm:"type:citext; not null"`
	ApprovalRulesetID string `gorm:"type:citext; not null"`
	ReviewableVersionBase

	ApplicationApprovalRulesetBinding ApplicationApprovalRulesetBinding            `gorm:"foreignKey:OrganizationID,ApplicationID,ApprovalRulesetID; references:OrganizationID,ApplicationID,ApprovalRulesetID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	Adjustment                        *ApplicationApprovalRulesetBindingAdjustment `gorm:"-"`
}

type ApplicationApprovalRulesetBindingAdjustment struct {
	BaseModel
	ApplicationApprovalRulesetBindingVersionID uint64 `gorm:"primaryKey; not null"`
	ReviewableAdjustmentBase
	Enabled bool `gorm:"not null; default:true"`

	Mode approvalrulesetbindingmode.Mode `gorm:"type:approval_ruleset_binding_mode; not null"`

	ApplicationApprovalRulesetBindingVersion ApplicationApprovalRulesetBindingVersion `gorm:"foreignKey:OrganizationID,ApplicationApprovalRulesetBindingVersionID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
}

//
// ******** Find/load functions ********
//

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

func LoadApplicationApprovalRulesetBindingsLatestVersionsAndAdjustments(db *gorm.DB, organizationID string, bindings []*ApplicationApprovalRulesetBinding) error {
	err := LoadApplicationApprovalRulesetBindingsLatestVersions(db, organizationID, bindings)
	if err != nil {
		return err
	}

	return LoadApplicationApprovalRulesetBindingVersionsLatestAdjustments(db, organizationID, CollectApplicationApprovalRulesetBindingVersions(bindings))
}

func LoadApplicationApprovalRulesetBindingsLatestVersions(db *gorm.DB, organizationID string, bindings []*ApplicationApprovalRulesetBinding) error {
	reviewables := make([]IReviewable, 0, len(bindings))
	for _, binding := range bindings {
		reviewables = append(reviewables, binding)
	}

	return LoadReviewablesLatestVersions(
		db,
		organizationID,
		reviewables,
		reflect.TypeOf(ApplicationApprovalRulesetBindingVersion{}),
		[]string{"application_id", "approval_ruleset_id"},
	)
}

func LoadApplicationApprovalRulesetBindingVersionsLatestAdjustments(db *gorm.DB, organizationID string, versions []*ApplicationApprovalRulesetBindingVersion) error {
	iversions := make([]IReviewableVersion, 0, len(versions))
	for _, version := range versions {
		iversions = append(iversions, version)
	}

	return LoadReviewableVersionsLatestAdjustments(
		db,
		organizationID,
		iversions,
		reflect.TypeOf(ApplicationApprovalRulesetBindingAdjustment{}),
		"application_approval_ruleset_binding_version_id",
	)
}

//
// ******** Other functions ********
//

func MakeApplicationApprovalRulesetBindingsPointerArray(bindings []ApplicationApprovalRulesetBinding) []*ApplicationApprovalRulesetBinding {
	result := make([]*ApplicationApprovalRulesetBinding, 0, len(bindings))
	for i := range bindings {
		result = append(result, &bindings[i])
	}
	return result
}

// CollectApplicationApprovalRulesetBindingVersions turns a `[]*ApplicationApprovalRulesetBinding`
// into a list of their associated ApplicationApprovalRulesetBindingVersions. It does not include nils.
func CollectApplicationApprovalRulesetBindingVersions(bindings []*ApplicationApprovalRulesetBinding) []*ApplicationApprovalRulesetBindingVersion {
	result := make([]*ApplicationApprovalRulesetBindingVersion, 0, len(bindings))
	for _, elem := range bindings {
		if elem.Version != nil {
			result = append(result, elem.Version)
		}
	}
	return result
}
