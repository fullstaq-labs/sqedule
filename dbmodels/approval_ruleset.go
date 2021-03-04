package dbmodels

import (
	"database/sql"
	"reflect"
	"time"

	"github.com/fullstaq-labs/sqedule/dbmodels/reviewstate"
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
