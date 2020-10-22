package dbmodels

import "database/sql"

// User ...
type User struct {
	OrganizationMember
	Email        string `gorm:"primaryKey; not null"`
	PasswordHash string `gorm:"not null"`
	FirstName    sql.NullString
	LastName     sql.NullString
}
