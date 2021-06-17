package dbmodels

import (
	"reflect"

	"github.com/fullstaq-labs/sqedule/server/dbutils"
	"gorm.io/gorm"
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
	Enabled bool `gorm:"not null; default:true"`

	DisplayName string `gorm:"not null"`
	Description string `gorm:"not null"`
	// TODO: this doesn't work because of the lack of a rule binding mode. move to level of rule binding.
	GloballyApplicable bool `gorm:"not null; default:false"`

	ApprovalRulesetVersion ApprovalRulesetVersion `gorm:"foreignKey:OrganizationID,ApprovalRulesetVersionID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
}

// ApprovalRulesetVersionAndAdjustmentKey uniquely identifies a specific Version+Adjustment
// of an ApprovalRuleset.
type ApprovalRulesetVersionAndAdjustmentKey struct {
	VersionID        uint64
	AdjustmentNumber uint32
}

type ApprovalRulesetWithStats struct {
	ApprovalRuleset
	NumBoundApplications uint
	NumBoundReleases     uint
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

func (ruleset ApprovalRuleset) CheckNewProposalsRequireReview(hasBoundApplications bool) bool {
	return false
	// TODO: comment out after we've implemented the review steps in the version creation process
	//return !hasBoundApplications
}

//
// ******** Find/load functions ********
//

// FindAllApprovalRulesetsWithStats ...
func FindAllApprovalRulesetsWithStats(db *gorm.DB, organizationID string, pagination dbutils.PaginationOptions) ([]ApprovalRulesetWithStats, error) {
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

// FindApprovalRulesetApprovedVersions finds, for a given ApprovalRuleset, all its approved Versions
// and returns them ordered by version number (descending).
func FindApprovalRulesetApprovedVersions(db *gorm.DB, organizationID string, rulesetID string) ([]ApprovalRulesetVersion, error) {
	var result []ApprovalRulesetVersion

	tx := db.Where("organization_id = ? AND approval_ruleset_id = ? AND version_number IS NOT NULL", organizationID, rulesetID).Order("version_number DESC")
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

func CollectApprovalRulesetsWithoutStats(rulesets []ApprovalRulesetWithStats) []*ApprovalRuleset {
	result := make([]*ApprovalRuleset, 0)
	for i := range rulesets {
		ruleset := &rulesets[i]
		result = append(result, &ruleset.ApprovalRuleset)
	}
	return result
}

func CollectApprovalRulesetsWithApplicationApprovalRulesetBindings(bindings []ApplicationApprovalRulesetBinding) []*ApprovalRuleset {
	result := make([]*ApprovalRuleset, 0)
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
