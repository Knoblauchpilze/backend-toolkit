package errors

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const someCode = ErrorCode(26)

var errSomeError = fmt.Errorf("some error")

func TestUnit_Error_New(t *testing.T) {
	err := New("foo")

	impl, ok := err.(*ErrorWithCode)
	require.True(t, ok)

	expected := &ErrorWithCode{
		Code:    GenericErrorCode,
		Message: "foo",
		Cause:   nil,
	}
	assert.Equal(t, expected, impl)
}

func TestUnit_Error_FromCode(t *testing.T) {
	t.Run("uses generic message when code is not known", func(t *testing.T) {
		err := FromCode(someCode)

		impl, ok := err.(*ErrorWithCode)
		require.True(t, ok)

		expected := &ErrorWithCode{
			Code:    someCode,
			Message: "an unexpected error occurred",
			Cause:   nil,
		}
		assert.Equal(t, expected, impl)
	})

	t.Run("correctly maps not implemented code", func(t *testing.T) {
		err := FromCode(errNotImplemented)

		impl, ok := err.(*ErrorWithCode)
		require.True(t, ok)

		expected := &ErrorWithCode{
			Code:    errNotImplemented,
			Message: "not implemented",
			Cause:   nil,
		}
		assert.Equal(t, expected, impl)
	})

	t.Run("correctly maps generic code", func(t *testing.T) {
		err := FromCode(GenericErrorCode)

		impl, ok := err.(*ErrorWithCode)
		require.True(t, ok)

		expected := &ErrorWithCode{
			Code:    GenericErrorCode,
			Message: "an unexpected error occurred",
			Cause:   nil,
		}
		assert.Equal(t, expected, impl)
	})
}

func TestUnit_Error_FromCodeAndDetails(t *testing.T) {
	err := FromCodeAndDetails(someCode, "message")

	impl, ok := err.(*ErrorWithCode)
	require.True(t, ok)

	expected := &ErrorWithCode{
		Code:    someCode,
		Message: "message",
		Cause:   nil,
	}
	assert.Equal(t, expected, impl)
}

func TestUnit_Error_Newf(t *testing.T) {
	err := Newf("foo %d", 22)

	impl, ok := err.(*ErrorWithCode)
	require.True(t, ok)

	expected := &ErrorWithCode{
		Code:    GenericErrorCode,
		Message: "foo 22",
		Cause:   nil,
	}
	assert.Equal(t, expected, impl)
}

func TestUnit_Error_Wrap(t *testing.T) {
	err := Wrap(errSomeError, "context")

	impl, ok := err.(*ErrorWithCode)
	require.True(t, ok)

	expected := &ErrorWithCode{
		Code:    GenericErrorCode,
		Message: "context",
		Cause:   errSomeError,
	}
	assert.Equal(t, expected, impl)
}

func TestUnit_Error_WrapCode(t *testing.T) {
	err := WrapCode(errSomeError, someCode)

	impl, ok := err.(*ErrorWithCode)
	require.True(t, ok)

	expected := &ErrorWithCode{
		Code:    someCode,
		Message: "an unexpected error occurred",
		Cause:   errSomeError,
	}
	assert.Equal(t, expected, impl)
}

func TestUnit_Error_Wrapf(t *testing.T) {
	err := Wrapf(errSomeError, "context %d", -44)

	impl, ok := err.(*ErrorWithCode)
	require.True(t, ok)

	expected := &ErrorWithCode{
		Code:    GenericErrorCode,
		Message: "context -44",
		Cause:   errSomeError,
	}
	assert.Equal(t, expected, impl)
}

func TestUnit_Error_Unwrap(t *testing.T) {
	err := Unwrap(nil)
	assert.Nil(t, err)

	err = Unwrap(errSomeError)
	assert.Nil(t, err)

	err = New("foo")
	cause := Unwrap(err)
	assert.Nil(t, cause)

	err = Wrap(errSomeError, "foo")
	cause = Unwrap(err)
	assert.Equal(t, errSomeError, cause)

	causeOfCause := Unwrap(cause)
	assert.Nil(t, causeOfCause)
}

func TestUnit_Error_Error(t *testing.T) {
	t.Run("error returns correct string for error without cause", func(t *testing.T) {
		err := Newf("context %d", -44)

		expected := "context -44. Code: 1"
		assert.Equal(t, expected, err.Error())
	})

	t.Run("error returns correct string for error with cause", func(t *testing.T) {
		err := Wrapf(errSomeError, "context %d", -44)

		expected := "context -44. Code: 1 (cause: some error)"
		assert.Equal(t, expected, err.Error())
	})

	t.Run("error returns correct string when code is not known", func(t *testing.T) {
		err := WrapCode(errSomeError, someCode)

		expected := "an unexpected error occurred. Code: 26 (cause: some error)"
		assert.Equal(t, expected, err.Error())
	})

}

func TestUnit_Error_MarshalJSON(t *testing.T) {
	t.Run("marshals error with code", func(t *testing.T) {
		err := FromCode(someCode)

		out, mErr := json.Marshal(err)
		require.NoError(t, mErr, "Actual err: %v", mErr)

		expected := `
{
	"Code": 26,
	"Message": "An unexpected error occurred"
}`
		assert.JSONEq(t, expected, string(out))
	})

	t.Run("marshals error with details", func(t *testing.T) {
		err := New("foo")

		out, mErr := json.Marshal(err)
		require.NoError(t, mErr, "Actual err: %v", mErr)

		expected := `
{
	"Code": 1,
	"Message": "foo"
}`
		assert.JSONEq(t, expected, string(out))
	})

	t.Run("marshals error with cause", func(t *testing.T) {
		err := Wrap(errSomeError, "hihi")

		out, mErr := json.Marshal(err)
		require.NoError(t, mErr, "Actual err: %v", mErr)

		expected := `
{
	"Code": 1,
	"Message": "hihi",
	"Cause": "some error"
}`
		assert.JSONEq(t, expected, string(out))
	})

	t.Run("marshals error with nested error", func(t *testing.T) {
		err := Wrap(New("foo"), "bar")

		out, mErr := json.Marshal(err)
		require.NoError(t, mErr, "Actual err: %v", mErr)

		expected := `
{
	"Code": 1,
	"Message": "bar",
	"Cause": {
		"Code": 1,
		"Message": "foo"
	}
}`
		assert.JSONEq(t, expected, string(out))
	})
}

func TestUnit_Error_Is(t *testing.T) {
	t.Run("does not detect random error", func(t *testing.T) {
		var err *ErrorWithCode

		testErr := errors.New("test error")
		ok := errors.Is(testErr, err)

		assert.False(t, ok)
		assert.Nil(t, err)
	})

	t.Run("does not detect error with code", func(t *testing.T) {
		var err *ErrorWithCode

		testErr := New("test error")
		ok := errors.Is(testErr, err)

		assert.False(t, ok)
		assert.Nil(t, err)
	})
}

func TestUnit_Error_As(t *testing.T) {
	t.Run("does not detect random error", func(t *testing.T) {
		var err *ErrorWithCode

		testErr := errors.New("test error")
		ok := errors.As(testErr, &err)

		assert.False(t, ok)
		assert.Nil(t, err)
	})

	t.Run("does not detect error with code", func(t *testing.T) {
		var err *ErrorWithCode

		testErr := New("test error")
		ok := errors.As(testErr, &err)

		require.True(t, ok)
		assert.Equal(t, testErr, err)
	})
}
