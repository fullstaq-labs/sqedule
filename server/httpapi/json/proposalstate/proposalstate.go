package proposalstate

import (
	"bytes"
	"encoding/json"
	"errors"
)

type ProposalState string

const (
	Unset   ProposalState = ""
	Draft   ProposalState = "draft"
	Final   ProposalState = "final"
	Abandon ProposalState = "abandon"
)

func (s ProposalState) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(string(s))
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

func (s *ProposalState) UnmarshalJSON(b []byte) error {
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
		return errors.New("Unknown proposal state")
	}
	return nil
}

func (s ProposalState) IsEffectivelyDraft() bool {
	return s == Unset || s == Draft
}
