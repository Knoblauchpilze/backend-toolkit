package db

import (
	"context"
	"testing"
	"time"

	"github.com/KnoblauchPilze/backend-toolkit/pkg/db/pgx"
	"github.com/KnoblauchPilze/backend-toolkit/pkg/db/postgresql"
	"github.com/KnoblauchPilze/backend-toolkit/pkg/errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnit_New_InvalidConfiguration(t *testing.T) {
	config := postgresql.Config{
		Host: ":/not-a-host",
	}

	conn, err := New(context.Background(), config)

	assert.Nil(t, conn)
	assert.NotNil(t, err)
}

func TestIT_New_ValidConfiguration_InvalidCredentials(t *testing.T) {
	config := dbTestConfig
	config.Password = "not-the-right-password"

	conn, err := New(context.Background(), config)

	assert.NotNil(t, conn)
	assert.True(t, errors.IsErrorWithCode(err, pgx.AuthenticationFailed), "Actual err: %v", err)
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
	assert.True(t, errors.IsErrorWithCode(err, NotConnected), "Actual err: %v", err)
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
	assert.True(t, errors.IsErrorWithCode(err, NotConnected), "Actual err: %v", err)
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
	assert.True(t, errors.IsErrorWithCode(err, pgx.UniqueConstraintViolation), "Actual err: %v", err)
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
	assert.True(t, errors.IsErrorWithCode(err, pgx.UniqueConstraintViolation), "Actual err: %v", err)

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
	assert.True(t, errors.IsErrorWithCode(err, pgx.GenericSqlError), "Actual err: %v", err)
}
