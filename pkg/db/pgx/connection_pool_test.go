package pgx

import (
	"context"
	"testing"

	"github.com/KnoblauchPilze/backend-toolkit/pkg/errors"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnit_New_InvalidConnectionString(t *testing.T) {
	pool, err := New(context.Background(), "invalid-connection-string")

	assert.Nil(t, pool)
	assert.NotNil(t, err)
	_, ok := err.(*pgconn.ParseConfigError)
	assert.True(t, ok)
}

func TestUnit_New_ValidConnectionString(t *testing.T) {
	const connStr = "postgres://user:password@localhost/my-db"
	pool, err := New(context.Background(), connStr)

	assert.NotNil(t, pool)
	assert.Nil(t, err)
}

func TestIT_New_ConnectsToDatabase(t *testing.T) {
	const connStr = "postgres://test_user:test_password@localhost:5432/test_db"
	pool, err := New(context.Background(), connStr)
	require.Nil(t, err)

	err = pool.Ping(context.Background())
	assert.Nil(t, err)
}

func TestIT_New_ConnectsToDatabase_WrongCredentials(t *testing.T) {
	type testCase struct {
		connStr       string
		expectedError errors.ErrorCode
	}

	testCases := []testCase{
		{
			connStr:       "postgres://test_user:comes-from-the-environment@localhost:5432/test_db",
			expectedError: AuthenticationFailed,
		},
	}

	for _, testCase := range testCases {
		t.Run("", func(t *testing.T) {
			pool, err := New(context.Background(), testCase.connStr)
			require.Nil(t, err)

			err = pool.Ping(context.Background())
			require.NotNil(t, err)

			actual := AnalyzeAndWrapPgError(err)
			assert.True(t, errors.IsErrorWithCode(actual, testCase.expectedError), "Actual err: %v", err)
		})
	}

}
