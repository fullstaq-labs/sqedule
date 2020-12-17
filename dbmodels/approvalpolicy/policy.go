package approvalpolicy

import "database/sql/driver"

// Policy ...
type Policy string

const (
	// Any ...
	Any Policy = "any"
	// All ...
	All Policy = "all"
	// Minimum ...
	Minimum Policy = "minimum"
)

// Scan ...
func (t *Policy) Scan(value interface{}) error {
	*t = Policy(value.([]byte))
	return nil
}

// Value ...
func (t Policy) Value() (driver.Value, error) {
	return string(t), nil
}
