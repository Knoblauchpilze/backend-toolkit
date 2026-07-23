package rest

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnit_Decoder(t *testing.T) {
	t.Run("decode json or string decodes json", func(t *testing.T) {
		actual, err := DecodeJSONOrString([]byte(`{"value":12}`))
		require.Nil(t, err)

		asMap, ok := actual.(map[string]any)
		require.True(t, ok)

		assert.Equal(t, float64(12), asMap["value"])
	})

	t.Run("decode json or string falls back to string", func(t *testing.T) {
		actual, err := DecodeJSONOrString([]byte("some-data"))
		require.Nil(t, err)

		asString, ok := actual.(string)
		require.True(t, ok)
		assert.Equal(t, "some-data", asString)
	})
}
