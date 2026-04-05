package logger

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestUnit_Logger_UsesProvidedOutput(t *testing.T) {
	var out bytes.Buffer

	log := New(&out)

	log.Debug("hello", slog.String("name", "John"))

	// Note: the additional character are due to the colored output
	assert.Regexp(t, ".*[0-9-]+ [0-9:]+.* DBG hello .*name=.*John\n", out.String())
}

func TestUnit_Logger_DoesNotLogWhenSeverityIsHigherThanMessage(t *testing.T) {
	var out bytes.Buffer

	log := NewWithLevel(&out, zerolog.WarnLevel)

	log.Info("hello", slog.String("name", "John"))

	assert.Empty(t, out)
}
