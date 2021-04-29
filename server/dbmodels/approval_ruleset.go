package dbmodels

import (
	"database/sql"
	"reflect"
	"time"

	"github.com/fullstaq-labs/sqedule/server/dbmodels/reviewstate"
	"github.com/fullstaq-labs/sqedule/server/dbutils"
	"gorm.io/gorm"
)

// ApprovalRuleset ...
type ApprovalRuleset struct {
	BaseModel
	ID                 string                       `gorm:"type:citext; primaryKey; not null"`
	CreatedAt          time.Time                    `gorm:"not null"`
	LatestMajorVersion *ApprovalRulesetMajorVersion `gorm:"-"`
	LatestMinorVersion *ApprovalRulesetMinorVersion `gorm:"-"`
}

// ApprovalRulesetMajorVersion ...
type ApprovalRulesetMajorVersion struct {
	OrganizationID    string       `gorm:"type:citext; primaryKey; not null; index:approval_ruleset_major_version_idx,unique"`
	Organization      Organization `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	ID                uint64       `gorm:"primaryKey; autoIncrement; not null"`
	ApprovalRulesetID string       `gorm:"type:citext; not null; index:approval_ruleset_major_version_idx,unique"`
	VersionNumber     *uint32      `gorm:"type:int; index:approval_ruleset_major_version_idx,sort:desc,where:version_number IS NOT NULL,unique; check:(version_number > 0)"`
	CreatedAt         time.Time    `gorm:"not null"`
	UpdatedAt         time.Time    `gorm:"not null"`

	ApprovalRuleset ApprovalRuleset `gorm:"foreignKey:OrganizationID,ApprovalRulesetID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
}

// ApprovalRulesetMinorVersion ...
type ApprovalRulesetMinorVersion struct {
	BaseModel
	ApprovalRulesetMajorVersionID uint64            `gorm:"primaryKey; not null"`
	VersionNumber                 uint32            `gorm:"type:int; primaryKey; not null; check:(version_number > 0)"`
	ReviewState                   reviewstate.State `gorm:"type:review_state; not null"`
	ReviewComments                sql.NullString
	CreatedAt                     time.Time `gorm:"not null"`
	Enabled                       bool      `gorm:"not null; default:true"`

	DisplayName string `gorm:"not null"`
	Description string `gorm:"not null"`
	// TODO: this doesn't work because of the lack of a rule binding mode. move to level of rule binding.
	GloballyApplicable bool `gorm:"not null; default:false"`

	ApprovalRulesetMajorVersion ApprovalRulesetMajorVersion `gorm:"foreignKey:OrganizationID,ApprovalRulesetMajorVersionID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
}

type ApprovalRulesetVersionKey struct {
	MajorVersionID     uint64
	MinorVersionNumber uint32
}

type ApprovalRulesetWithStats struct {
	ApprovalRuleset
	NumBoundApplications uint
	NumBoundReleases     uint
}

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

// LoadApprovalRulesetsLatestVersions ...
func LoadApprovalRulesetsLatestVersions(db *gorm.DB, organizationID string, rulesets []*ApprovalRuleset) error {
	reviewables := make([]IReviewable, 0, len(rulesets))
	for _, ruleset := range rulesets {
		reviewables = append(reviewables, ruleset)
	}

	return LoadReviewablesLatestVersions(
		db,
		reflect.TypeOf(ApprovalRuleset{}.ID),
		[]string{"approval_ruleset_id"},
		reflect.TypeOf(ApprovalRuleset{}.ID),
		reflect.TypeOf(ApprovalRulesetMajorVersion{}),
		reflect.TypeOf(ApprovalRulesetMajorVersion{}.ID),
		"approval_ruleset_major_version_id",
		reflect.TypeOf(ApprovalRulesetMinorVersion{}),
		organizationID,
		reviewables,
	)
}
