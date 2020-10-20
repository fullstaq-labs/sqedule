package dbmodels

type ServiceAccount struct {
	OrganizationMember
	Name       string `gorm:"primaryKey; not null"`
	SecretHash string `gorm:"not null"`
}
