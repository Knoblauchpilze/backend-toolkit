package rest

import (
	"encoding/json"
	"fmt"
)

type Status string

const (
	StatusSuccess Status = "SUCCESS"
	StatusError   Status = "ERROR"
)

func (s Status) String() string {
	if s == StatusSuccess || s == StatusError {
		return string(s)
	}

	return fmt.Sprintf("Status(%q)", string(s))
}

func (s Status) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(s))
}

func (s *Status) UnmarshalJSON(data []byte) error {
	var str string
	err := json.Unmarshal(data, &str)
	if err != nil {
		return err
	}

	switch str {
	case "SUCCESS":
		*s = StatusSuccess
	case "ERROR":
		*s = StatusError
	default:
		return fmt.Errorf("invalid status: %q", str)
	}

	return nil
}
