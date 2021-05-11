package dbmodels

import (
	"fmt"
	"time"

	"github.com/fullstaq-labs/sqedule/server/dbmodels/organizationmemberrole"
	"github.com/matthewhartstonge/argon2"
	"gorm.io/gorm"
)

type OrganizationMemberType string

// These values must be short and must not change, because they're used
// in JWT tokens.
const (
	// UserType ...
	UserType OrganizationMemberType = "user"
	// ServiceAccountType ...
	ServiceAccountType OrganizationMemberType = "sa"
)

type OrganizationMember struct {
	BaseModel
	Role         organizationmemberrole.Role `gorm:"type:organization_member_role; not null"`
	PasswordHash string                      `gorm:"not null"`
	CreatedAt    time.Time                   `gorm:"not null"`
	UpdatedAt    time.Time                   `gorm:"not null"`
}

type IOrganizationMember interface {
	IBaseModel

	// Type returns a name of the concrete type. This name is short,
	// suitable for machine use, not user display purposes.
	Type() OrganizationMemberType

	// ID returns the primary key's value, i.e. that of `User.Email`
	// or `ServiceAccount.Name`.
	ID() string

	// IDTypeDisplayName returns the primary key's type as a lowercase string suitable
	// for user display, i.e. "email" (for User) and "service account name" (for ServiceAccount).
	IDTypeDisplayName() string

	// GetRole returns this organization member's role.
	GetRole() organizationmemberrole.Role

	// Authenticate checks whether the given password successfully authenticates
	// this organization member.
	Authenticate(password string) (bool, error)
}

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

func (orgMember OrganizationMember) GetRole() organizationmemberrole.Role {
	return orgMember.Role
}

func (orgMember OrganizationMember) Authenticate(password string) (bool, error) {
	return argon2.VerifyEncoded([]byte(password), []byte(orgMember.PasswordHash))
}
