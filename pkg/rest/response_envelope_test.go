package rest

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnit_ResponseEnvelope_SuccessfullyMarshalsTypedDetails(t *testing.T) {
	type details struct {
		Field int `json:"field"`
	}

	r := ResponseEnvelope[details]{
		RequestId: "1348f004-7620-4c80-915d-26da0ac144f6",
		Status:    StatusSuccess,
		Details:   details{Field: 32},
	}

	out, err := json.Marshal(r)

	assert.Nil(t, err)
	expectedJson := `
	{
		"requestId": "1348f004-7620-4c80-915d-26da0ac144f6",
		"status": "SUCCESS",
		"details": {
			"field": 32
		}
	}`
	assert.JSONEq(t, expectedJson, string(out))
}

func TestUnit_ResponseEnvelope_SuccessfullyMarshalsSimpleDetails(t *testing.T) {
	r := ResponseEnvelope[int32]{
		RequestId: "1348f004-7620-4c80-915d-26da0ac144f6",
		Status:    StatusSuccess,
		Details:   int32(16),
	}

	out, err := json.Marshal(r)

	assert.Nil(t, err)
	expectedJson := `
	{
		"requestId": "1348f004-7620-4c80-915d-26da0ac144f6",
		"status": "SUCCESS",
		"details": 16
	}`
	assert.JSONEq(t, expectedJson, string(out))
}
