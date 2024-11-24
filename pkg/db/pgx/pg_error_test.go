package pgx

import (
	"fmt"
	"testing"

	"github.com/KnoblauchPilze/backend-toolkit/pkg/errors"
	jpgx "github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
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
			expectedError: ForeignKeyValidation,
		},
		{
			code:          "23505",
			expectedError: UniqueConstraintViolation,
		},
		{
			code:          "not-a-code",
			expectedError: GenericSqlError,
		},
	}

	for _, testCase := range testCases {
		t.Run("", func(t *testing.T) {
			err := &jpgx.PgError{
				Code: testCase.code,
			}

			actual := AnalyzeAndWrapPgError(err)

			assert.True(t, errors.IsErrorWithCode(actual, testCase.expectedError), "Actual err: %v", err)
			cause := errors.Unwrap(actual)
			assert.Equal(t, err, cause)
		})
	}
}
