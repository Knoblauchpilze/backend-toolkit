package db

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIT_Transaction_Close(t *testing.T) {
	t.Run("calling close is idempotent", func(t *testing.T) {
		_, tx := newTestTransaction(t)
		tx.Close(t.Context())

		tx.Close(t.Context())
	})
}

func TestIT_Transaction_Exec(t *testing.T) {
	t.Run("successfully selects data", func(t *testing.T) {
		_, tx := newTestTransaction(t)

		affectedRows, err := tx.Exec(t.Context(), "SELECT COUNT(*) FROM my_table WHERE name = 'test-name'")
		require.NoError(t, err, "Actual err: %v", err)

		assert.Equal(t, int64(1), affectedRows)
	})

	t.Run("returns error when already committed", func(t *testing.T) {
		_, tx := newTestTransaction(t)
		tx.Close(t.Context())

		affectedRows, err := tx.Exec(t.Context(), "SELECT COUNT(*) FROM my_table WHERE name = 'test-name'")

		assert.Equal(t, int64(0), affectedRows)
		assert.ErrorIs(t, ErrAlreadyCommitted, err, "Actual err: %v", err)
	})

	t.Run("successfully inserts data", func(t *testing.T) {
		conn, tx := newTestTransaction(t)

		id := uuid.New()
		// Also using a uuid for the name to easily generate characters
		name := uuid.New()

		_, err := tx.Exec(t.Context(), "INSERT INTO my_table VALUES ($1, $2)", id, name)
		require.NoError(t, err, "Actual err: %v", err)

		tx.Close(t.Context())

		assertNameForId(t, conn, id, name.String())
	})

	t.Run("successfull updates data", func(t *testing.T) {
		conn, tx := newTestTransaction(t)
		element := insertTestDataTx(t, tx)

		newName := uuid.New().String()
		_, err := tx.Exec(t.Context(), "UPDATE my_table SET name = $1 WHERE id = $2", newName, element.Id)
		require.NoError(t, err, "Actual err: %v", err)

		tx.Close(t.Context())

		assertNameForId(t, conn, element.Id, newName)
	})

	t.Run("successfully deletes data", func(t *testing.T) {
		conn, tx := newTestTransaction(t)
		element := insertTestDataTx(t, tx)

		_, err := tx.Exec(t.Context(), "DELETE FROM my_table WHERE id = $1", element.Id)
		require.NoError(t, err, "Actual err: %v", err)

		tx.Close(t.Context())
		assertIdDoesNotExist(t, conn, element.Id)
	})

	t.Run("successfully propagates provided arguments", func(t *testing.T) {
		_, tx := newTestTransaction(t)
		defer tx.Close(t.Context())

		affectedRows, err := tx.Exec(t.Context(), "SELECT COUNT(*) FROM my_table WHERE name = $1", "test-name")
		require.NoError(t, err, "Actual err: %v", err)

		assert.Equal(t, int64(1), affectedRows)
	})

	t.Run("returns error when SQL query is invalid", func(t *testing.T) {
		_, tx := newTestTransaction(t)
		defer tx.Close(t.Context())

		affectedRows, err := tx.Exec(t.Context(), "DESELECT COUNT(*) FROM my_table WHERE name = 'test-name'")

		assert.Equal(t, int64(0), affectedRows)
		actual, ok := AsDatabaseError(err)
		require.True(t, ok)

		expected := &DatabaseError{
			Code:       ErrGenericSqlError,
			Message:    "syntax error at or near \"DESELECT\"",
			SqlCode:    "42601",
			Schema:     "",
			Table:      "",
			Column:     "",
			Constraint: "",
			Cause:      actual.Cause,
		}
		assert.Equal(t, expected, actual, "Actual err: %v", err)
	})

	t.Run("rollbacks when error is detected", func(t *testing.T) {
		conn, tx := newTestTransaction(t)

		element := insertTestDataTx(t, tx)
		_, err := tx.Exec(t.Context(), "DESELECT COUNT(*) FROM my_table WHERE name = $1", element.Name)
		require.NotNil(t, err)

		tx.Close(t.Context())

		assertIdDoesNotExist(t, conn, element.Id)
	})
}
