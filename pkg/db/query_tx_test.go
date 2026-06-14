package db

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type dummyTransaction struct {
	Transaction
}

func TestUnit_QueryOneTx(t *testing.T) {
	t.Run("returns error when transaction is not supported", func(t *testing.T) {
		_, err := QueryOneTx[int](context.Background(), &dummyTransaction{}, sampleSqlQuery)

		assert.ErrorIs(t, ErrUnsupportedOperation, err, "Actual err: %v", err)
	})

	t.Run("returns error when already committed", func(t *testing.T) {
		_, tx := newTestTransaction(t)
		tx.Close(context.Background())

		_, err := QueryOneTx[int](context.Background(), tx, sampleSqlQuery)

		assert.ErrorIs(t, ErrAlreadyCommitted, err, "Actual err: %v", err)
	})

	t.Run("returns error when SQL query is invalid", func(t *testing.T) {
		_, tx := newTestTransaction(t)

		sqlQuery := "SELECT name FROM my_tables"
		_, err := QueryOneTx[string](context.Background(), tx, sqlQuery)

		actual, ok := AsDatabaseError(err)
		require.True(t, ok)
		assert.Equal(t, ErrGenericSqlError, actual.Code, "Actual err: %v", err)
		assert.NotNil(t, actual.Cause)
	})

	t.Run("returns error when no row matches", func(t *testing.T) {
		_, tx := newTestTransaction(t)

		sqlQuery := "SELECT id, name FROM my_table WHERE name = $1"
		_, err := QueryOneTx[element](context.Background(), tx, sqlQuery, "does-not-exist")

		assert.ErrorIs(t, ErrNoMatchingRows, err, "Actual err: %v", err)
	})

	t.Run("returns error when more than one row matches", func(t *testing.T) {
		_, tx := newTestTransaction(t)
		v1 := insertTestDataTx(t, tx)
		v2 := insertTestDataTx(t, tx)

		sqlQuery := "SELECT id, name FROM my_table WHERE id IN ($1, $2)"
		_, err := QueryOneTx[element](context.Background(), tx, sqlQuery, v1.Id, v2.Id)

		assert.ErrorIs(t, ErrTooManyMatchingRows, err, "Actual err: %v", err)
	})

	t.Run("returns error when SQL constraint is violated", func(t *testing.T) {
		conn, tx := newTestTransaction(t)
		data := insertTestData(t, conn)

		duplicate := element{
			Id:   uuid.New(),
			Name: data.Name,
		}

		sqlQuery := "INSERT INTO my_table (id, name) VALUES($1, $2)"
		_, err := QueryOneTx[element](context.Background(), tx, sqlQuery, duplicate.Id, duplicate.Name)

		actual, ok := AsDatabaseError(err)
		require.True(t, ok)
		expected := &DatabaseError{
			Code:       ErrUniqueConstraintViolation,
			Message:    "duplicate key value violates unique constraint \"my_table_name_key\"",
			SqlCode:    "23505",
			Schema:     "test_db_schema",
			Table:      "my_table",
			Column:     "",
			Constraint: "my_table_name_key",
			Cause:      actual.Cause,
		}
		assert.Equal(t, expected, actual, "Actual err: %v", err)
	})

	t.Run("successfully maps to struct", func(t *testing.T) {
		_, tx := newTestTransaction(t)
		expected := insertTestDataTx(t, tx)

		sqlQuery := "SELECT id, name FROM my_table WHERE name = $1"
		actual, err := QueryOneTx[element](context.Background(), tx, sqlQuery, expected.Name)
		require.NoError(t, err, "Actual err: %v", err)

		assert.Equal(t, expected, actual)
	})

	t.Run("successfully maps to string", func(t *testing.T) {
		_, tx := newTestTransaction(t)
		expected := insertTestDataTx(t, tx)

		sqlQuery := "SELECT name FROM my_table WHERE id = $1"
		actual, err := QueryOneTx[string](context.Background(), tx, sqlQuery, expected.Id)
		require.NoError(t, err, "Actual err: %v", err)

		assert.Equal(t, expected.Name, actual)
	})

	t.Run("successfully maps to UUID", func(t *testing.T) {
		_, tx := newTestTransaction(t)
		expected := insertTestDataTx(t, tx)

		sqlQuery := "SELECT id FROM my_table WHERE name = $1"
		actual, err := QueryOneTx[uuid.UUID](context.Background(), tx, sqlQuery, expected.Name)
		require.NoError(t, err, "Actual err: %v", err)

		assert.Equal(t, expected.Id, actual)
	})

	t.Run("successfully maps to time", func(t *testing.T) {
		conn, tx := newTestTransaction(t)
		beforeInsert := time.Now()
		expected := insertTestData(t, conn)

		sqlQuery := "SELECT updated_at FROM my_table WHERE name = $1"
		actual, err := QueryOneTx[time.Time](context.Background(), tx, sqlQuery, expected.Name)
		require.NoError(t, err, "Actual err: %v", err)

		assert.True(t, beforeInsert.Before(actual))
	})
}

