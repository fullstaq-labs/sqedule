package dbmodels

import (
	"reflect"

	"github.com/fullstaq-labs/sqedule/lib"
	"github.com/fullstaq-labs/sqedule/server/dbmodels/approvalrulesetbindingmode"
	"github.com/fullstaq-labs/sqedule/server/dbmodels/reviewstate"
	"github.com/fullstaq-labs/sqedule/server/dbutils"
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
	Enabled *bool `gorm:"not null; default:true"`

	Mode approvalrulesetbindingmode.Mode `gorm:"type:approval_ruleset_binding_mode; not null"`

	ApplicationApprovalRulesetBindingVersion ApplicationApprovalRulesetBindingVersion `gorm:"foreignKey:OrganizationID,ApplicationApprovalRulesetBindingVersionID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
}

//
// ******** ApplicationApprovalRulesetBinding methods ********
//

// NewDraftVersion returns an unsaved ApplicationApprovalRulesetBindingVersion and ApplicationApprovalRulesetBindingAdjustment
// in draft proposal state. Their contents are identical to the currently loaded Version and Adjustment.
func (binding ApplicationApprovalRulesetBinding) NewDraftVersion() (*ApplicationApprovalRulesetBindingVersion, *ApplicationApprovalRulesetBindingAdjustment) {
	var adjustment ApplicationApprovalRulesetBindingAdjustment
	var version *ApplicationApprovalRulesetBindingVersion = &adjustment.ApplicationApprovalRulesetBindingVersion

	if binding.Version != nil && binding.Version.Adjustment != nil {
		adjustment = *binding.Version.Adjustment
	}

	version.BaseModel = binding.BaseModel
	version.ReviewableVersionBase = ReviewableVersionBase{}
	version.ApplicationApprovalRulesetBinding = binding
	version.ApplicationID = binding.ApplicationID
	version.ApprovalRulesetID = binding.ApprovalRulesetID
	version.Adjustment = &adjustment

	adjustment.BaseModel = binding.BaseModel
	adjustment.ApplicationApprovalRulesetBindingVersionID = 0
	adjustment.ReviewableAdjustmentBase = ReviewableAdjustmentBase{
		AdjustmentNumber: 1,
		ReviewState:      reviewstate.Draft,
	}

	return version, &adjustment
}

func (binding ApplicationApprovalRulesetBinding) CheckNewProposalsRequireReview(action ReviewableAction, newMode approvalrulesetbindingmode.Mode) bool {
	switch action {
	case ReviewableActionCreate:
		return newMode == approvalrulesetbindingmode.Enforcing
	case ReviewableActionUpdate:
		if binding.Version == nil {
			return newMode == approvalrulesetbindingmode.Enforcing
		} else {
			return binding.Version.Adjustment.Mode != newMode
		}
	default:
		panic("Unsupported action " + action)
	}
}

//
// ******** ApplicationApprovalRulesetBindingAdjustment methods ********
//

// NewAdjustment returns an unsaved ApplicationApprovalRulesetBindingAdjustment in draft state. Its contents
// are identical to the previous Adjustment.
func (adjustment ApplicationApprovalRulesetBindingAdjustment) NewAdjustment() ApplicationApprovalRulesetBindingAdjustment {
	result := adjustment
	result.ReviewableAdjustmentBase = ReviewableAdjustmentBase{
		AdjustmentNumber: adjustment.AdjustmentNumber + 1,
		ReviewState:      reviewstate.Draft,
	}
	result.Enabled = lib.CopyBoolPtr(adjustment.Enabled)
	return result
}

//
// ******** ApplicationApprovalRulesetBindingAdjustment methods ********
//

func (adjustment ApplicationApprovalRulesetBindingAdjustment) IsEnabled() bool {
	return lib.DerefBoolPtrWithDefault(adjustment.Enabled, true)
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

func FindApplicationApprovalRulesetBinding(db *gorm.DB, organizationID string, applicationID string, rulesetID string) (ApplicationApprovalRulesetBinding, error) {
	var result ApplicationApprovalRulesetBinding

	tx := db.Where("organization_id = ? AND application_id = ? AND approval_ruleset_id = ?", organizationID, applicationID, rulesetID)
	tx.Take(&result)
	return result, dbutils.CreateFindOperationError(tx)
}

func FindApplicationApprovalRulesetBindingVersionByNumber(db *gorm.DB, organizationID string, applicationID string, rulesetID string, versionNumber uint32) (ApplicationApprovalRulesetBindingVersion, error) {
	var result ApplicationApprovalRulesetBindingVersion

	tx := db.Where("organization_id = ? AND application_id = ? AND approval_ruleset_id = ? AND version_number = ?", organizationID, applicationID, rulesetID, versionNumber)
	tx.Take(&result)
	return result, dbutils.CreateFindOperationError(tx)
}

func FindApplicationApprovalRulesetBindingVersionByID(db *gorm.DB, organizationID string, applicationID string, rulesetID string, versionID uint64) (ApplicationApprovalRulesetBindingVersion, error) {
	var result ApplicationApprovalRulesetBindingVersion

	tx := db.Where("organization_id = ? AND application_id = ? AND approval_ruleset_id = ? AND id = ?", organizationID, applicationID, rulesetID, versionID)
	tx.Take(&result)
	return result, dbutils.CreateFindOperationError(tx)
}

func FindApplicationApprovalRulesetBindingProposals(db *gorm.DB, organizationID string, applicationID string, rulesetID string) ([]ApplicationApprovalRulesetBindingVersion, error) {
	var result []ApplicationApprovalRulesetBindingVersion

	tx := db.Where("organization_id = ? AND application_id = ? AND approval_ruleset_id = ? AND version_number IS NULL", organizationID, applicationID, rulesetID)
	tx.Find(&result)
	return result, tx.Error
}

func FindApplicationApprovalRulesetBindingProposalByID(db *gorm.DB, organizationID string, applicationID string, rulesetID string, versionID uint64) (ApplicationApprovalRulesetBindingVersion, error) {
	return FindApplicationApprovalRulesetBindingVersionByID(db.Where("version_number IS NULL"), organizationID, applicationID, rulesetID, versionID)
}

// FindApplicationApprovalRulesetBindingVersions finds, for a given ApplicationApprovalRulesetBinding, all its Versions
// and returns them ordered by version number (descending).
//
// The `approved` parameter determines whether it finds approved or proposed versions.
func FindApplicationApprovalRulesetBindingVersions(db *gorm.DB, organizationID string, applicationID string, rulesetID string, approved bool, pagination dbutils.PaginationOptions) ([]ApplicationApprovalRulesetBindingVersion, error) {
	var result []ApplicationApprovalRulesetBindingVersion

	tx := db.Where("organization_id = ? AND application_id = ? AND approval_ruleset_id = ?", organizationID, applicationID, rulesetID)
	if approved {
		tx = tx.Where("version_number IS NOT NULL").Order("version_number DESC")
	} else {
		tx = tx.Where("version_number IS NULL")
	}
	tx = dbutils.ApplyDbQueryPaginationOptions(tx, pagination)
	tx.Find(&result)
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
// ******** Deletion functions ********
//

func DeleteApplicationApprovalRulesetBindingAdjustmentsForProposal(db *gorm.DB, organizationID string, proposalID uint64) error {
	return db.
		Where("organization_id = ? AND application_approval_ruleset_binding_version_id = ?", organizationID, proposalID).
		Delete(ApplicationApprovalRulesetBindingAdjustment{}).
		Error
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

// MakeApplicationApprovalRulesetBindingVersionsPointerArray turns a `[]ApplicationApprovalRulesetBindingVersion` into a `[]*ApplicationApprovalRulesetBindingVersion`.
func MakeApplicationApprovalRulesetBindingVersionsPointerArray(versions []ApplicationApprovalRulesetBindingVersion) []*ApplicationApprovalRulesetBindingVersion {
	result := make([]*ApplicationApprovalRulesetBindingVersion, 0, len(versions))
	for i := range versions {
		result = append(result, &versions[i])
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

// CollectApplicationApprovalRulesetBindingVersionIDEquals returns the first ApplicationApprovalRulesetBindingVersion
// whose ID equals `versionID`.
func CollectApplicationApprovalRulesetBindingVersionIDEquals(versions []ApplicationApprovalRulesetBindingVersion, versionID uint64) *ApplicationApprovalRulesetBindingVersion {
	for i := range versions {
		if versions[i].ID == versionID {
			return &versions[i]
		}
	}
	return nil
}

// CollectApplicationApprovalRulesetBindingVersionIDNotEquals returns those ApplicationApprovalRulesetBindingVersion
// whose IDs don't equal `versionID`.
func CollectApplicationApprovalRulesetBindingVersionIDNotEquals(versions []ApplicationApprovalRulesetBindingVersion, versionID uint64) []*ApplicationApprovalRulesetBindingVersion {
	l := len(versions)
	if l > 0 {
		l -= 1
	}

	result := make([]*ApplicationApprovalRulesetBindingVersion, 0, l)
	for i := range versions {
		if versions[i].ID != versionID {
			result = append(result, &versions[i])
		}
	}
	return result
}
