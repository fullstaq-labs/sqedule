package dbmodels

import (
	"fmt"
	"time"

	"github.com/fullstaq-labs/sqedule/dbmodels/organizationmemberrole"
	"gorm.io/gorm"
)

// OrganizationMemberType ...
type OrganizationMemberType string

// These values must be short and must not change, because they're used
// in JWT tokens.
const (
	// UserType ...
	UserType OrganizationMemberType = "user"
	// ServiceAccountType ...
	ServiceAccountType OrganizationMemberType = "sa"
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
	// Type returns a name of the concrete type. This name is short,
	// suitable for machine use, not user display purposes.
	Type() OrganizationMemberType

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

// FindOrganizationMember ...
func FindOrganizationMember(db *gorm.DB, organizationID string, orgMemberType OrganizationMemberType, orgMemberID string) (IOrganizationMember, error) {
	switch orgMemberType {
	case UserType:
		return FindUserByEmail(db, organizationID, orgMemberID)
	case ServiceAccountType:
		return FindServiceAccountByName(db, organizationID, orgMemberID)
	default:
		panic(fmt.Errorf("Bug: unsupported organization member type %s", orgMemberType))
	}
}
