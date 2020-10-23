package dbmodels

// Organization ...
type Organization struct {
	ID          string `gorm:"type: citext; primaryKey; not null"`
	DisplayName string `gorm:"not null"`
}
