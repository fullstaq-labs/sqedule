package dbmodels

// User ...
type User struct {
	OrganizationMember
	Email        string `gorm:"type:citext; primaryKey; not null"`
	PasswordHash string `gorm:"not null"`
	FirstName    string `gorm:"not null"`
	LastName     string `gorm:"not null"`
}
