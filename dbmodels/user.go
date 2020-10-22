package dbmodels

// User ...
type User struct {
	OrganizationMember
	Email        string `gorm:"primaryKey; not null"`
	PasswordHash string `gorm:"not null"`
	FirstName    string `gorm:"not null"`
	LastName     string `gorm:"not null"`
}
