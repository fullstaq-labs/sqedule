package dbmodels

import (
	"reflect"

	"github.com/fullstaq-labs/sqedule/lib"
	"github.com/fullstaq-labs/sqedule/server/dbmodels/proposalstate"
	"github.com/fullstaq-labs/sqedule/server/dbutils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

//
// ******** Types, constants & variables ********
//

type ApprovalRuleset struct {
	BaseModel
	ID string `gorm:"type:citext; primaryKey; not null"`
	ReviewableBase

	Version *ApprovalRulesetVersion `gorm:"-"`
}

type ApprovalRulesetVersion struct {
	BaseModel
	ReviewableVersionBase
	ApprovalRulesetID string          `gorm:"type:citext; not null"`
	ApprovalRuleset   ApprovalRuleset `gorm:"foreignKey:OrganizationID,ApprovalRulesetID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`

	Adjustment *ApprovalRulesetAdjustment `gorm:"-"`
}

type ApprovalRulesetAdjustment struct {
	BaseModel
	ApprovalRulesetVersionID uint64 `gorm:"primaryKey; not null"`
	ReviewableAdjustmentBase
	Enabled *bool `gorm:"not null; default:true"`

	DisplayName string `gorm:"not null"`
	Description string `gorm:"not null"`
	// TODO: this doesn't work because of the lack of a rule binding mode. move to level of rule binding.
	GloballyApplicable bool `gorm:"not null; default:false"`

	ApprovalRulesetVersion ApprovalRulesetVersion `gorm:"foreignKey:OrganizationID,ApprovalRulesetVersionID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`

	Rules ApprovalRulesetContents `gorm:"-"`
	// Set by LoadApprovalRulesetAdjustmentsStats. Not a real table column, so don't add to migrations.
	NumBoundReleases uint `gorm:"<-:false"`
}

// ApprovalRulesetVersionAndAdjustmentKey uniquely identifies a specific Version+Adjustment
// of an ApprovalRuleset.
type ApprovalRulesetVersionAndAdjustmentKey struct {
	VersionID        uint64
	AdjustmentNumber uint32
}

type ApprovalRulesetWithStats struct {
	ApprovalRuleset
	NumBoundApplications uint `gorm:"<-:false"`
	NumBoundReleases     uint `gorm:"<-:false"`
}

// ApprovalRulesetContents is a collection of ApprovalRules.
//
// This is unlike ApprovalRuleset, which is a database model representing the `approval_rulesets`
// database table. ApprovalRuleset is not capable of actually physically containing all the
// associated ApprovalRules. In contrast, ApprovalRulesetContents *is* a container which
// physically contains ApprovalRules.
//
// ApprovalRulesetContents can contain all supported ApprovalRule types in a way that doesn't
// use pointers, and that doesn't require typecasting. It's a more efficient alternative to
// `[]*ApprovalRule`. This latter requires pointers and is thus not as memory-efficient.
// Furthermore, this latter requires the use of casting to find out which specific subtype an
// element is.
type ApprovalRulesetContents struct {
	HTTPApiApprovalRules  []HTTPApiApprovalRule
	ScheduleApprovalRules []ScheduleApprovalRule
	ManualApprovalRules   []ManualApprovalRule
}

//
// ******** ApprovalRulesetContents methods ********
//

// NumRules returns the total number of rules in this ApprovalRulesetContents.
func (c ApprovalRulesetContents) NumRules() uint {
	return uint(len(c.HTTPApiApprovalRules)) +
		uint(len(c.ScheduleApprovalRules)) +
		uint(len(c.ManualApprovalRules))
}

func (c ApprovalRulesetContents) CopyAsUnsaved() ApprovalRulesetContents {
	//nolint:errcheck
	c.ForEach(func(rule IApprovalRule) error {
		rule.ClearPrimaryKey()
		rule.AssociateWithApprovalRulesetAdjustment(ApprovalRulesetAdjustment{})
		return nil
	})
	return c
}

func (c *ApprovalRulesetContents) ForEach(callback func(rule IApprovalRule) error) error {
	var ruleTypesProcessed uint = 0
	var err error

	ruleTypesProcessed++
	for i := range c.HTTPApiApprovalRules {
		err = callback(&c.HTTPApiApprovalRules[i])
		if err != nil {
			return err
		}
	}

	ruleTypesProcessed++
	for i := range c.ScheduleApprovalRules {
		err = callback(&c.ScheduleApprovalRules[i])
		if err != nil {
			return err
		}
	}

	ruleTypesProcessed++
	for i := range c.ManualApprovalRules {
		err = callback(&c.ManualApprovalRules[i])
		if err != nil {
			return err
		}
	}

	if ruleTypesProcessed != NumApprovalRuleTypes {
		panic("Bug: code does not cover all approval rule types")
	}

	return err
}

//
// ******** ApprovalRuleset methods ********
//

// NewDraftVersion returns an unsaved ApprovalRulesetVersion and ApprovalRulesetAdjustment
// in draft proposal state. Their contents are identical to the currently loaded Version and Adjustment.
func (ruleset ApprovalRuleset) NewDraftVersion() (*ApprovalRulesetVersion, *ApprovalRulesetAdjustment) {
	var adjustment ApprovalRulesetAdjustment
	var version *ApprovalRulesetVersion = &adjustment.ApprovalRulesetVersion

	if ruleset.Version != nil && ruleset.Version.Adjustment != nil {
		adjustment = *ruleset.Version.Adjustment
	}

	version.BaseModel = ruleset.BaseModel
	version.ReviewableVersionBase = ReviewableVersionBase{}
	version.ApprovalRuleset = ruleset
	version.ApprovalRulesetID = ruleset.ID
	version.Adjustment = &adjustment

	adjustment.BaseModel = ruleset.BaseModel
	adjustment.ApprovalRulesetVersionID = 0
	adjustment.ReviewableAdjustmentBase = ReviewableAdjustmentBase{
		AdjustmentNumber: 1,
		ProposalState:    proposalstate.Draft,
	}

	return version, &adjustment
}

func (ruleset ApprovalRuleset) CheckNewProposalsRequireReview(action ReviewableAction, hasBoundApplications bool, rulesChanged bool) bool {
	return false
	// switch action {
	// case ReviewableActionCreate:
	// 	return false
	// case ReviewableActionUpdate:
	// 	return hasBoundApplications && rulesChanged
	// default:
	// 	panic("Unsupported action " + action)
	// }
}

//
// ******** ApprovalRulesetAdjustment methods ********
//

func (adjustment ApprovalRulesetAdjustment) IsEnabled() bool {
	return lib.DerefBoolPtrWithDefault(adjustment.Enabled, true)
}

func (adjustment ApprovalRulesetAdjustment) ApprovalRulesetVersionAndAdjustmentKey() ApprovalRulesetVersionAndAdjustmentKey {
	return ApprovalRulesetVersionAndAdjustmentKey{
		VersionID:        adjustment.ApprovalRulesetVersionID,
		AdjustmentNumber: adjustment.AdjustmentNumber,
	}
}

// NewAdjustment returns an unsaved ApprovalRulesetAdjustment in draft state. Its contents
// are identical to the previous Adjustment. It even copies over the previous Adjustment's Rules,
// but the new Rules are unsaved.
func (adjustment ApprovalRulesetAdjustment) NewAdjustment() ApprovalRulesetAdjustment {
	result := adjustment
	result.ReviewableAdjustmentBase = ReviewableAdjustmentBase{
		AdjustmentNumber: adjustment.AdjustmentNumber + 1,
		ProposalState:    proposalstate.Draft,
	}
	result.Enabled = lib.CopyBoolPtr(adjustment.Enabled)
	result.Rules = adjustment.Rules.CopyAsUnsaved()
	return result
}

// Create creates a database record for this adjustment as well as for its rules.
func (adjustment *ApprovalRulesetAdjustment) Create(db *gorm.DB) error {
	tx := db.Omit(clause.Associations).Create(adjustment)
	if tx.Error != nil {
		return tx.Error
	}

	return adjustment.Rules.ForEach(func(rule IApprovalRule) error {
		rule.AssociateWithApprovalRulesetAdjustment(*adjustment)
		return db.Omit(clause.Associations).Create(rule).Error
	})
}

//
// ******** Find/load functions ********
//

func FindApprovalRulesetsWithStats(db *gorm.DB, organizationID string, pagination dbutils.PaginationOptions) ([]ApprovalRulesetWithStats, error) {
	var result []ApprovalRulesetWithStats
	tx := db.
		Model(&ApprovalRuleset{}).
		Select("approval_rulesets.*, "+
			"bound_apps_stats.num_bound_applications, "+
			"bound_releases_stats.num_bound_releases").
		Where("approval_rulesets.organization_id = ?", organizationID).
		Joins("LEFT JOIN ("+
			"SELECT approval_rulesets.id, "+
			"  COUNT(application_approval_ruleset_bindings.application_id) AS num_bound_applications "+
			"FROM approval_rulesets "+
			"LEFT JOIN application_approval_ruleset_bindings "+
			"  ON application_approval_ruleset_bindings.organization_id = approval_rulesets.organization_id "+
			"  AND application_approval_ruleset_bindings.approval_ruleset_id = approval_rulesets.id "+
			"WHERE approval_rulesets.organization_id = ? "+
			"GROUP BY approval_rulesets.organization_id, approval_rulesets.id "+
			") bound_apps_stats "+
			"ON bound_apps_stats.id = approval_rulesets.id",
			organizationID,
		).
		Joins("LEFT JOIN ("+
			"SELECT approval_rulesets.id, "+
			"  COUNT(CASE WHEN release_approval_ruleset_bindings.application_id IS NULL "+
			"      AND release_approval_ruleset_bindings.release_id IS NULL "+
			"    THEN NULL "+
			"    ELSE (release_approval_ruleset_bindings.application_id, release_approval_ruleset_bindings.release_id) "+
			"    END) AS num_bound_releases "+
			"FROM approval_rulesets "+
			"LEFT JOIN release_approval_ruleset_bindings "+
			"  ON release_approval_ruleset_bindings.organization_id = approval_rulesets.organization_id "+
			"  AND release_approval_ruleset_bindings.approval_ruleset_id = approval_rulesets.id "+
			"WHERE approval_rulesets.organization_id = ? "+
			"GROUP BY approval_rulesets.organization_id, approval_rulesets.id "+
			") bound_releases_stats "+
			"ON bound_releases_stats.id = approval_rulesets.id",
			organizationID,
		)
	tx = dbutils.ApplyDbQueryPaginationOptions(tx, pagination)
	tx = tx.Find(&result)
	return result, tx.Error
}

func FindApprovalRuleset(db *gorm.DB, organizationID string, id string) (ApprovalRuleset, error) {
	var result ApprovalRuleset

	tx := db.Where("organization_id = ? AND id = ?", organizationID, id)
	tx.Take(&result)
	return result, dbutils.CreateFindOperationError(tx)
}

func FindApprovalRulesetVersionByNumber(db *gorm.DB, organizationID string, rulesetID string, versionNumber uint32) (ApprovalRulesetVersion, error) {
	var result ApprovalRulesetVersion

	tx := db.Where("organization_id = ? AND approval_ruleset_id = ? AND version_number = ?", organizationID, rulesetID, versionNumber)
	tx.Take(&result)
	return result, dbutils.CreateFindOperationError(tx)
}

func FindApprovalRulesetVersionByID(db *gorm.DB, organizationID string, rulesetID string, versionID uint64) (ApprovalRulesetVersion, error) {
	var result ApprovalRulesetVersion

	tx := db.Where("organization_id = ? AND approval_ruleset_id = ? AND id = ?", organizationID, rulesetID, versionID)
	tx.Take(&result)
	return result, dbutils.CreateFindOperationError(tx)
}

func FindApprovalRulesetProposals(db *gorm.DB, organizationID string, rulesetID string) ([]ApprovalRulesetVersion, error) {
	var result []ApprovalRulesetVersion

	tx := db.Where("organization_id = ? AND approval_ruleset_id = ? AND version_number IS NULL", organizationID, rulesetID)
	tx.Find(&result)
	return result, tx.Error
}

func FindApprovalRulesetProposalByID(db *gorm.DB, organizationID string, rulesetID string, versionID uint64) (ApprovalRulesetVersion, error) {
	return FindApprovalRulesetVersionByID(db.Where("version_number IS NULL"), organizationID, rulesetID, versionID)
}

// FindApprovalRulesetVersions finds, for a given ApprovalRuleset, all its Versions
// and returns them ordered by version number (descending).
//
// The `approved` parameter determines whether it finds approved or proposed versions.
func FindApprovalRulesetVersions(db *gorm.DB, organizationID string, rulesetID string, approved bool, pagination dbutils.PaginationOptions) ([]ApprovalRulesetVersion, error) {
	var result []ApprovalRulesetVersion

	tx := db.Where("organization_id = ? AND approval_ruleset_id = ?", organizationID, rulesetID)
	if approved {
		tx = tx.Where("version_number IS NOT NULL").Order("version_number DESC")
	} else {
		tx = tx.Where("version_number IS NULL")
	}
	tx = dbutils.ApplyDbQueryPaginationOptions(tx, pagination)
	tx.Find(&result)
	return result, tx.Error
}

func FindApprovalRulesetAdjustments(db *gorm.DB, organizationID string, versionID uint64) ([]ApprovalRulesetAdjustment, error) {
	var result []ApprovalRulesetAdjustment

	tx := db.Where("organization_id = ? AND approval_ruleset_version_id = ?", organizationID, versionID)
	tx.Find(&result)
	return result, tx.Error
}

func LoadApprovalRulesetsLatestVersionsAndAdjustments(db *gorm.DB, organizationID string, rulesets []*ApprovalRuleset) error {
	err := LoadApprovalRulesetsLatestVersions(db, organizationID, rulesets)
	if err != nil {
		return err
	}

	return LoadApprovalRulesetVersionsLatestAdjustments(db, organizationID, CollectApprovalRulesetVersions(rulesets))
}

func LoadApprovalRulesetsLatestVersions(db *gorm.DB, organizationID string, rulesets []*ApprovalRuleset) error {
	reviewables := make([]IReviewable, 0, len(rulesets))
	for _, ruleset := range rulesets {
		reviewables = append(reviewables, ruleset)
	}

	return LoadReviewablesLatestVersions(
		db,
		organizationID,
		reviewables,
		reflect.TypeOf(ApprovalRulesetVersion{}),
		[]string{"approval_ruleset_id"},
	)
}

func LoadApprovalRulesetVersionsLatestAdjustments(db *gorm.DB, organizationID string, versions []*ApprovalRulesetVersion) error {
	iversions := make([]IReviewableVersion, 0, len(versions))
	for _, version := range versions {
		iversions = append(iversions, version)
	}

	return LoadReviewableVersionsLatestAdjustments(
		db,
		organizationID,
		iversions,
		reflect.TypeOf(ApprovalRulesetAdjustment{}),
		"approval_ruleset_version_id",
	)
}

func LoadApprovalRulesetAdjustmentsApprovalRules(db *gorm.DB, organizationID string, adjustments []*ApprovalRulesetAdjustment) error {
	var adjustmentIndex map[ApprovalRulesetVersionAndAdjustmentKey][]*ApprovalRulesetAdjustment = indexAdjustmentsByKey(adjustments)
	var query, tx *gorm.DB
	var ruleTypesProcessed uint = 0
	var httpAPIApprovalRules []HTTPApiApprovalRule
	var scheduleApprovalRules []ScheduleApprovalRule
	var manualApprovalRules []ManualApprovalRule

	query = db.Where("organization_id = ? AND (approval_ruleset_version_id, approval_ruleset_adjustment_number) IN ?",
		organizationID, collectApprovalRulesetAdjustmentsQueryValues(adjustments))

	ruleTypesProcessed++
	tx = db.Where(query).Find(&httpAPIApprovalRules)
	if tx.Error != nil {
		return tx.Error
	}
	for _, rule := range httpAPIApprovalRules {
		key := rule.ApprovalRulesetVersionAndAdjustmentKey()
		matchingAdjustments := adjustmentIndex[key]
		for _, adjustment := range matchingAdjustments {
			adjustment.Rules.HTTPApiApprovalRules = append(adjustment.Rules.HTTPApiApprovalRules, rule)
		}
	}

	ruleTypesProcessed++
	tx = db.Where(query).Find(&scheduleApprovalRules)
	if tx.Error != nil {
		return tx.Error
	}
	for _, rule := range scheduleApprovalRules {
		key := rule.ApprovalRulesetVersionAndAdjustmentKey()
		matchingAdjustments := adjustmentIndex[key]
		for _, adjustment := range matchingAdjustments {
			adjustment.Rules.ScheduleApprovalRules = append(adjustment.Rules.ScheduleApprovalRules, rule)
		}
	}

	ruleTypesProcessed++
	tx = db.Where(query).Find(&manualApprovalRules)
	if tx.Error != nil {
		return tx.Error
	}
	for _, rule := range manualApprovalRules {
		key := rule.ApprovalRulesetVersionAndAdjustmentKey()
		matchingAdjustments := adjustmentIndex[key]
		for _, adjustment := range matchingAdjustments {
			adjustment.Rules.ManualApprovalRules = append(adjustment.Rules.ManualApprovalRules, rule)
		}
	}

	if ruleTypesProcessed != NumApprovalRuleTypes {
		panic("Bug: code does not cover all approval rule types")
	}

	return nil
}

func indexAdjustmentsByKey(adjustments []*ApprovalRulesetAdjustment) map[ApprovalRulesetVersionAndAdjustmentKey][]*ApprovalRulesetAdjustment {
	result := make(map[ApprovalRulesetVersionAndAdjustmentKey][]*ApprovalRulesetAdjustment, len(adjustments))
	for _, adjustment := range adjustments {
		key := adjustment.ApprovalRulesetVersionAndAdjustmentKey()
		list, ok := result[key]
		if !ok {
			list = make([]*ApprovalRulesetAdjustment, 0, 1)
		}
		result[key] = append(list, adjustment)
	}
	return result
}

func collectApprovalRulesetAdjustmentsQueryValues(adjustments []*ApprovalRulesetAdjustment) [][]uint64 {
	result := make([][]uint64, 0, len(adjustments))
	for _, adjustment := range adjustments {
		elem := make([]uint64, 0, 2)
		elem = append(elem, adjustment.ApprovalRulesetVersionID)
		elem = append(elem, uint64(adjustment.AdjustmentNumber))
		result = append(result, elem)
	}
	return result
}

func LoadApprovalRulesetAdjustmentsStats(db *gorm.DB, organizationID string, adjustments []*ApprovalRulesetAdjustment) error {
	var adjustmentIndex map[ApprovalRulesetVersionAndAdjustmentKey][]*ApprovalRulesetAdjustment = indexAdjustmentsByKey(adjustments)
	var result = make([]struct {
		ApprovalRulesetVersionID uint64
		AdjustmentNumber         uint32
		NumBoundReleases         uint
	}, 0, len(adjustments))

	tx := db.
		Model(&ApprovalRulesetAdjustment{}).
		Select("approval_ruleset_adjustments.approval_ruleset_version_id, "+
			"approval_ruleset_adjustments.adjustment_number, "+
			"COUNT(release_approval_ruleset_bindings.*) AS num_bound_releases").
		Where("approval_ruleset_adjustments.organization_id = ? "+
			"AND (approval_ruleset_adjustments.approval_ruleset_version_id, approval_ruleset_adjustments.adjustment_number) IN ?",
			organizationID, collectApprovalRulesetAdjustmentsQueryValues(adjustments)).
		Joins("LEFT JOIN release_approval_ruleset_bindings "+
			"ON release_approval_ruleset_bindings.organization_id = ? "+
			"AND release_approval_ruleset_bindings.approval_ruleset_version_id = approval_ruleset_adjustments.approval_ruleset_version_id "+
			"AND release_approval_ruleset_bindings.approval_ruleset_adjustment_number = approval_ruleset_adjustments.adjustment_number",
			organizationID).
		Group("approval_ruleset_adjustments.organization_id, " +
			"approval_ruleset_adjustments.approval_ruleset_version_id, " +
			"approval_ruleset_adjustments.adjustment_number")
	tx = tx.Find(&result)
	if tx.Error != nil {
		return tx.Error
	}

	for _, elem := range result {
		key := ApprovalRulesetVersionAndAdjustmentKey{
			VersionID:        elem.ApprovalRulesetVersionID,
			AdjustmentNumber: elem.AdjustmentNumber,
		}
		matchingAdjustments := adjustmentIndex[key]
		for _, adjustment := range matchingAdjustments {
			adjustment.NumBoundReleases = elem.NumBoundReleases
		}
	}

	return nil
}

//
// ******** Deletion functions ********
//

func DeleteApprovalRulesetAdjustmentsForProposal(db *gorm.DB, organizationID string, proposalID uint64) error {
	return db.
		Where("organization_id = ? AND approval_ruleset_version_id = ?", organizationID, proposalID).
		Delete(ApprovalRulesetAdjustment{}).
		Error
}

//
// ******** Other functions ********
//

// MakeApprovalRulesetVersionsPointerArray turns a `[]ApprovalRulesetVersion` into a `[]*ApprovalRulesetVersion`.
func MakeApprovalRulesetVersionsPointerArray(versions []ApprovalRulesetVersion) []*ApprovalRulesetVersion {
	result := make([]*ApprovalRulesetVersion, 0, len(versions))
	for i := range versions {
		result = append(result, &versions[i])
	}
	return result
}

// MakeApprovalRulesetAdjustmentsPointerArray turns a `[]ApprovalRulesetAdjustment` into a `[]*ApprovalRulesetAdjustment`.
func MakeApprovalRulesetAdjustmentsPointerArray(adjustment []ApprovalRulesetAdjustment) []*ApprovalRulesetAdjustment {
	result := make([]*ApprovalRulesetAdjustment, 0, len(adjustment))
	for i := range adjustment {
		result = append(result, &adjustment[i])
	}
	return result
}

func CollectApprovalRulesetsWithoutStats(rulesets []ApprovalRulesetWithStats) []*ApprovalRuleset {
	result := make([]*ApprovalRuleset, 0, len(rulesets))
	for i := range rulesets {
		ruleset := &rulesets[i]
		result = append(result, &ruleset.ApprovalRuleset)
	}
	return result
}

func CollectApprovalRulesetsWithApplicationApprovalRulesetBindings(bindings []ApplicationApprovalRulesetBinding) []*ApprovalRuleset {
	result := make([]*ApprovalRuleset, 0, len(bindings))
	for i := range bindings {
		binding := &bindings[i]
		result = append(result, &binding.ApprovalRuleset)
	}
	return result
}

// CollectApprovalRulesetVersions turns a `[]*ApprovalRuleset` into a list of their associated ApprovalRulesetVersions.
// It does not include nils.
func CollectApprovalRulesetVersions(rulesets []*ApprovalRuleset) []*ApprovalRulesetVersion {
	result := make([]*ApprovalRulesetVersion, 0, len(rulesets))
	for _, elem := range rulesets {
		if elem.Version != nil {
			result = append(result, elem.Version)
		}
	}
	return result
}

// CollectApprovalRulesetVersionIDEquals returns the first ApprovalRulesetVersion
// whose ID equals `versionID`.
func CollectApprovalRulesetVersionIDEquals(versions []ApprovalRulesetVersion, versionID uint64) *ApprovalRulesetVersion {
	for i := range versions {
		if versions[i].ID == versionID {
			return &versions[i]
		}
	}
	return nil
}

// CollectApprovalRulesetVersionIDNotEquals returns those ApprovalRulesetVersion
// whose IDs don't equal `versionID`.
func CollectApprovalRulesetVersionIDNotEquals(versions []ApprovalRulesetVersion, versionID uint64) []*ApprovalRulesetVersion {
	l := len(versions)
	if l > 0 {
		l -= 1
	}

	result := make([]*ApprovalRulesetVersion, 0, l)
	for i := range versions {
		if versions[i].ID != versionID {
			result = append(result, &versions[i])
		}
	}
	return result
}

// CollectApprovalRulesetAdjustmentsFromVersions turns a `[]*ApprovalRulesetVersion` into a list of their associated ApprovalRulesetAdjustments.
// It does not include nils.
func CollectApprovalRulesetAdjustmentsFromVersions(versions []*ApprovalRulesetVersion) []*ApprovalRulesetAdjustment {
	result := make([]*ApprovalRulesetAdjustment, 0, len(versions))
	for _, elem := range versions {
		if elem.Adjustment != nil {
			result = append(result, elem.Adjustment)
		}
	}
	return result
}
