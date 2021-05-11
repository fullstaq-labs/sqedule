package dbmodels

import (
	"github.com/fullstaq-labs/sqedule/server/dbutils"
	"gorm.io/gorm"
)

// User ...
type User struct {
	OrganizationMember
	Email     string `gorm:"type:citext; primaryKey; not null"`
	FirstName string `gorm:"not null"`
	LastName  string `gorm:"not null"`
}

// FindUserByEmail looks up a User by its email address.
// When not found, returns a `gorm.ErrRecordNotFound` error.
func FindUserByEmail(db *gorm.DB, organizationID string, email string) (User, error) {
	var result User

	tx := db.Where("organization_id = ? AND email = ?", organizationID, email)
	tx.Take(&result)
	return result, dbutils.CreateFindOperationError(tx)
}

// Type returns a name of the concrete type. This name is short,
// suitable for machine use, not user display purposes.
func (user User) Type() OrganizationMemberType {
	return UserType
}

// ID returns the primary key's value, i.e. that of `User.Email`
// or `ServiceAccount.Name`.
func (user User) ID() string {
	return user.Email
}

// IDTypeDisplayName returns the primary key's type as a lowercase string suitable
// for user display, i.e. "email" (for User) and "service account name" (for ServiceAccount).
func (user User) IDTypeDisplayName() string {
	return "email"
}
