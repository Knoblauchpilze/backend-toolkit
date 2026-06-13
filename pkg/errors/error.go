package errors

import (
	"encoding/json"
	"errors"
	"fmt"
)

type ErrorWithCode struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Cause   error     `json:"cause,omitempty"`
}

func New(message string) error {
	return &ErrorWithCode{
		Code:    GenericErrorCode,
		Message: message,
	}
}

func FromCode(code ErrorCode) error {
	return &ErrorWithCode{
		Code:    code,
		Message: determineCommonErrorMessage(code),
	}
}

func FromCodeAndDetails(code ErrorCode, details string) error {
	return &ErrorWithCode{
		Code:    code,
		Message: details,
	}
}

func Newf(format string, args ...interface{}) error {
	return &ErrorWithCode{
		Code:    GenericErrorCode,
		Message: fmt.Sprintf(format, args...),
	}
}

func Wrap(cause error, message string) error {
	return &ErrorWithCode{
		Code:    GenericErrorCode,
		Message: message,
		Cause:   cause,
	}
}

func WrapCode(cause error, code ErrorCode) error {
	return &ErrorWithCode{
		Code:    code,
		Message: determineCommonErrorMessage(code),
		Cause:   cause,
	}
}

func Wrapf(cause error, format string, args ...interface{}) error {
	return &ErrorWithCode{
		Code:    GenericErrorCode,
		Message: fmt.Sprintf(format, args...),
		Cause:   cause,
	}
}

func Unwrap(err error) error {
	if err == nil {
		return nil
	}

	ie, ok := err.(*ErrorWithCode)
	if !ok {
		return nil
	}

	return ie.Cause
}

func AsErrorWithCode(err error) (*ErrorWithCode, bool) {
	var errWithCode *ErrorWithCode
	if errors.As(err, &errWithCode) {
		return errWithCode, true
	}

	return nil, false
}

func (e *ErrorWithCode) Error() string {
	var out string

	out += e.Message
	out += fmt.Sprintf(". Code: %d", e.Code)

	if e.Cause != nil {
		out += fmt.Sprintf(" (cause: %v)", e.Cause.Error())
	}

	return out
}

func (e *ErrorWithCode) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Code    ErrorCode       `json:"code"`
		Message string          `json:"message,omitempty"`
		Cause   json.RawMessage `json:"cause,omitempty"`
	}{
		Code:    e.Code,
		Message: e.Message,
		Cause:   e.marshalCause(),
	})
}

func (e *ErrorWithCode) marshalCause() json.RawMessage {
	if e.Cause == nil {
		return nil
	}

	var out []byte

	// Voluntarily ignoring the marshalling errors as there's nothing we
	// can do about it.
	if impl, ok := e.Cause.(*ErrorWithCode); ok {
		out, _ = json.Marshal(impl)
	} else {
		out, _ = json.Marshal(e.Cause.Error())
	}

	return out
}

func determineCommonErrorMessage(code ErrorCode) string {
	switch code {
	case errNotImplemented:
		return "not implemented"
	default:
		return "an unexpected error occurred"
	}
}
