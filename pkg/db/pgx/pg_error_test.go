package pgx

import (
	"fmt"
	"testing"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnit_AnalyzeAndWrapPgError_Nil(t *testing.T) {
	err := AnalyzeAndWrapPgError(nil)

	assert.Nil(t, err)
}

func TestUnit_AnalyzeAndWrapPgError_WhenNotAKnownError_ExpectUnchanged(t *testing.T) {
	err := fmt.Errorf("some error")

	actual := AnalyzeAndWrapPgError(err)

	assert.Equal(t, err, actual)
}

func TestUnit_AnalyzeAndWrapPgError_PgError(t *testing.T) {
	type testCase struct {
		code          string
		expectedError errors.ErrorCode
	}

	testCases := []testCase{
		{
			code:          "23503",
			expectedError: ErrForeignKeyValidation,
		},
		{
			code:          "23505",
			expectedError: ErrUniqueConstraintViolation,
		},
		{
			code:          "not-a-code",
			expectedError: ErrGenericSqlError,
		},
	}

	for _, testCase := range testCases {
		t.Run("", func(t *testing.T) {
			err := &pgconn.PgError{
				Code: testCase.code,
			}

			rawErr := AnalyzeAndWrapPgError(err)

			actual, ok := errors.AsErrorWithCode(rawErr)
			require.True(t, ok)

			expected := &errors.ErrorWithCode{
				Code:    testCase.expectedError,
				Message: "an unexpected error occurred",
				Cause:   err,
			}
			assert.Equal(t, expected, actual)
		})
	}
}
