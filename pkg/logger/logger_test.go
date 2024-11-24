package logger

import (
	"bytes"
	"regexp"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestUnit_Logger_Printf(t *testing.T) {
	var out bytes.Buffer

	log := New(&out)

	log.Printf("hello %s", "John")

	expectedJson := `
	{
		"level": "debug",
		"message": "hello John"
	}`
	assert.JSONEq(t, expectedJson, out.String())
}

func TestUnit_Logger_Printf_NoArg(t *testing.T) {
	var out bytes.Buffer

	log := New(&out)

	log.Printf("hello")

	expectedJson := `
	{
		"level": "debug",
		"message": "hello"
	}`
	assert.JSONEq(t, expectedJson, out.String())
}

func TestUnit_Logger_WhenUsingPrintfInConsole_ExpectNilTime(t *testing.T) {
	var buf bytes.Buffer
	out := zerolog.ConsoleWriter{
		Out:          &buf,
		TimeLocation: time.UTC,
		TimeFormat:   time.DateTime,
	}

	log := New(out)

	log.Printf("test")

	expectedString := "\x1b[90m<nil>\x1b[0m DBG test\n"
	assert.Equal(t, expectedString, buf.String())
}

func TestUnit_Logger_Tracef(t *testing.T) {
	var out bytes.Buffer

	log := New(&out)
	log.SetLevel(TRACE)

	log.Tracef("hello %s", "John")

	traceRegex := regexp.MustCompile(`{"level":"trace","time":"[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}([0-9:+]+|Z)?","message":"hello John"}`)
	assert.Regexp(t, traceRegex, out.String())
}

func TestUnit_Logger_Debugf(t *testing.T) {
	var out bytes.Buffer

	log := New(&out)

	log.Debugf("hello %s", "John")

	traceRegex := regexp.MustCompile(`{"level":"debug","time":"[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}([0-9:+]+|Z)?","message":"hello John"}`)
	assert.Regexp(t, traceRegex, out.String())
}

func TestUnit_Logger_Infof(t *testing.T) {
	var out bytes.Buffer

	log := New(&out)

	log.Infof("hello %s", "John")

	traceRegex := regexp.MustCompile(`{"level":"info","time":"[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}([0-9:+]+|Z)?","message":"hello John"}`)
	assert.Regexp(t, traceRegex, out.String())
}

func TestUnit_Logger_Warnf(t *testing.T) {
	var out bytes.Buffer

	log := New(&out)

	log.Warnf("hello %s", "John")

	traceRegex := regexp.MustCompile(`{"level":"warn","time":"[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}([0-9:+]+|Z)?","message":"hello John"}`)
	assert.Regexp(t, traceRegex, out.String())
}

func TestUnit_Logger_Errorf(t *testing.T) {
	var out bytes.Buffer

	log := New(&out)

	log.Errorf("hello %s", "John")

	traceRegex := regexp.MustCompile(`{"level":"error","time":"[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}([0-9:+]+|Z)?","message":"hello John"}`)
	assert.Regexp(t, traceRegex, out.String())
}

func TestUnit_Logger_WhenLogLevelDoesNotAllowLogging_ExpectNotLogged(t *testing.T) {
	var out bytes.Buffer

	log := New(&out)
	log.SetLevel(INFO)

	log.Debugf("hello %s", "John")

	assert.Regexp(t, 0, out.Len())
}

func TestUnit_Logger_WhenOnlyPrefixIsSet_ExpectItToBeLogged(t *testing.T) {
	var out bytes.Buffer

	log := New(&out)
	log.SetPrefix("my-prefix")

	log.Infof("hello %s", "John")

	traceRegex := regexp.MustCompile(`{"level":"info","time":"[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}([0-9:+]+|Z)?","message":"\[my-prefix\] hello John"}`)
	assert.Regexp(t, traceRegex, out.String())
}

func TestUnit_Logger_WhenOnlyHeaderIsSet_ExpectItToBeLogged(t *testing.T) {
	var out bytes.Buffer

	log := New(&out)
	log.SetHeader("my-header")

	log.Infof("hello %s", "John")

	traceRegex := regexp.MustCompile(`{"level":"info","time":"[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}([0-9:+]+|Z)?","message":"\[my-header\] hello John"}`)
	assert.Regexp(t, traceRegex, out.String())
}

func TestUnit_Logger_WhenBothPrefixHeaderIsSet_ExpectThemToBeLoggedInTheRightOrder(t *testing.T) {
	var out bytes.Buffer

	log := New(&out)
	log.SetHeader("my-header")
	log.SetPrefix("my-prefix")

	log.Infof("hello %s", "John")

	traceRegex := regexp.MustCompile(`{"level":"info","time":"[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}([0-9:+]+|Z)?","message":"\[my-header\] \[my-prefix\] hello John"}`)
	assert.Regexp(t, traceRegex, out.String())
}

func TestUnit_Logger_WhenSettingNewOutput_ExpectItToBeUsed(t *testing.T) {
	var out1 bytes.Buffer
	log := New(&out1)

	var out2 bytes.Buffer
	log.SetOutput(&out2)

	log.Printf("hello %s", "John")
	expectedJson := `
	{
		"level": "debug",
		"message": "hello John"
	}`
	assert.JSONEq(t, expectedJson, out2.String())
	assert.Equal(t, 0, out1.Len())
}

func TestUnit_Logger_WhenCloning_WritesToSameOutput(t *testing.T) {
	var out bytes.Buffer
	log := New(&out)
	log.SetPrefix("my-prefix")
	log.SetHeader("my-header")
	copy := Clone(log)

	copy.Printf("hello %s", "John")

	expectedJson := `
	{
		"level": "debug",
		"message": "hello John"
	}`
	assert.JSONEq(t, expectedJson, out.String())
	assert.Equal(t, "my-prefix", copy.Prefix())
	assert.Equal(t, "my-header", copy.Header())
}
