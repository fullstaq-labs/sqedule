package reviewstate

import (
	"bytes"
	"encoding/json"
	"errors"
)

type Input string

const (
	Approved Input = "approved"
	Rejected Input = "rejected"
)

func (s Input) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(string(s))
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

func (s *Input) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		return nil
	}

	var value string
	err := json.Unmarshal(b, &value)
	if err != nil {
		return err
	}

	switch value {
	case "approved":
		*s = Approved
	case "rejected":
		*s = Rejected
	default:
		return errors.New("Invalid review state input")
	}
	return nil
}
