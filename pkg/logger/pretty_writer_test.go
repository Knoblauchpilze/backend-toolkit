package logger

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnit_NewPrettyWriter(t *testing.T) {
	var out bytes.Buffer

	sampleText := `
	{
		"time": "2024-11-15T21:54:53+01:00",
		"level": "info",
		"key": 1,
		"name": "John",
		"greeting": "hello"
	}`

	w := NewPrettyWriter(&out)
	n, err := w.Write([]byte(sampleText))

	assert.Nil(t, err)
	assert.Equal(t, 116, n)
	expectedOutput := "\x1b[90m2024-11-15 20:54:53\x1b[0m \x1b[32mINF\x1b[0m \x1b[36mgreeting=\x1b[0mhello \x1b[36mkey=\x1b[0m1 \x1b[36mname=\x1b[0mJohn\n"
	assert.Equal(t, expectedOutput, out.String())
}

func TestUnit_NewPrettyWriter_WhenTimeNotSet_ExpectNil(t *testing.T) {
	var out bytes.Buffer

	sampleText := `
	{
		"level": "info",
		"key": 1,
		"name": "John",
		"greeting": "hello"
	}`

	w := NewPrettyWriter(&out)
	n, err := w.Write([]byte(sampleText))

	assert.Nil(t, err)
	assert.Equal(t, 77, n)
	expectedOutput := "\x1b[90m<nil>\x1b[0m \x1b[32mINF\x1b[0m \x1b[36mgreeting=\x1b[0mhello \x1b[36mkey=\x1b[0m1 \x1b[36mname=\x1b[0mJohn\n"
	assert.Equal(t, expectedOutput, out.String())
}

func TestUnit_NewPrettyWriter_WhenLevelNotSet_ExpectQuestionMarks(t *testing.T) {
	var out bytes.Buffer

	sampleText := `
	{
		"time": "2024-11-15T21:54:53+01:00",
		"key": 1,
		"name": "John",
		"greeting": "hello"
	}`

	w := NewPrettyWriter(&out)
	n, err := w.Write([]byte(sampleText))

	assert.Nil(t, err)
	assert.Equal(t, 97, n)
	expectedOutput := "\x1b[90m2024-11-15 20:54:53\x1b[0m ??? \x1b[36mgreeting=\x1b[0mhello \x1b[36mkey=\x1b[0m1 \x1b[36mname=\x1b[0mJohn\n"
	assert.Equal(t, expectedOutput, out.String())
}
