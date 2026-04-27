package rest

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type decoderDetails struct {
	Value int `json:"value"`
}

func TestUnit_Decoder_DecodeJSONTo_Succeeds(t *testing.T) {
	actual, err := DecodeJSONTo[decoderDetails]([]byte(`{"value":12}`))
	require.Nil(t, err)

	expected := decoderDetails{Value: 12}
	assert.Equal(t, expected, actual)
}

func TestUnit_Decoder_DecodeJSONTo_FailsOnInvalidJSON(t *testing.T) {
	_, err := DecodeJSONTo[decoderDetails]([]byte(`{"value":`))
	require.Error(t, err)
}

func TestUnit_Decoder_DecodeRawBytes_ReturnsInput(t *testing.T) {
	actual, err := DecodeRawBytes([]byte("some-data"))
	require.Nil(t, err)

	assert.Equal(t, []byte("some-data"), actual)
}

func TestUnit_Decoder_DecodeString_ReturnsInputAsString(t *testing.T) {
	actual, err := DecodeString([]byte("some-data"))
	require.Nil(t, err)

	assert.Equal(t, "some-data", actual)
}

func TestUnit_Decoder_DecodeJSONOrString_DecodesJSON(t *testing.T) {
	actual, err := DecodeJSONOrString([]byte(`{"value":12}`))
	require.Nil(t, err)

	asMap, ok := actual.(map[string]any)
	require.True(t, ok)

	assert.Equal(t, float64(12), asMap["value"])
}

func TestUnit_Decoder_DecodeJSONOrString_FallbacksToString(t *testing.T) {
	actual, err := DecodeJSONOrString([]byte("some-data"))
	require.Nil(t, err)

	asString, ok := actual.(string)
	require.True(t, ok)
	assert.Equal(t, "some-data", asString)
}
