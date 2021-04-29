package organizationmemberrole

import "database/sql/driver"

// Role ...
type Role string

const (
	// OrgAdmin ...
	OrgAdmin Role = "org_admin"
	// Admin ...
	Admin Role = "admin"
	// ChangeManager ...
	ChangeManager Role = "change_manager"
	// Technician ...
	Technician Role = "technician"
	// Viewer ...
	Viewer Role = "viewer"
)

// Scan ...
func (t *Role) Scan(value interface{}) error {
	*t = Role(value.(string))
	return nil
}

// Value ...
func (t Role) Value() (driver.Value, error) {
	return string(t), nil
}
