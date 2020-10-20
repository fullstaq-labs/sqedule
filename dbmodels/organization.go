package dbmodels

type Organization struct {
	ID          string `gorm:"primaryKey; not null"`
	DisplayName string `gorm:"not null"`
}
