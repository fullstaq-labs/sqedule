package retrypolicy

import "database/sql/driver"

type Policy string

const (
	Never       Policy = "never"
	RetryOnFail Policy = "retry_on_fail"
)

func (t *Policy) Scan(value interface{}) error {
	*t = Policy(value.([]byte))
	return nil
}

func (t Policy) Value() (driver.Value, error) {
	return string(t), nil
}
