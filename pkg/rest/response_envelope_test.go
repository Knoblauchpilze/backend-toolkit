package rest

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnit_ResponseEnvelope_MarshalsToCamelCase(t *testing.T) {
	r := responseEnvelope{
		RequestId: "1348f004-7620-4c80-915d-26da0ac144f6",
		Status:    "SUCCESS",
		Details:   json.RawMessage([]byte(`{"Field":32}`)),
	}

	out, err := json.Marshal(r)

	assert.Nil(t, err)
	expectedJson := `
	{
		"requestId": "1348f004-7620-4c80-915d-26da0ac144f6",
		"status": "SUCCESS",
		"details": {
			"Field": 32
		}
	}`
	assert.JSONEq(t, expectedJson, string(out))
}
