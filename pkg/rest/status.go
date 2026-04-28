package rest

import (
	"encoding/json"
	"fmt"
)

type Status int

const (
	StatusSuccess Status = iota
	StatusError
)

func (s Status) String() string {
	switch s {
	case StatusSuccess:
		return "SUCCESS"
	case StatusError:
		return "ERROR"
	default:
		return fmt.Sprintf("Status(%d)", s)
	}
}

func (s Status) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
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
