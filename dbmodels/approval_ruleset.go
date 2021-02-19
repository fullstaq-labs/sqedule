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
	ApprovalRulesetID string       `gorm:"type:citext; index:approval_ruleset_major_version_idx,sort:desc,where:version_number IS NOT NULL,unique"`
	VersionNumber     *uint32      `gorm:"index:approval_ruleset_major_version_idx,unique"`
	CreatedAt         time.Time    `gorm:"not null"`
	UpdatedAt         time.Time    `gorm:"not null"`

	ApprovalRuleset ApprovalRuleset `gorm:"foreignKey:OrganizationID,ApprovalRulesetID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
}

// ApprovalRulesetMinorVersion ...
type ApprovalRulesetMinorVersion struct {
	BaseModel
	ApprovalRulesetMajorVersionID uint64            `gorm:"primaryKey; not null"`
	VersionNumber                 uint32            `gorm:"primaryKey; not null"`
	ReviewState                   reviewstate.State `gorm:"type:review_state; not null"`
	ReviewComments                sql.NullString
	CreatedAt                     time.Time `gorm:"not null"`
	Enabled                       bool      `gorm:"not null; default:true"`

	DisplayName        string `gorm:"not null"`
	Description        string `gorm:"not null"`
	GloballyApplicable bool   `gorm:"not null; default:false"`

	ApprovalRulesetMajorVersion ApprovalRulesetMajorVersion `gorm:"foreignKey:OrganizationID,ApprovalRulesetMajorVersionID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
}

// GetID ...
func (app ApprovalRuleset) GetID() interface{} {
	return app.ID
}

// SetLatestMajorVersion ...
func (app *ApprovalRuleset) SetLatestMajorVersion(majorVersion IReviewableMajorVersion) {
	app.LatestMajorVersion = majorVersion.(*ApprovalRulesetMajorVersion)
}

// SetLatestMinorVersion ...
func (app *ApprovalRuleset) SetLatestMinorVersion(minorVersion IReviewableMinorVersion) {
	app.LatestMinorVersion = minorVersion.(*ApprovalRulesetMinorVersion)
}

// GetID ...
func (major ApprovalRulesetMajorVersion) GetID() interface{} {
	return major.ID
}

// GetReviewableID ...
func (major ApprovalRulesetMajorVersion) GetReviewableID() interface{} {
	return major.ApprovalRulesetID
}

// AssociateWithReviewable ...
func (major *ApprovalRulesetMajorVersion) AssociateWithReviewable(reviewable IReviewable) {
	ruleset := reviewable.(*ApprovalRuleset)
	major.ApprovalRulesetID = ruleset.ID
	major.ApprovalRuleset = *ruleset
}

// GetMajorVersionID ...
func (minor ApprovalRulesetMinorVersion) GetMajorVersionID() interface{} {
	return minor.ApprovalRulesetMajorVersionID
}

// AssociateWithMajorVersion ...
func (minor *ApprovalRulesetMinorVersion) AssociateWithMajorVersion(majorVersion IReviewableMajorVersion) {
	concreteMajorVersion := majorVersion.(*ApprovalRulesetMajorVersion)
	minor.ApprovalRulesetMajorVersionID = concreteMajorVersion.ID
	minor.ApprovalRulesetMajorVersion = *concreteMajorVersion
}

// LoadApprovalRulesetsLatestVersions ...
func LoadApprovalRulesetsLatestVersions(db *gorm.DB, organizationID string, rulesets []*ApprovalRuleset) error {
	reviewables := make([]IReviewable, 0, len(rulesets))
	for _, ruleset := range rulesets {
		reviewables = append(reviewables, ruleset)
	}

	return LoadReviewablesLatestVersions(
		db,
		reflect.TypeOf(string("")),
		"approval_ruleset_id",
		reflect.TypeOf(ApprovalRulesetMajorVersion{}),
		reflect.TypeOf(uint64(0)),
		"approval_ruleset_major_version_id",
		reflect.TypeOf(ApprovalRulesetMinorVersion{}),
		organizationID,
		reviewables,
	)
}
