package rest

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnit_Status_StringerInterface(t *testing.T) {
	tests := []struct {
		name     string
		status   Status
		expected string
	}{
		{
			name:     "StatusSuccess returns SUCCESS",
			status:   StatusSuccess,
			expected: "SUCCESS",
		},
		{
			name:     "StatusError returns ERROR",
			status:   StatusError,
			expected: "ERROR",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.status.String())
		})
	}
}

func TestUnit_Status_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		status   Status
		expected string
	}{
		{
			name:     "StatusSuccess marshals as \"SUCCESS\"",
			status:   StatusSuccess,
			expected: `"SUCCESS"`,
		},
		{
			name:     "StatusError marshals as \"ERROR\"",
			status:   StatusError,
			expected: `"ERROR"`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(tc.status)
			require.Nil(t, err)
			assert.Equal(t, tc.expected, string(data))
		})
	}
}

func TestUnit_Status_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Status
	}{
		{
			name:     "\"SUCCESS\" unmarshals to StatusSuccess",
			input:    `"SUCCESS"`,
			expected: StatusSuccess,
		},
		{
			name:     "\"ERROR\" unmarshals to StatusError",
			input:    `"ERROR"`,
			expected: StatusError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var s Status
			err := json.Unmarshal([]byte(tc.input), &s)
			require.Nil(t, err)
			assert.Equal(t, tc.expected, s)
		})
	}
}

func TestUnit_Status_FailsOnInvalidStatus(t *testing.T) {
	var s Status

	err := json.Unmarshal([]byte(`"INVALID"`), &s)

	require.Error(t, err)
}

func TestUnit_Status_FailsOnInvalidJSON(t *testing.T) {
	var s Status

	err := json.Unmarshal([]byte(`invalid`), &s)

	require.Error(t, err)
}

func TestUnit_Status_FailsOnNonStringType(t *testing.T) {
	var s Status

	err := json.Unmarshal([]byte(`123`), &s)

	require.Error(t, err)
}

func TestUnit_Status_FailsOnNull(t *testing.T) {
	var s Status

	err := json.Unmarshal([]byte(`null`), &s)

	require.Error(t, err)
}

func TestUnit_Status_RoundtripInStruct(t *testing.T) {
	type envelope struct {
		Status Status `json:"status"`
	}

	tests := []struct {
		name   string
		status Status
	}{
		{
			name:   "StatusSuccess roundtrip",
			status: StatusSuccess,
		},
		{
			name:   "StatusError roundtrip",
			status: StatusError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			original := envelope{Status: tc.status}

			data, err := json.Marshal(original)
			require.Nil(t, err)

			var unmarshaled envelope
			err = json.Unmarshal(data, &unmarshaled)
			require.Nil(t, err)

			assert.Equal(t, original.Status, unmarshaled.Status)
		})
	}
}

func TestUnit_Status_ProducesValidJSON(t *testing.T) {
	type envelope struct {
		Status Status `json:"status"`
	}

	e := envelope{Status: StatusSuccess}

	data, err := json.Marshal(e)
	require.Nil(t, err)

	expectedJSON := `{"status":"SUCCESS"}`
	assert.JSONEq(t, expectedJSON, string(data))
}

func TestUnit_Status_CanBeUsedInPrintf(t *testing.T) {
	s := StatusSuccess

	result := s.String()

	assert.Equal(t, "SUCCESS", result)
}

func TestUnit_Status_UnmarshalJSON_FailsOnInvalidCase(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "lowercase fails",
			input: `"success"`,
		},
		{
			name:  "mixed case fails",
			input: `"Success"`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var s Status
			err := json.Unmarshal([]byte(tc.input), &s)
			require.Error(t, err)
		})
	}
}

func TestUnit_Status_SuccessfullyMarshalsArray(t *testing.T) {
	statuses := []Status{StatusSuccess, StatusError}

	data, err := json.Marshal(statuses)

	require.Nil(t, err)
	assert.JSONEq(t, `["SUCCESS","ERROR"]`, string(data))
}

func TestUnit_Status_SuccessfullyUnmarshalsArray(t *testing.T) {
	var statuses []Status

	err := json.Unmarshal([]byte(`["SUCCESS","ERROR"]`), &statuses)

	require.Nil(t, err)
	assert.Equal(t, []Status{StatusSuccess, StatusError}, statuses)
}
