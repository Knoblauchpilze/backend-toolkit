package logger

import (
	"testing"

	"github.com/labstack/gommon/log"
	"github.com/stretchr/testify/assert"
)

func TestUnit_Level_fromEchoLogLevel(t *testing.T) {
	testCases := map[log.Lvl]Level{
		log.DEBUG: DEBUG,
		log.INFO:  INFO,
		log.WARN:  WARN,
		log.ERROR: ERROR,
		log.OFF:   TRACE,
	}

	for in, expected := range testCases {
		t.Run("", func(t *testing.T) {
			actual := fromEchoLogLevel(in)
			assert.Equal(t, expected, actual)
		})
	}
}

func TestUnit_Level_toEchoLogLevel(t *testing.T) {
	testCases := map[Level]log.Lvl{
		DEBUG: log.DEBUG,
		INFO:  log.INFO,
		WARN:  log.WARN,
		ERROR: log.ERROR,
		TRACE: log.OFF,
	}

	for in, expected := range testCases {
		t.Run("", func(t *testing.T) {
			actual := toEchoLogLevel(in)
			assert.Equal(t, expected, actual)
		})
	}
}
