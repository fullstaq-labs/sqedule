package proposalstate

import "database/sql/driver"

type State string

const (
	Draft     State = "draft"
	Reviewing State = "reviewing"
	Approved  State = "approved"
	Rejected  State = "rejected"
	Abandoned State = "abandoned"
)

func (t *State) Scan(value interface{}) error {
	*t = State(value.(string))
	return nil
}

func (t State) Value() (driver.Value, error) {
	return string(t), nil
}
