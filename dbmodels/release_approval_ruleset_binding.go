package dbmodels

import (
	"github.com/fullstaq-labs/sqedule/dbmodels/approvalrulesetbindingmode"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ReleaseApprovalRulesetBinding ...
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

	Mode approvalrulesetbindingmode.Mode `gorm:"type:approval_ruleset_binding_mode; not null"`
}

// NewReleaseApprovalRulesetBindingFromApplicationApprovalRulesetBinding ...
func NewReleaseApprovalRulesetBindingFromApplicationApprovalRulesetBinding(appRuleBinding ApplicationApprovalRulesetBinding) ReleaseApprovalRulesetBinding {
	if appRuleBinding.LatestMinorVersion == nil {
		panic("Given binding must have an associated latest minor version")
	}

	if appRuleBinding.ApprovalRuleset.LatestMajorVersion == nil {
		panic("Given binding's ruleset must have an associated latest major version")
	}
	if appRuleBinding.ApprovalRuleset.LatestMajorVersion.VersionNumber == nil {
		panic("Given binding's ruleset's associated latest major version must be finalized")
	}
	if appRuleBinding.ApprovalRuleset.LatestMinorVersion == nil {
		panic("Given binding ruleset's must have an associated latest minor version")
	}

	return ReleaseApprovalRulesetBinding{
		BaseModel:                         appRuleBinding.BaseModel,
		ApplicationID:                     appRuleBinding.ApplicationID,
		ApprovalRulesetID:                 appRuleBinding.ApprovalRulesetID,
		ApprovalRuleset:                   appRuleBinding.ApprovalRuleset,
		ApprovalRulesetMajorVersionID:     appRuleBinding.ApprovalRuleset.LatestMajorVersion.ID,
		ApprovalRulesetMajorVersion:       *appRuleBinding.ApprovalRuleset.LatestMajorVersion,
		ApprovalRulesetMinorVersionNumber: appRuleBinding.ApprovalRuleset.LatestMinorVersion.VersionNumber,
		ApprovalRulesetMinorVersion:       *appRuleBinding.ApprovalRuleset.LatestMinorVersion,
		Mode:                              appRuleBinding.LatestMinorVersion.Mode,
	}
}

// CreateReleaseApprovalRulesetBindings ...
func CreateReleaseApprovalRulesetBindings(db *gorm.DB, releaseID uint64, appRuleBindings []ApplicationApprovalRulesetBinding) ([]ReleaseApprovalRulesetBinding, error) {
	result := make([]ReleaseApprovalRulesetBinding, 0, len(appRuleBindings))

	for _, appRuleBinding := range appRuleBindings {
		releaseRuleBinding := NewReleaseApprovalRulesetBindingFromApplicationApprovalRulesetBinding(appRuleBinding)
		releaseRuleBinding.ReleaseID = releaseID
		tx := db.Omit(clause.Associations).Create(&releaseRuleBinding)
		if tx.Error != nil {
			return []ReleaseApprovalRulesetBinding{}, tx.Error
		}

		result = append(result, releaseRuleBinding)
	}

	return result, nil
}

// FindAllReleaseApprovalRulesetBindings ...
func FindAllReleaseApprovalRulesetBindings(db *gorm.DB, organizationID string, applicationID string, releaseID uint64) ([]ReleaseApprovalRulesetBinding, error) {
	var result []ReleaseApprovalRulesetBinding
	tx := db.Where("organization_id = ? AND application_id = ? AND release_id = ?",
		organizationID, applicationID, releaseID)
	tx = tx.Find(&result)
	return result, tx.Error
}

func FindAllReleaseApprovalRulesetBindingsWithApprovalRuleset(db *gorm.DB, organizationID string, rulesetID string) ([]ReleaseApprovalRulesetBinding, error) {
	var result []ReleaseApprovalRulesetBinding
	tx := db.Where("organization_id = ? AND approval_ruleset_id = ?", organizationID, rulesetID)
	tx = tx.Find(&result)
	return result, tx.Error
}
