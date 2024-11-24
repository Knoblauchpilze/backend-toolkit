package errors

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var errSomeError = fmt.Errorf("some error")

const someCode = ErrorCode(26)

func TestUnit_Error_New(t *testing.T) {
	err := New("haha")

	impl, ok := err.(errorImpl)
	assert.True(t, ok)
	assert.Equal(t, "haha", impl.Message)
	assert.Nil(t, impl.Cause)
	assert.Equal(t, GenericErrorCode, impl.Value)
}

func TestUnit_Error_NewCode(t *testing.T) {
	err := NewCode(someCode)

	impl, ok := err.(errorImpl)
	assert.True(t, ok)
	assert.Equal(t, "An unexpected error occurred", impl.Message)
	assert.Nil(t, impl.Cause)
	assert.Equal(t, someCode, impl.Value)
}

func TestUnit_Error_NewCode_WhenGenericCode_ExpectMessage(t *testing.T) {
	testCases := []struct {
		code            ErrorCode
		expectedMessage string
	}{
		{code: NotImplementedCode, expectedMessage: "Not implemented"},
		{code: GenericErrorCode, expectedMessage: "An unexpected error occurred"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.expectedMessage, func(t *testing.T) {
			err := NewCode(testCase.code)
			impl, ok := err.(errorImpl)
			assert.True(t, ok)
			assert.Equal(t, testCase.expectedMessage, impl.Message)
		})
	}
}

func TestUnit_Error_NewNotImplemented(t *testing.T) {
	err := NotImplemented()

	impl, ok := err.(errorImpl)
	assert.True(t, ok)
	assert.Equal(t, "Not implemented", impl.Message)
	assert.Nil(t, impl.Cause)
	assert.Equal(t, NotImplementedCode, impl.Value)
}

func TestUnit_Error_NewCodeWithDetails(t *testing.T) {
	err := NewCodeWithDetails(someCode, "message")

	impl, ok := err.(errorImpl)
	assert.True(t, ok)
	assert.Equal(t, "message", impl.Message)
	assert.Nil(t, impl.Cause)
	assert.Equal(t, someCode, impl.Value)
}

func TestUnit_Error_Newf(t *testing.T) {
	err := Newf("haha %d", 22)

	impl, ok := err.(errorImpl)
	assert.True(t, ok)
	assert.Equal(t, "haha 22", impl.Message)
	assert.Nil(t, impl.Cause)
}

func TestUnit_Error_Wrap(t *testing.T) {
	err := Wrap(errSomeError, "context")

	impl, ok := err.(errorImpl)
	assert.True(t, ok)
	assert.Equal(t, "context", impl.Message)
	assert.Equal(t, errSomeError, impl.Cause)
}

func TestUnit_Error_WrapCode(t *testing.T) {
	err := WrapCode(errSomeError, someCode)

	impl, ok := err.(errorImpl)
	assert.True(t, ok)
	assert.Equal(t, "An unexpected error occurred", impl.Message)
	assert.Equal(t, errSomeError, impl.Cause)
	assert.Equal(t, someCode, impl.Value)
}

func TestUnit_Error_Wrapf(t *testing.T) {
	err := Wrapf(errSomeError, "context %d", -44)

	impl, ok := err.(errorImpl)
	assert.True(t, ok)
	assert.Equal(t, "context -44", impl.Message)
	assert.Equal(t, errSomeError, impl.Cause)
}

func TestUnit_Error_Unwrap(t *testing.T) {
	err := Unwrap(nil)
	assert.Nil(t, err)

	err = Unwrap(errSomeError)
	assert.Nil(t, err)

	err = New("haha")
	cause := Unwrap(err)
	assert.Nil(t, cause)

	err = Wrap(errSomeError, "haha")
	cause = Unwrap(err)
	assert.Equal(t, errSomeError, cause)

	causeOfCause := Unwrap(cause)
	assert.Nil(t, causeOfCause)
}

func TestUnit_Error_Error(t *testing.T) {
	err := Wrapf(errSomeError, "context %d", -44)

	expected := "context -44. Code: 1 (cause: some error)"
	assert.Equal(t, expected, err.Error())

	err = WrapCode(errSomeError, someCode)

	expected = "An unexpected error occurred. Code: 26 (cause: some error)"
	assert.Equal(t, expected, err.Error())
}

func TestUnit_Error_Code(t *testing.T) {
	err := NewCode(someCode)

	impl, ok := err.(ErrorWithCode)
	assert.True(t, ok)
	assert.Equal(t, someCode, impl.Code())
}

func TestUnit_Error_MarshalJSON(t *testing.T) {
	err := New("haha")
	out, mErr := json.Marshal(err)

	expected := `
	{
		"Code": 1,
		"Message": "haha"
	}`
	assert.Nil(t, mErr)
	assert.JSONEq(t, expected, string(out))

	err = NewCode(someCode)
	out, mErr = json.Marshal(err)

	assert.Nil(t, mErr)
	expected = `
	{
		"Code": 26,
		"Message": "An unexpected error occurred"
	}`
	assert.JSONEq(t, expected, string(out))

	err = Wrap(errSomeError, "hihi")
	out, mErr = json.Marshal(err)

	expected = `
	{
		"Code": 1,
		"Message": "hihi",
		"Cause": "some error"
	}`
	assert.Nil(t, mErr)
	assert.JSONEq(t, expected, string(out))

	err = Wrap(New("haha"), "hihi")
	out, mErr = json.Marshal(err)

	expected = `
	{
		"Code": 1,
		"Message": "hihi",
		"Cause": {
			"Code": 1,
			"Message": "haha"
		}
	}`
	assert.Nil(t, mErr)
	assert.JSONEq(t, expected, string(out))
}

func TestUnit_Error_IsErrorWithCode(t *testing.T) {
	assert.False(t, IsErrorWithCode(nil, someCode))
	assert.False(t, IsErrorWithCode(errSomeError, someCode))
	assert.True(t, IsErrorWithCode(NewCode(someCode), someCode))
	assert.False(t, IsErrorWithCode(NewCode(27), someCode))
	assert.False(t, IsErrorWithCode(New("haha"), someCode))
}
