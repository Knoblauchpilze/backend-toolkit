package db

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type dummyConnection struct {
	Connection
}

const sampleSqlQuery = "SELECT name FROM my_table"

func TestUnit_QueryOne(t *testing.T) {
	t.Run("returns error when connection is not supported", func(t *testing.T) {
		_, err := QueryOne[int](t.Context(), &dummyConnection{}, sampleSqlQuery)

		assert.ErrorIs(t, ErrUnsupportedOperation, err, "Actual err: %v", err)
	})

	t.Run("returns error when connection is closed", func(t *testing.T) {
		conn := newTestConnection(t)
		conn.Close(t.Context())

		_, err := QueryOne[int](t.Context(), conn, sampleSqlQuery)

		assert.ErrorIs(t, ErrNotConnected, err, "Actual err: %v", err)
	})

	t.Run("returns error when SQL query is invalid", func(t *testing.T) {
		conn := newTestConnection(t)

		sqlQuery := "SELECT name FROM my_tables"
		_, err := QueryOne[string](t.Context(), conn, sqlQuery)

		actual, ok := AsDatabaseError(err)
		require.True(t, ok)
		assert.Equal(t, ErrGenericSqlError, actual.Code, "Actual err: %v", err)
		assert.NotNil(t, actual.Cause)
	})

	t.Run("returns error when no row matches", func(t *testing.T) {
		conn := newTestConnection(t)

		sqlQuery := "SELECT id, name FROM my_table WHERE name = $1"
		_, err := QueryOne[element](t.Context(), conn, sqlQuery, "does-not-exist")

		assert.ErrorIs(t, ErrNoMatchingRows, err, "Actual err: %v", err)

	})

	t.Run("returns error when more than one row matches", func(t *testing.T) {
		conn := newTestConnection(t)
		v1 := insertTestData(t, conn)
		v2 := insertTestData(t, conn)

		sqlQuery := "SELECT id, name FROM my_table WHERE id IN ($1, $2)"
		_, err := QueryOne[element](t.Context(), conn, sqlQuery, v1.Id, v2.Id)

		assert.ErrorIs(t, ErrTooManyMatchingRows, err, "Actual err: %v", err)
	})

	t.Run("returns error when SQL constraint is violated", func(t *testing.T) {
		conn := newTestConnection(t)
		data := insertTestData(t, conn)

		duplicate := element{
			Id:   uuid.New(),
			Name: data.Name,
		}

		sqlQuery := "INSERT INTO my_table (id, name) VALUES($1, $2)"
		_, err := QueryOne[element](t.Context(), conn, sqlQuery, duplicate.Id, duplicate.Name)

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
		conn := newTestConnection(t)
		expected := insertTestData(t, conn)

		sqlQuery := "SELECT id, name FROM my_table WHERE name = $1"
		actual, err := QueryOne[element](t.Context(), conn, sqlQuery, expected.Name)
		require.NoError(t, err, "Actual err: %v", err)

		assert.Equal(t, expected, actual)
	})

	t.Run("successfully maps to string", func(t *testing.T) {
		conn := newTestConnection(t)
		expected := insertTestData(t, conn)

		sqlQuery := "SELECT name FROM my_table WHERE id = $1"
		actual, err := QueryOne[string](t.Context(), conn, sqlQuery, expected.Id)
		require.NoError(t, err, "Actual err: %v", err)

		assert.Equal(t, expected.Name, actual)
	})

	t.Run("successfully maps to UUID", func(t *testing.T) {
		conn := newTestConnection(t)
		expected := insertTestData(t, conn)

		sqlQuery := "SELECT id FROM my_table WHERE name = $1"
		actual, err := QueryOne[uuid.UUID](t.Context(), conn, sqlQuery, expected.Name)
		require.NoError(t, err, "Actual err: %v", err)

		assert.Equal(t, expected.Id, actual)
	})

	t.Run("successfully maps to time", func(t *testing.T) {
		conn := newTestConnection(t)
		beforeInsert := time.Now()
		expected := insertTestData(t, conn)

		sqlQuery := "SELECT updated_at FROM my_table WHERE name = $1"
		actual, err := QueryOne[time.Time](t.Context(), conn, sqlQuery, expected.Name)
		require.NoError(t, err, "Actual err: %v", err)

		assert.True(t, beforeInsert.Before(actual))
	})
}

