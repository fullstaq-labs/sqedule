package approvalrulesetbindingmode

import "database/sql/driver"

type Mode string

const (
	Permissive Mode = "permissive"
	Enforcing  Mode = "enforcing"
)

func (t *Mode) Scan(value interface{}) error {
	*t = Mode(value.([]byte))
	return nil
}

func (t Mode) Value() (driver.Value, error) {
	return string(t), nil
}
