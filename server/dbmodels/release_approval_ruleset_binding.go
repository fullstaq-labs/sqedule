package dbmodels

import (
	"github.com/fullstaq-labs/sqedule/server/dbmodels/approvalrulesetbindingmode"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

//
// ******** Types, constants & variables ********
//

type ReleaseApprovalRulesetBinding struct {
	BaseModel

	ApplicationID string  `gorm:"type:citext; primaryKey; not null"`
	ReleaseID     uint64  `gorm:"primaryKey; not null"`
	Release       Release `gorm:"foreignKey:OrganizationID,ApplicationID,ReleaseID; references:OrganizationID,ApplicationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`

	ApprovalRulesetID string          `gorm:"type:citext; primaryKey; not null"`
	ApprovalRuleset   ApprovalRuleset `gorm:"foreignKey:OrganizationID,ApprovalRulesetID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`

	ApprovalRulesetVersionID uint64                 `gorm:"not null"`
	ApprovalRulesetVersion   ApprovalRulesetVersion `gorm:"foreignKey:OrganizationID,ApprovalRulesetVersionID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`

	ApprovalRulesetAdjustmentNumber uint32                    `gorm:"type:int; not null"`
	ApprovalRulesetAdjustment       ApprovalRulesetAdjustment `gorm:"foreignKey:OrganizationID,ApprovalRulesetVersionID,ApprovalRulesetAdjustmentNumber; references:OrganizationID,ApprovalRulesetVersionID,AdjustmentNumber; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`

	Mode approvalrulesetbindingmode.Mode `gorm:"type:approval_ruleset_binding_mode; not null"`
}

//
// ******** Constructor functions ********
//

func NewReleaseApprovalRulesetBindingFromApplicationApprovalRulesetBinding(appRuleBinding ApplicationApprovalRulesetBinding) ReleaseApprovalRulesetBinding {
	if appRuleBinding.Version == nil {
		panic("Given binding must have an associated version")
	}
	if appRuleBinding.Version.Adjustment == nil {
		panic("Given binding must have an associated adjustment")
	}

	if appRuleBinding.ApprovalRuleset.Version == nil {
		panic("Given binding's ruleset must have an associated version")
	}
	if appRuleBinding.ApprovalRuleset.Version.VersionNumber == nil {
		panic("Given binding's ruleset's associated version must be finalized")
	}
	if appRuleBinding.ApprovalRuleset.Version.Adjustment == nil {
		panic("Given binding ruleset's must have an associated adjustment")
	}

	return ReleaseApprovalRulesetBinding{
		BaseModel:                       appRuleBinding.BaseModel,
		ApplicationID:                   appRuleBinding.ApplicationID,
		ApprovalRulesetID:               appRuleBinding.ApprovalRulesetID,
		ApprovalRuleset:                 appRuleBinding.ApprovalRuleset,
		ApprovalRulesetVersionID:        appRuleBinding.ApprovalRuleset.Version.ID,
		ApprovalRulesetVersion:          *appRuleBinding.ApprovalRuleset.Version,
		ApprovalRulesetAdjustmentNumber: appRuleBinding.ApprovalRuleset.Version.Adjustment.AdjustmentNumber,
		ApprovalRulesetAdjustment:       *appRuleBinding.ApprovalRuleset.Version.Adjustment,
		Mode:                            appRuleBinding.Version.Adjustment.Mode,
	}
}

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

//
// ******** Find/load functions ********
//

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
