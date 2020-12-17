package organizationmemberrole

import "database/sql/driver"

// Role ...
type Role string

const (
	// Owner ...
	Owner Role = "owner"
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
	*t = Role(value.([]byte))
	return nil
}

// Value ...
func (t Role) Value() (driver.Value, error) {
	return string(t), nil
}
