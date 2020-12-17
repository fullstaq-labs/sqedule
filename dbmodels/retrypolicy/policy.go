package retrypolicy

import "database/sql/driver"

// Policy ...
type Policy string

const (
	// Never ...
	Never Policy = "never"
	// RetryOnFail ...
	RetryOnFail Policy = "retry_on_fail"
)

// Scan ...
func (t *Policy) Scan(value interface{}) error {
	*t = Policy(value.(string))
	return nil
}

// Value ...
func (t Policy) Value() (driver.Value, error) {
	return string(t), nil
}
