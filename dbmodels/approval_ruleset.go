package dbmodels

import (
	"database/sql"
	"time"

	"github.com/fullstaq-labs/sqedule/dbmodels/reviewstate"
)

type ApprovalRuleset struct {
	BaseModel
	ID        string    `gorm:"primaryKey; not null"`
	CreatedAt time.Time `gorm:"not null"`
}

type ApprovalRulesetMajorVersion struct {
	OrganizationID    string    `gorm:"primaryKey; not null; index:version,unique"`
	ID                uint64    `gorm:"primaryKey; not null"`
	ApprovalRulesetID string    `gorm:"index:version,unique"`
	VersionNumber     *uint32   `gorm:"index:version,unique"`
	CreatedAt         time.Time `gorm:"not null"`
	UpdatedAt         time.Time `gorm:"not null"`

	ApprovalRuleset ApprovalRuleset `gorm:"foreignKey:OrganizationID,ApprovalRulesetID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
}

type ApprovalRulesetMinorVersion struct {
	BaseModel
	ApprovalRulesetMajorVersionID uint64            `gorm:"primaryKey; not null"`
	VersionNumber                 uint32            `gorm:"primaryKey; not null"`
	ReviewState                   reviewstate.State `gorm:"type:review_state; not null"`
	ReviewComments                sql.NullString
	CreatedAt                     time.Time `gorm:"not null"`
	Enabled                       bool      `gorm:"not null"`

	DisplayName        string `gorm:"not null"`
	Description        string `gorm:"not null"`
	GloballyApplicable string `gorm:"not null"` // TODO: index this

	ApprovalRulesetMajorVersion ApprovalRulesetMajorVersion `gorm:"foreignKey:OrganizationID,ApprovalRulesetMajorVersionID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
}
