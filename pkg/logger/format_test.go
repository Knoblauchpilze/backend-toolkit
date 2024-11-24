package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnit_Format(t *testing.T) {
	type testCase struct {
		color        Color
		expectedText string
	}

	testCases := map[string]testCase{
		"blue": {

			color:        Blue,
			expectedText: "\033[1;34mhello\033[0m",
		},
		"cyan": {
			color:        Cyan,
			expectedText: "\033[1;36mhello\033[0m",
		},
		"gray": {
			color:        Gray,
			expectedText: "\033[1;90mhello\033[0m",
		},
		"green": {
			color:        Green,
			expectedText: "\033[1;32mhello\033[0m",
		},
		"magenta": {
			color:        Magenta,
			expectedText: "\033[1;35mhello\033[0m",
		},
		"yellow": {
			color:        Yellow,
			expectedText: "\033[1;33mhello\033[0m",
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			actual := FormatWithColor("hello", testCase.color)

			assert.Equal(t, testCase.expectedText, actual)
		})
	}
}
