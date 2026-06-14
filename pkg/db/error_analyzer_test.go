package db

import (
	"errors"
	"testing"

	berrors "github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnit_AnalyzeAndWrapDatabaseError(t *testing.T) {
	t.Run("does not wrap when error is nil", func(t *testing.T) {
		err := analyzeAndWrapDatabaseError(nil)
		assert.Nil(t, err)
	})

	t.Run("does not wrap error when not a PgError", func(t *testing.T) {
		err := errors.New("some error")

		actual := analyzeAndWrapDatabaseError(err)

		assert.Equal(t, err, actual)
	})
}

func TestUnit_AnalyzeAndWrapDatabaseError_PgError(t *testing.T) {
	type testCase struct {
		name          string
		code          string
		expectedError berrors.ErrorCode
	}

	testCases := []testCase{
		{
			name:          "foreign key validation",
			code:          "23503",
			expectedError: ErrForeignKeyValidation,
		},
		{
			name:          "unique constraint violation",
			code:          "23505",
			expectedError: ErrUniqueConstraintViolation,
		},
		{
			name:          "generic error",
			code:          "not-a-code",
			expectedError: ErrGenericSqlError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := &pgconn.PgError{
				Code: testCase.code,
			}

			rawErr := analyzeAndWrapDatabaseError(err)

			actual, ok := AsDatabaseError(rawErr)
			require.True(t, ok)

			expected := &DatabaseError{
				Code:    testCase.expectedError,
				SqlCode: testCase.code,
				Cause:   err,
			}
			assert.Equal(t, expected, actual)
		})
	}
}
