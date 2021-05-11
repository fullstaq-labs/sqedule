package dbmodels

import (
	"github.com/fullstaq-labs/sqedule/server/dbutils"
	"gorm.io/gorm"
)

// ServiceAccount ...
type ServiceAccount struct {
	OrganizationMember
	Name string `gorm:"type:citext; primaryKey; not null"`
}

// FindServiceAccountByName looks up a ServiceAccount by its name.
// When not found, returns a `gorm.ErrRecordNotFound` error.
func FindServiceAccountByName(db *gorm.DB, organizationID string, name string) (ServiceAccount, error) {
	var result ServiceAccount

	tx := db.Where("organization_id = ? AND name = ?", organizationID, name)
	tx.Take(&result)
	return result, dbutils.CreateFindOperationError(tx)
}

// Type returns a name of the concrete type. This name is short,
// suitable for machine use, not user display purposes.
func (sa ServiceAccount) Type() OrganizationMemberType {
	return ServiceAccountType
}

// ID returns the primary key's value, i.e. that of `User.Email`
// or `ServiceAccount.Name`.
func (sa ServiceAccount) ID() string {
	return sa.Name
}

// IDTypeDisplayName returns the primary key's type as a lowercase string suitable
// for user display, i.e. "email" (for User) and "service account name" (for ServiceAccount).
func (sa ServiceAccount) IDTypeDisplayName() string {
	return "service account name"
}
