package approvalrulesetbindingmode

import "database/sql/driver"

// Mode ...
type Mode string

const (
	// Permissive ...
	Permissive Mode = "permissive"
	// Enforcing ...
	Enforcing Mode = "enforcing"
)

// Scan ...
func (t *Mode) Scan(value interface{}) error {
	*t = Mode(value.(string))
	return nil
}

// Value ...
func (t Mode) Value() (driver.Value, error) {
	return string(t), nil
}
