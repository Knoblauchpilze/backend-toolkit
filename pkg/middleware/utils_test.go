package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"testing"

	"github.com/KnoblauchPilze/backend-toolkit/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestUnit_FormatHttpStatusCode(t *testing.T) {
	type testCase struct {
		httpStatusCode          int
		expectedFormattedString string
	}

	testCases := []testCase{
		{
			httpStatusCode:          http.StatusOK,
			expectedFormattedString: "\x1b[1;32m200\x1b[0m",
		},
		{
			httpStatusCode:          http.StatusAccepted,
			expectedFormattedString: "\x1b[1;32m202\x1b[0m",
		},
		{
			httpStatusCode:          http.StatusFound,
			expectedFormattedString: "\x1b[1;36m302\x1b[0m",
		},
		{
			httpStatusCode:          http.StatusNotModified,
			expectedFormattedString: "\x1b[1;36m304\x1b[0m",
		},
		{
			httpStatusCode:          http.StatusBadRequest,
			expectedFormattedString: "\x1b[1;33m400\x1b[0m",
		},
		{
			httpStatusCode:          http.StatusForbidden,
			expectedFormattedString: "\x1b[1;33m403\x1b[0m",
		},
		{
			httpStatusCode:          http.StatusInternalServerError,
			expectedFormattedString: "\x1b[1;31m500\x1b[0m",
		},
		{
			httpStatusCode:          http.StatusBadGateway,
			expectedFormattedString: "\x1b[1;31m502\x1b[0m",
		},
	}

	for _, testCase := range testCases {
		t.Run(strconv.Itoa(testCase.httpStatusCode), func(t *testing.T) {
			actual := formatHttpStatusCode(testCase.httpStatusCode)
			assert.Equal(t, testCase.expectedFormattedString, actual)
		})
	}
}

func TestUnit_WrapToHttpError(t *testing.T) {
	err := fmt.Errorf("some error")

	actual := wrapToHttpError(err)

	assertIsHttpErrorWithMessageAndCode(t, actual, "some error", http.StatusInternalServerError)
}

func TestUnit_WrapToHttpError_ErrorWithCode(t *testing.T) {
	err := errors.NewCode(UncaughtPanic)

	actual := wrapToHttpError(err)

	assertIsHttpErrorWithMessageAndCode(t, actual, "An unexpected error occurred. Code: 400", http.StatusInternalServerError)
}

func TestUnit_WrapToHttpError_ErrorWithCodeWithCause(t *testing.T) {
	err := errors.WrapCode(fmt.Errorf("some error"), UncaughtPanic)

	actual := wrapToHttpError(err)

	assertIsHttpErrorWithMessageAndCode(t, actual, "An unexpected error occurred. Code: 400 (cause: some error)", http.StatusInternalServerError)
}
