package reviewstate

import "database/sql/driver"

// State ...
type State string

const (
	// Draft ...
	Draft State = "draft"
	// Reviewing ...
	Reviewing State = "reviewing"
	// Approved ...
	Approved State = "approved"
	// Rejected ...
	Rejected State = "rejected"
	// Abandoned ...
	Abandoned State = "abandoned"
)

// Scan ...
func (t *State) Scan(value interface{}) error {
	*t = State(value.(string))
	return nil
}

// Value ...
func (t State) Value() (driver.Value, error) {
	return string(t), nil
}
