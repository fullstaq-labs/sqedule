package dbmodels

import (
	"database/sql"
	"time"

	"github.com/fullstaq-labs/sqedule/dbmodels/reviewstate"
)

// ApprovalRuleset ...
type ApprovalRuleset struct {
	BaseModel
	ID        string    `gorm:"type:citext; primaryKey; not null"`
	CreatedAt time.Time `gorm:"not null"`
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
