package middleware

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
)

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
