package logger

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnit_SafeConsoleWriter(t *testing.T) {
	t.Run("writes to provided buffer", func(t *testing.T) {
		var out bytes.Buffer

		safeWriter := newSafeConsoleWriter(&out)

		data := []byte("hello")
		actual, err := safeWriter.Write(data)
		require.NoError(t, err, "Actual err: %v", err)

		assert.Equal(t, len(data), actual)
	})

	t.Run("returns error when writer fails", func(t *testing.T) {
		m := &mockWriter{
			err: errors.New("some error"),
		}

		safeWriter := newSafeConsoleWriter(m)

		_, err := safeWriter.Write([]byte{})

		assert.ErrorIs(t, err, m.err, "Actual err: %v", err)
	})
}

type mockWriter struct {
	err error
}

func (m *mockWriter) Write(p []byte) (int, error) {
	return 0, m.err
}
