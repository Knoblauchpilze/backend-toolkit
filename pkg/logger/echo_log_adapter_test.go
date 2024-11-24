package logger

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/KnoblauchPilze/backend-toolkit/pkg/errors"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnit_EchoLogAdapter_Duplicate_WhenAdapaterIsCopied_ExpectSuccess(t *testing.T) {
	var out bytes.Buffer
	log := New(&out)
	echoLog := Wrap(log)

	_, err := Duplicate(echoLog)

	assert.Nil(t, err)
}

func TestUnit_EchoLogAdapter_Duplicate_WhenRealLoggerIsCopied_ExpectFailure(t *testing.T) {
	type dummyLogger struct {
		echo.Logger
	}
	log := dummyLogger{}

	_, err := Duplicate(log)

	assert.True(t, errors.IsErrorWithCode(err, UnsupportedLogger), "Actual err: %v", err)
}

func TestUnit_EchoLogAdapter_Duplicate_WriteToTheSameOutput(t *testing.T) {
	var out bytes.Buffer
	log := New(&out)
	echoLog := Wrap(log)

	copy, err := Duplicate(echoLog)
	require.Nil(t, err)

	beforeCall := time.Now()
	copy.Debugf("sample text")

	actual := unmarshalLogOutput(t, out)
	assert.Equal(t, "debug", actual.Level)
	safetyMargin := 5 * time.Second
	assert.True(t, areTimeCloserThan(beforeCall, actual.Time, safetyMargin))
	assert.Equal(t, "sample text", actual.Message)
}

func TestUnit_EchoLogAdapter_Duplicate_ExpectPrefixToBeSeparate(t *testing.T) {
	var out bytes.Buffer
	log := New(&out)
	log.SetPrefix("prefix-1")
	echoLog := Wrap(log)

	copy, err := Duplicate(echoLog)
	copy.SetPrefix("prefix-2")
	require.Nil(t, err)

	log.Infof("log text")
	actual := unmarshalLogOutput(t, out)
	assert.Equal(t, "[prefix-1] log text", actual.Message)
	out = bytes.Buffer{}

	copy.Infof("copy text")
	actual = unmarshalLogOutput(t, out)
	assert.Equal(t, "[prefix-2] copy text", actual.Message)
}

func TestUnit_EchoLogAdapter_Duplicate_ExpectHeaderToBeSeparate(t *testing.T) {
	var out bytes.Buffer
	log := New(&out)
	log.SetHeader("header-1")
	echoLog := Wrap(log)

	copy, err := Duplicate(echoLog)
	copy.SetHeader("header-2")
	require.Nil(t, err)

	log.Infof("log text")
	actual := unmarshalLogOutput(t, out)
	assert.Equal(t, "[header-1] log text", actual.Message)
	out = bytes.Buffer{}

	copy.Infof("copy text")
	actual = unmarshalLogOutput(t, out)
	assert.Equal(t, "[header-2] copy text", actual.Message)
}

type message struct {
	Level   string
	Time    time.Time
	Message string
}

func unmarshalLogOutput(t *testing.T, out bytes.Buffer) message {
	var actual message

	err := json.Unmarshal(out.Bytes(), &actual)
	require.Nil(t, err)

	return actual
}

func areTimeCloserThan(t1 time.Time, t2 time.Time, distance time.Duration) bool {
	diff := t1.Sub(t2).Abs()
	return diff <= distance
}
