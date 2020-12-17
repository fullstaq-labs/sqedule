package deploymentrequeststate

import "database/sql/driver"

// State ...
type State string

const (
	// InProgress ...
	InProgress State = "in_progress"
	// Cancelled ...
	Cancelled State = "cancelled"
	// Approved ...
	Approved State = "approved"
	// Rejected ...
	Rejected State = "rejected"
)

// Scan ...
func (t *State) Scan(value interface{}) error {
	*t = State(value.([]byte))
	return nil
}

// Value ...
func (t State) Value() (driver.Value, error) {
	return string(t), nil
}
