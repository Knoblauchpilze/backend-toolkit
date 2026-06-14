package db

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db/postgresql"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
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
	require.Nil(t, err)

	conn.Close(context.Background())
	err = conn.Ping(context.Background())
	assert.Equal(t, ErrNotConnected, err, "Actual err: %v", err)
}

func TestIT_Connection_BeginTx_TimeStampIsValid(t *testing.T) {
	conn := newTestConnection(t)

	beforeTx := time.Now()
	tx, err := conn.BeginTx(context.Background())

	assert.Nil(t, err)
	assert.True(t, beforeTx.Before(tx.TimeStamp()))
}

func TestIT_Connection_BeginTx_ClosedConnection(t *testing.T) {
	conn := newTestConnection(t)
	conn.Close(context.Background())

	tx, err := conn.BeginTx(context.Background())

	assert.Nil(t, tx)
	assert.Equal(t, ErrNotConnected, err, "Actual err: %v", err)
}

func TestIT_Connection_Exec_Select(t *testing.T) {
	conn := newTestConnection(t)
	element := insertTestData(t, conn)

	affectedRows, err := conn.Exec(context.Background(), "SELECT COUNT(*) FROM my_table WHERE id = $1", element.Id)

	assert.Equal(t, int64(1), affectedRows)
	assert.Nil(t, err)
}

func TestIT_Connection_Exec_Insert(t *testing.T) {
	conn := newTestConnection(t)

	id := uuid.New()
	// Also using a uuid for the name to easily generate characters
	name := uuid.New()
	affectedRows, err := conn.Exec(context.Background(), "INSERT INTO my_table VALUES ($1, $2)", id, name)

	assert.Equal(t, int64(1), affectedRows)
	assert.Nil(t, err)

	assertIdExists(t, conn, id)
}

func TestIT_Connection_Exec_InsertDuplicate(t *testing.T) {
	conn := newTestConnection(t)
	element := insertTestData(t, conn)
	id := uuid.New()

	affectedRows, err := conn.Exec(context.Background(), "INSERT INTO my_table VALUES ($1, $2)", id, element.Name)

	assert.Equal(t, int64(0), affectedRows)
	actual, ok := errors.AsErrorWithCode(err)
	require.True(t, ok)
	assert.Equal(t, ErrUniqueConstraintViolation, actual.Code, "Actual err: %v", err)
	assertIdDoesNotExist(t, conn, id)
}

func TestIT_Connection_Exec_Update(t *testing.T) {
	conn := newTestConnection(t)
	element := insertTestData(t, conn)
	newName := uuid.New().String()

	affectedRows, err := conn.Exec(context.Background(), "UPDATE my_table SET name = $1 WHERE id = $2", newName, element.Id)
	assert.Equal(t, int64(1), affectedRows)
	assert.Nil(t, err)

	assertNameForId(t, conn, element.Id, newName)
}

func TestIT_Connection_Exec_UpdateDuplicate(t *testing.T) {
	conn := newTestConnection(t)
	element := insertTestData(t, conn)
	anotherElement := insertTestData(t, conn)

	affectedRows, err := conn.Exec(context.Background(), "UPDATE my_table SET name = $1 WHERE id = $2", anotherElement.Name, element.Id)
	assert.Equal(t, int64(0), affectedRows)
	actual, ok := errors.AsErrorWithCode(err)
	require.True(t, ok)
	assert.Equal(t, ErrUniqueConstraintViolation, actual.Code, "Actual err: %v", err)

	assertNameForId(t, conn, element.Id, element.Name)
}

func TestIT_Connection_Exec_Delete(t *testing.T) {
	conn := newTestConnection(t)
	element := insertTestData(t, conn)

	affectedRows, err := conn.Exec(context.Background(), "DELETE FROM my_table WHERE id = $1", element.Id)
	assert.Equal(t, int64(1), affectedRows)
	assert.Nil(t, err)

	assertIdDoesNotExist(t, conn, element.Id)
}

func TestIT_Connection_Exec_WithArguments(t *testing.T) {
	conn := newTestConnection(t)
	element := insertTestData(t, conn)

	affectedRows, err := conn.Exec(context.Background(), "SELECT COUNT(*) FROM my_table WHERE name = $1", element.Name)

	assert.Equal(t, int64(1), affectedRows)
	assert.Nil(t, err)
}

func TestIT_Connection_Exec_WrongSyntax(t *testing.T) {
	conn := newTestConnection(t)
	element := insertTestData(t, conn)

	affectedRows, err := conn.Exec(context.Background(), "DESELECT COUNT(*) FROM my_table WHERE name = $1", element.Name)

	assert.Equal(t, int64(0), affectedRows)
	actual, ok := errors.AsErrorWithCode(err)
	require.True(t, ok)
	assert.Equal(t, ErrGenericSqlError, actual.Code, "Actual err: %v", err)
}

func TestIT_Connection_Exec_ReturnsZonedTimeAsUTC(t *testing.T) {
	conn := newTestConnection(t)

	berlinTz, err := time.LoadLocation("Europe/Berlin")
	require.NoError(t, err)

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
	require.NoError(t, err)

	actual, err := QueryOne[time.Time](
		context.Background(),
		conn,
		"SELECT created_at FROM my_table WHERE id = $1",
		element.Id,
	)
	require.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, zonedTime.UTC(), actual)
}

func TestIT_Connection_Exec_ReturnsUTCTimeAsUTC(t *testing.T) {
	conn := newTestConnection(t)

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
	require.NoError(t, err)

	actual, err := QueryOne[time.Time](
		context.Background(),
		conn,
		"SELECT created_at FROM my_table WHERE id = $1",
		element.Id,
	)
	require.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, utcTime, actual)
}
