package deploymentrequeststate

import "database/sql/driver"

type State string

const (
	InProgress State = "in_progress"
	Cancelled  State = "cancelled"
	Approved   State = "approved"
	Rejected   State = "rejected"
)

func (t *State) Scan(value interface{}) error {
	*t = State(value.([]byte))
	return nil
}

func (t State) Value() (driver.Value, error) {
	return string(t), nil
}
