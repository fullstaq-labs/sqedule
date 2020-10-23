package dbmodels

import (
	"time"

	"github.com/fullstaq-labs/sqedule/dbmodels/organizationmemberrole"
)

// OrganizationMember ...
type OrganizationMember struct {
	BaseModel
	Role      organizationmemberrole.Role `gorm:"type:organization_member_role; not null"`
	CreatedAt time.Time                   `gorm:"not null"`
	UpdatedAt time.Time                   `gorm:"not null"`
}
