package dbmodels

//
// ******** Types, constants & variables ********
//

type BaseModel struct {
	OrganizationID string       `gorm:"type:citext; primaryKey; not null"`
	Organization   Organization `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

type IBaseModel interface {
	GetOrganizationID() string
}

//
// ******** BaseModel methods ********
//

func (m BaseModel) GetOrganizationID() string {
	return m.OrganizationID
}
