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

// IOrganizationMember ...
type IOrganizationMember interface {
	// TypeShortName returns a name of the concrete type. This name is short,
	// suitable for machine use, not user display purposes.
	TypeShortName() string

	// ID returns the primary key's value, i.e. that of `User.Email`
	// or `ServiceAccount.Name`.
	ID() string

	// IDTypeDisplayName returns the primary key's type as a lowercase string suitable
	// for user display, i.e. "email" (for User) and "service account name" (for ServiceAccount).
	IDTypeDisplayName() string

	// OrganizationMember returns the OrganizationMember embedded in this object.
	GetOrganizationMember() *OrganizationMember

	// Authenticate checks whether the given access token successfully authenticates
	// this organization member.
	Authenticate(accessToken string) (bool, error)
}
