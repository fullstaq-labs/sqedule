package dbmodels

type BaseModel struct {
	OrganizationID string       `gorm:"type:citext; primaryKey; not null"`
	Organization   Organization `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

type IBaseModel interface {
	GetOrganizationID() string
}

func (m BaseModel) GetOrganizationID() string {
	return m.OrganizationID
}
