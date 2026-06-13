package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	berrors "github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errSomeError = fmt.Errorf("some error")

func TestUnit_Error_AsDatabaseError(t *testing.T) {
	t.Run("does not detect random error", func(t *testing.T) {
		testErr := errors.New("test error")
		actual, ok := AsDatabaseError(testErr)

		assert.False(t, ok)
		assert.Nil(t, actual)
	})

	t.Run("detects databse error", func(t *testing.T) {
		testErr := &DatabaseError{SqlCode: "23503"}
		actual, ok := AsDatabaseError(testErr)

		require.True(t, ok)
		assert.Equal(t, testErr, actual)
	})
}

func TestUnit_Error_Error(t *testing.T) {
	t.Run("returns correct string for error without cause", func(t *testing.T) {
		err := DatabaseError{
			Code:       errAlreadyCommitted,
			Message:    "context",
			SqlCode:    "44",
			Schema:     "schema",
			Table:      "table",
			Column:     "column",
			Constraint: "constraint",
			Cause:      nil,
		}

		expected := "context, code: 102, sql code: 44"
		assert.Equal(t, expected, err.Error())
	})

	t.Run("returns correct string for error with cause", func(t *testing.T) {
		err := DatabaseError{
			Code:       errUnsupportedOperation,
			Message:    "context",
			SqlCode:    "44",
			Schema:     "schema",
			Table:      "table",
			Column:     "column",
			Constraint: "constraint",
			Cause:      errSomeError,
		}

		expected := "context, code: 101, sql code: 44 (cause: some error)"
		assert.Equal(t, expected, err.Error())
	})

}

func TestUnit_Error_MarshalJSON(t *testing.T) {
	t.Run("marshals error with no cause", func(t *testing.T) {
		err := &DatabaseError{
			Code:       errTooManyMatchingRows,
			Message:    "context",
			SqlCode:    "44",
			Schema:     "the-schema",
			Table:      "the-table",
			Column:     "the-column",
			Constraint: "the-constraint",
			Cause:      nil,
		}

		out, mErr := json.Marshal(err)
		require.NoError(t, mErr, "Actual err: %v", mErr)

		expected := `
{
	"code": 111,
	"sql_code": "44",
	"message": "context",
	"schema": "the-schema",
	"table": "the-table",
	"column": "the-column",
	"constraint": "the-constraint"
}`
		assert.JSONEq(t, expected, string(out))
	})

	t.Run("marshals error with cause", func(t *testing.T) {
		err := &DatabaseError{
			Code:       errNoMatchingRows,
			Message:    "context",
			SqlCode:    "44",
			Schema:     "the-schema",
			Table:      "the-table",
			Column:     "the-column",
			Constraint: "the-constraint",
			Cause:      errSomeError,
		}

		out, mErr := json.Marshal(err)
		require.NoError(t, mErr, "Actual err: %v", mErr)

		expected := `
{
	"code": 110,
	"sql_code": "44",
	"message": "context",
	"schema": "the-schema",
	"table": "the-table",
	"column": "the-column",
	"constraint": "the-constraint",
	"cause": "some error"
}`
		assert.JSONEq(t, expected, string(out))
	})

	t.Run("marshals error with nested error with code", func(t *testing.T) {
		err := &DatabaseError{
			Code:       errNotConnected,
			Message:    "context",
			SqlCode:    "44",
			Schema:     "the-schema",
			Table:      "the-table",
			Column:     "the-column",
			Constraint: "the-constraint",
			Cause:      berrors.New("foo"),
		}

		out, mErr := json.Marshal(err)
		require.NoError(t, mErr, "Actual err: %v", mErr)

		expected := `
{
	"code": 100,
	"sql_code": "44",
	"message": "context",
	"schema": "the-schema",
	"table": "the-table",
	"column": "the-column",
	"constraint": "the-constraint",
	"cause": {
		"code": 1,
		"message": "foo"
	}
}`
		assert.JSONEq(t, expected, string(out))
	})
}

func TestUnit_Error_Is(t *testing.T) {
	t.Run("does not detect random error", func(t *testing.T) {
		var err *DatabaseError

		testErr := errors.New("test error")
		ok := errors.Is(testErr, err)

		assert.False(t, ok)
		assert.Nil(t, err)
	})

	t.Run("does not detect database error", func(t *testing.T) {
		var err *DatabaseError

		testErr := &DatabaseError{SqlCode: "23503"}
		ok := errors.Is(testErr, err)

		assert.False(t, ok)
		assert.Nil(t, err)
	})
}

func TestUnit_Error_As(t *testing.T) {
	t.Run("does not detect random error", func(t *testing.T) {
		var err *DatabaseError

		testErr := errors.New("test error")
		ok := errors.As(testErr, &err)

		assert.False(t, ok)
		assert.Nil(t, err)
	})

	t.Run("detects database error", func(t *testing.T) {
		var err *DatabaseError

		testErr := &DatabaseError{SqlCode: "23503"}
		ok := errors.As(testErr, &err)

		require.True(t, ok)
		assert.Equal(t, testErr, err)
	})
}