func TestUnit_QueryAll(t *testing.T) {
	t.Run("returns error when connection is not supported", func(t *testing.T) {
		_, err := QueryAll[int](t.Context(), &dummyConnection{}, sampleSqlQuery)

		assert.ErrorIs(t, ErrUnsupportedOperation, err, "Actual err: %v", err)
	})

	t.Run("returns error when connection is closed", func(t *testing.T) {
		conn := newTestConnection(t)
		conn.Close(t.Context())

		_, err := QueryAll[int](t.Context(), conn, sampleSqlQuery)

		assert.ErrorIs(t, ErrNotConnected, err, "Actual err: %v", err)
	})

	t.Run("returns error when SQL query is invalid", func(t *testing.T) {
		conn := newTestConnection(t)

		sqlQuery := "SELECT name FROM my_tables"
		_, err := QueryAll[string](t.Context(), conn, sqlQuery)

		actual, ok := AsDatabaseError(err)
		require.True(t, ok)
		assert.Equal(t, ErrGenericSqlError, actual.Code, "Actual err: %v", err)
		assert.NotNil(t, actual.Cause)
	})

	t.Run("successfully fetches no rows", func(t *testing.T) {
		conn := newTestConnection(t)

		sqlQuery := "SELECT id, name FROM my_table WHERE name = $1"
		out, err := QueryAll[element](t.Context(), conn, sqlQuery, "does-not-exist")
		require.NoError(t, err, "Actual err: %v", err)

		assert.Empty(t, out)
	})

	t.Run("successfully maps to struct", func(t *testing.T) {
		conn := newTestConnection(t)
		v1 := insertTestData(t, conn)
		v2 := insertTestData(t, conn)

		sqlQuery := `SELECT id, name FROM my_table WHERE id IN ($1, $2)`
		actual, err := QueryAll[element](t.Context(), conn, sqlQuery, v1.Id, v2.Id)
		require.NoError(t, err, "Actual err: %v", err)

		expected := []element{v1, v2}
		assert.Equal(t, expected, actual)
	})

	t.Run("successfully maps to string", func(t *testing.T) {
		conn := newTestConnection(t)
		v1 := insertTestData(t, conn)
		v2 := insertTestData(t, conn)

		sqlQuery := `SELECT name FROM my_table WHERE id IN ($1, $2)`
		actual, err := QueryAll[string](t.Context(), conn, sqlQuery, v1.Id, v2.Id)
		require.NoError(t, err, "Actual err: %v", err)

		expected := []string{v1.Name, v2.Name}
		assert.Equal(t, expected, actual)
	})

	t.Run("successfully maps to UUID", func(t *testing.T) {
		conn := newTestConnection(t)
		v1 := insertTestData(t, conn)
		v2 := insertTestData(t, conn)

		sqlQuery := `SELECT id FROM my_table WHERE name IN ($1, $2)`
		actual, err := QueryAll[uuid.UUID](t.Context(), conn, sqlQuery, v1.Name, v2.Name)
		require.NoError(t, err, "Actual err: %v", err)

		expected := []uuid.UUID{v1.Id, v2.Id}
		assert.Equal(t, expected, actual)
	})

	t.Run("successfully maps to time", func(t *testing.T) {
		conn := newTestConnection(t)
		beforeInsert := time.Now()
		v1 := insertTestData(t, conn)
		v2 := insertTestData(t, conn)

		sqlQuery := "SELECT updated_at FROM my_table WHERE id IN ($1, $2)"
		actual, err := QueryAll[time.Time](t.Context(), conn, sqlQuery, v1.Id, v2.Id)
		require.NoError(t, err, "Actual err: %v", err)

		assert.True(t, beforeInsert.Before(actual[0]))
		assert.True(t, beforeInsert.Before(actual[1]))
	})
}
