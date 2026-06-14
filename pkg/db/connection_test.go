package db

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db/postgresql"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnit_New(t *testing.T) {
	t.Run("fails when configuration is invalid", func(t *testing.T) {
		config := postgresql.Config{
			Host: ":/not-a-host",
		}

		conn, err := New(context.Background(), config)

		assert.Nil(t, conn)
		assert.Error(t, err)
	})

	t.Run("fails when credentials are invalid", func(t *testing.T) {
		config := dbTestConfig
		config.Password = "not-the-right-password"

		fmt.Printf("config: %+v\n", config)
		conn, err := New(context.Background(), config)

		assert.NotNil(t, conn)
		assert.Equal(t, ErrAuthenticationFailed, err, "Actual err: %v", err)
	})
}

func TestIT_New_ValidConfiguration(t *testing.T) {
	conn, err := New(context.Background(), dbTestConfig)

	assert.NotNil(t, conn)
	assert.Nil(t, err)
}

func TestIT_Connection_Ping(t *testing.T) {
	conn := newTestConnection(t)

	err := conn.Ping(context.Background())
	assert.Nil(t, err)
}

func TestIT_Connection_Close(t *testing.T) {
	conn := newTestConnection(t)

	err := conn.Ping(context.Background())
	require.NoError(t, err, "Actual err: %v", err)

	conn.Close(context.Background())
	err = conn.Ping(context.Background())
	assert.Equal(t, ErrNotConnected, err, "Actual err: %v", err)
}

func TestIT_Connection_BeginTx_TimeStampIsValid(t *testing.T) {
	conn := newTestConnection(t)

	t.Run("assigns timestamp when beginning transaction", func(t *testing.T) {

		beforeTx := time.Now()
		tx, err := conn.BeginTx(context.Background())
		require.NoError(t, err, "Actual err: %v", err)

		defer func() {
			// nolint: errcheck
			tx.Close(t.Context())
		}()

		assert.True(t, beforeTx.Before(tx.TimeStamp()))
	})

	t.Run("returns error when connection is closed", func(t *testing.T) {
		conn.Close(context.Background())

		tx, err := conn.BeginTx(context.Background())

		assert.Nil(t, tx)
		assert.ErrorIs(t, ErrNotConnected, err, "Actual err: %v", err)
	})
}

func TestIT_Connection_Exec(t *testing.T) {
	conn := newTestConnection(t)

	t.Run("successfully selects data", func(t *testing.T) {
		element := insertTestData(t, conn)

		affectedRows, err := conn.Exec(context.Background(), "SELECT COUNT(*) FROM my_table WHERE id = $1", element.Id)
		require.NoError(t, err, "Actual err: %v", err)

		assert.Equal(t, int64(1), affectedRows)
	})

	t.Run("successfully inserts data", func(t *testing.T) {
		id := uuid.New()
		// Also using a uuid for the name to easily generate characters
		name := uuid.New()
		affectedRows, err := conn.Exec(context.Background(), "INSERT INTO my_table VALUES ($1, $2)", id, name)
		require.NoError(t, err, "Actual err: %v", err)

		assert.Equal(t, int64(1), affectedRows)
		assertIdExists(t, conn, id)
	})

	t.Run("returns error when unique constraint is violated", func(t *testing.T) {
		element := insertTestData(t, conn)
		id := uuid.New()

		affectedRows, err := conn.Exec(context.Background(), "INSERT INTO my_table VALUES ($1, $2)", id, element.Name)

		assert.Equal(t, int64(0), affectedRows)
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
		assertIdDoesNotExist(t, conn, id)
	})

	t.Run("successfull updates data", func(t *testing.T) {
		element := insertTestData(t, conn)
		newName := uuid.New().String()

		affectedRows, err := conn.Exec(context.Background(), "UPDATE my_table SET name = $1 WHERE id = $2", newName, element.Id)
		require.NoError(t, err, "Actual err: %v", err)

		assert.Equal(t, int64(1), affectedRows)
		assertNameForId(t, conn, element.Id, newName)
	})

	t.Run("returns error when update leads to unique constraint being violated", func(t *testing.T) {
		element := insertTestData(t, conn)
		anotherElement := insertTestData(t, conn)

		affectedRows, err := conn.Exec(context.Background(), "UPDATE my_table SET name = $1 WHERE id = $2", anotherElement.Name, element.Id)

		assert.Equal(t, int64(0), affectedRows)
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
		assertNameForId(t, conn, element.Id, element.Name)
	})

	t.Run("successfully deletes data", func(t *testing.T) {
		element := insertTestData(t, conn)

		affectedRows, err := conn.Exec(context.Background(), "DELETE FROM my_table WHERE id = $1", element.Id)
		require.NoError(t, err, "Actual err: %v", err)

		assert.Equal(t, int64(1), affectedRows)
		assertIdDoesNotExist(t, conn, element.Id)
	})

	t.Run("successfully propagates provided arguments", func(t *testing.T) {
		element := insertTestData(t, conn)

		affectedRows, err := conn.Exec(context.Background(), "SELECT COUNT(*) FROM my_table WHERE name = $1", element.Name)
		require.NoError(t, err, "Actual err: %v", err)

		assert.Equal(t, int64(1), affectedRows)
	})

	t.Run("returns error when SQL query is invalid", func(t *testing.T) {
		element := insertTestData(t, conn)

		affectedRows, err := conn.Exec(context.Background(), "DESELECT COUNT(*) FROM my_table WHERE name = $1", element.Name)

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

	t.Run("returns zoned time as UTC", func(t *testing.T) {
		berlinTz, err := time.LoadLocation("Europe/Berlin")
		require.NoError(t, err, "Actual err: %v", err)

		zonedTime := time.Date(2026, 05, 30, 13, 57, 29, 0, berlinTz)

		element := element{
			Id:   uuid.New(),
			Name: uuid.NewString(),
		}
		_, err = conn.Exec(
			context.Background(),
			"INSERT INTO my_table VALUES ($1, $2, $3)",
			element.Id, element.Name, zonedTime,
		)
		require.NoError(t, err, "Actual err: %v", err)

		actual, err := QueryOne[time.Time](
			context.Background(),
			conn,
			"SELECT created_at FROM my_table WHERE id = $1",
			element.Id,
		)
		require.NoError(t, err, "Actual err: %v", err)

		assert.Equal(t, zonedTime.UTC(), actual)
	})

	t.Run("returns UTC time as UTC", func(t *testing.T) {
		utcTime := time.Date(2026, 05, 30, 13, 57, 29, 0, time.UTC)

		element := element{
			Id:   uuid.New(),
			Name: uuid.NewString(),
		}
		_, err := conn.Exec(
			context.Background(),
			"INSERT INTO my_table VALUES ($1, $2, $3)",
			element.Id, element.Name, utcTime,
		)
		require.NoError(t, err, "Actual err: %v", err)

		actual, err := QueryOne[time.Time](
			context.Background(),
			conn,
			"SELECT created_at FROM my_table WHERE id = $1",
			element.Id,
		)
		require.NoError(t, err, "Actual err: %v", err)

		assert.Equal(t, utcTime, actual)
	})
}