func TestUnit_QueryAllTx(t *testing.T) {
	t.Run("returns error when transaction is not supported", func(t *testing.T) {
		_, err := QueryAllTx[int](context.Background(), &dummyTransaction{}, sampleSqlQuery)

		assert.ErrorIs(t, ErrUnsupportedOperation, err, "Actual err: %v", err)
	})

	t.Run("returns error when already committed", func(t *testing.T) {
		_, tx := newTestTransaction(t)
		tx.Close(context.Background())

		_, err := QueryAllTx[int](context.Background(), tx, sampleSqlQuery)

		assert.ErrorIs(t, ErrAlreadyCommitted, err, "Actual err: %v", err)
	})

	t.Run("returns error when connection is closed", func(t *testing.T) {
		conn := newTestConnection(t)
		conn.Close(context.Background())

		_, err := QueryAll[int](context.Background(), conn, sampleSqlQuery)

		assert.ErrorIs(t, ErrNotConnected, err, "Actual err: %v", err)
	})

	t.Run("returns error when SQL query is invalid", func(t *testing.T) {
		_, tx := newTestTransaction(t)

		sqlQuery := "SELECT name FROM my_tables"
		_, err := QueryAllTx[string](context.Background(), tx, sqlQuery)

		actual, ok := AsDatabaseError(err)
		require.True(t, ok)
		assert.Equal(t, ErrGenericSqlError, actual.Code, "Actual err: %v", err)
		assert.NotNil(t, actual.Cause)
	})

	t.Run("successfully fetches no rows", func(t *testing.T) {
		_, tx := newTestTransaction(t)

		sqlQuery := "SELECT id, name FROM my_table WHERE name = $1"
		out, err := QueryAllTx[element](context.Background(), tx, sqlQuery, "does-not-exist")
		require.NoError(t, err, "Actual err: %v", err)

		assert.Empty(t, out)
	})

	t.Run("successfully maps to struct", func(t *testing.T) {
		_, tx := newTestTransaction(t)
		v1 := insertTestDataTx(t, tx)
		v2 := insertTestDataTx(t, tx)

		sqlQuery := `SELECT id, name FROM my_table WHERE id IN ($1, $2)`
		actual, err := QueryAllTx[element](context.Background(), tx, sqlQuery, v1.Id, v2.Id)
		require.NoError(t, err, "Actual err: %v", err)

		expected := []element{v1, v2}
		assert.Equal(t, expected, actual)
	})

	t.Run("successfully maps to string", func(t *testing.T) {
		_, tx := newTestTransaction(t)
		v1 := insertTestDataTx(t, tx)
		v2 := insertTestDataTx(t, tx)

		sqlQuery := `SELECT name FROM my_table WHERE id IN ($1, $2)`
		actual, err := QueryAllTx[string](context.Background(), tx, sqlQuery, v1.Id, v2.Id)
		require.NoError(t, err, "Actual err: %v", err)

		expected := []string{v1.Name, v2.Name}
		assert.Equal(t, expected, actual)
	})

	t.Run("successfully maps to UUID", func(t *testing.T) {
		_, tx := newTestTransaction(t)
		v1 := insertTestDataTx(t, tx)
		v2 := insertTestDataTx(t, tx)

		sqlQuery := `SELECT id FROM my_table WHERE name IN ($1, $2)`
		actual, err := QueryAllTx[uuid.UUID](context.Background(), tx, sqlQuery, v1.Name, v2.Name)
		require.NoError(t, err, "Actual err: %v", err)

		expected := []uuid.UUID{v1.Id, v2.Id}
		assert.Equal(t, expected, actual)
	})

	t.Run("successfully maps to time", func(t *testing.T) {
		conn, tx := newTestTransaction(t)
		beforeInsert := time.Now()
		v1 := insertTestData(t, conn)
		v2 := insertTestData(t, conn)

		sqlQuery := "SELECT updated_at FROM my_table WHERE id IN ($1, $2)"
		actual, err := QueryAllTx[time.Time](context.Background(), tx, sqlQuery, v1.Id, v2.Id)
		require.NoError(t, err, "Actual err: %v", err)

		assert.True(t, beforeInsert.Before(actual[0]))
		assert.True(t, beforeInsert.Before(actual[1]))
	})
}
