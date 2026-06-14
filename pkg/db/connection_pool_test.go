package db

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnit_NewPool(t *testing.T) {
	t.Run("returns error when connection string is invalid", func(t *testing.T) {
		pool, err := newPool(context.Background(), "invalid-connection-string")

		assert.Nil(t, pool)
		assert.NotNil(t, err)
		_, ok := err.(*pgconn.ParseConfigError)
		assert.True(t, ok)
	})

	t.Run("returns no error when connection string is valid", func(t *testing.T) {
		connStr := "postgres://user:password@localhost/my-db"
		pool, err := newPool(context.Background(), connStr)

		assert.NotNil(t, pool)
		assert.Nil(t, err)
	})
}

func TestIT_NewPool(t *testing.T) {
	t.Run("successfully connects to database", func(t *testing.T) {
		connStr := "postgres://test_user:test_password@localhost:5432/test_db"
		pool, err := newPool(context.Background(), connStr)
		require.NoError(t, err, "Actual err: %v", err)

		err = pool.Ping(context.Background())
		require.NoError(t, err, "Actual err: %v", err)
	})
}
