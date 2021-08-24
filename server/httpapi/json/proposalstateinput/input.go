package proposalstateinput

import (
	"bytes"
	"encoding/json"
	"errors"
)

type Input string

const (
	Unset   Input = ""
	Draft   Input = "draft"
	Final   Input = "final"
	Abandon Input = "abandon"
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
	case "draft":
		*s = Draft
	case "final":
		*s = Final
	case "abandon":
		*s = Abandon
	default:
		return errors.New("Unknown proposal state input")
	}
	return nil
}

func (s Input) IsEffectivelyDraft() bool {
	return s == Unset || s == Draft
}
