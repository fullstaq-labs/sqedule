package organizationmemberrole

import "database/sql/driver"

type Role string

const (
	Owner         Role = "owner"
	Admin         Role = "admin"
	ChangeManager Role = "change_manager"
	Technician    Role = "technician"
	Viewer        Role = "viewer"
)

func (t *Role) Scan(value interface{}) error {
	*t = Role(value.([]byte))
	return nil
}

func (t Role) Value() (driver.Value, error) {
	return string(t), nil
}
