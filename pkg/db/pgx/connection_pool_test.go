package pgx

import (
	"context"
	"testing"

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
