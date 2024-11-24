package logger

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnit_SafeConsoleWriter_WritesToProvidedWriter(t *testing.T) {
	var out bytes.Buffer

	safeWriter := newSafeConsoleWriter(&out)

	data := []byte("hello")
	actual, err := safeWriter.Write(data)

	assert.Nil(t, err)
	assert.Equal(t, len(data), actual)
}

type mockWriter struct {
	err error
}

func (m *mockWriter) Write(p []byte) (int, error) {
	return 0, m.err
}

func TestUnit_SafeConsoleWriter_WhenWriterFails_ExpectFailure(t *testing.T) {
	m := &mockWriter{
		err: fmt.Errorf("some error"),
	}

	safeWriter := newSafeConsoleWriter(m)

	_, err := safeWriter.Write([]byte{})

	assert.Equal(t, m.err, err)
}
