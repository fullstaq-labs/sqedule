package approvalrulesetbindingmode

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type Mode string

const (
	Permissive Mode = "permissive"
	Enforcing  Mode = "enforcing"
)

func (t *Mode) Scan(value interface{}) error {
	*t = Mode(value.(string))
	return nil
}

func (t Mode) Value() (driver.Value, error) {
	return string(t), nil
}

func (t Mode) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(string(t))
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

func (t *Mode) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		return nil
	}

	var value string
	err := json.Unmarshal(b, &value)
	if err != nil {
		return err
	}

	switch value {
	case "permissive":
		*t = Permissive
	case "enforcing":
		*t = Enforcing
	default:
		return errors.New("Unknown approval ruleset binding mode")
	}
	return nil
}
