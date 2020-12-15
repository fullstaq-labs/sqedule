package approvalpolicy

import "database/sql/driver"

type Policy string

const (
	Any     Policy = "any"
	All     Policy = "all"
	Minimum Policy = "minimum"
)

func (t *Policy) Scan(value interface{}) error {
	*t = Policy(value.([]byte))
	return nil
}

func (t Policy) Value() (driver.Value, error) {
	return string(t), nil
}
