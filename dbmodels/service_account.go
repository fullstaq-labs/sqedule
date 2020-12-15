package dbmodels

// ServiceAccount ...
type ServiceAccount struct {
	OrganizationMember
	Name       string `gorm:"type:citext; primaryKey; not null"`
	SecretHash string `gorm:"not null"`
}
