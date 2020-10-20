package dbmodels

type BaseModel struct {
	OrganizationID string       `gorm:"primaryKey; not null"`
	Organization   Organization `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}
