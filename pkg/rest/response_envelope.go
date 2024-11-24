package rest

import (
	"encoding/json"
)

type responseEnvelope struct {
	RequestId string          `json:"requestId"`
	Status    string          `json:"status"`
	Details   json.RawMessage `json:"details,omitempty"`
}
