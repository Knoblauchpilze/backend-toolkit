package db

import (
	"context"
	"testing"
	"time"

	"github.com/KnoblauchPilze/backend-toolkit/pkg/db/pgx"
	"github.com/KnoblauchPilze/backend-toolkit/pkg/errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type dummyConnection struct {
	Connection
}

const sampleSqlQuery = "SELECT name FROM my_table"

func TestUnit_QueryOne_UnsupportedConnection(t *testing.T) {
	_, err := QueryOne[int](context.Background(), &dummyConnection{}, sampleSqlQuery)

	assert.NotNil(t, err)
	assert.True(t, errors.IsErrorWithCode(err, UnsupportedOperation), "Actual err: %v", err)
}

func TestIT_QueryOne_WhenClosed_ExpectFailure(t *testing.T) {
	conn := newTestConnection(t)
	conn.Close(context.Background())

	_, err := QueryOne[int](context.Background(), conn, sampleSqlQuery)

	assert.NotNil(t, err)
	assert.True(t, errors.IsErrorWithCode(err, NotConnected), "Actual err: %v", err)
}

func TestIT_QueryOne_WhenConnectionFails_ExpectFailure(t *testing.T) {
	conn := newTestConnection(t)

	sqlQuery := "SELECT name FROM my_tables"
	_, err := QueryOne[string](context.Background(), conn, sqlQuery)

	assert.NotNil(t, err)
	assert.True(t, errors.IsErrorWithCode(err, pgx.GenericSqlError), "Actual err: %v", err)

	cause := errors.Unwrap(err)
	assert.NotNil(t, cause)
}

func TestIT_QueryOne_WhenNoData_ExpectFailure(t *testing.T) {
	conn := newTestConnection(t)

	sqlQuery := "SELECT id, name FROM my_table WHERE name = $1"
	_, err := QueryOne[element](context.Background(), conn, sqlQuery, "does-not-exist")

	assert.NotNil(t, err)
	assert.True(t, errors.IsErrorWithCode(err, NoMatchingRows), "Actual err: %v", err)
}

func TestIT_QueryOne_WhenTooManyRows_ExpectFailure(t *testing.T) {
	conn := newTestConnection(t)
	v1 := insertTestData(t, conn)
	v2 := insertTestData(t, conn)

	sqlQuery := "SELECT id, name FROM my_table WHERE id IN ($1, $2)"
	_, err := QueryOne[element](context.Background(), conn, sqlQuery, v1.Id, v2.Id)

	assert.NotNil(t, err)
	assert.True(t, errors.IsErrorWithCode(err, TooManyMatchingRows), "Actual err: %v", err)
}

func TestIT_QueryOne_ToStruct(t *testing.T) {
	conn := newTestConnection(t)
	expected := insertTestData(t, conn)

	sqlQuery := "SELECT id, name FROM my_table WHERE name = $1"
	actual, err := QueryOne[element](context.Background(), conn, sqlQuery, expected.Name)

	assert.Nil(t, err)
	assert.Equal(t, expected, actual)
}

func TestIT_QueryOne_ToString(t *testing.T) {
	conn := newTestConnection(t)
	expected := insertTestData(t, conn)

	sqlQuery := "SELECT name FROM my_table WHERE id = $1"
	actual, err := QueryOne[string](context.Background(), conn, sqlQuery, expected.Id)

	assert.Nil(t, err)
	assert.Equal(t, expected.Name, actual)
}

func TestIT_QueryOne_ToUuid(t *testing.T) {
	conn := newTestConnection(t)
	expected := insertTestData(t, conn)

	sqlQuery := "SELECT id FROM my_table WHERE name = $1"
	actual, err := QueryOne[uuid.UUID](context.Background(), conn, sqlQuery, expected.Name)

	assert.Nil(t, err)
	assert.Equal(t, expected.Id, actual)
}

func TestIT_QueryOne_ToTime(t *testing.T) {
	conn := newTestConnection(t)
	beforeInsert := time.Now()
	expected := insertTestData(t, conn)

	sqlQuery := "SELECT updated_at FROM my_table WHERE name = $1"
	actual, err := QueryOne[time.Time](context.Background(), conn, sqlQuery, expected.Name)

	assert.Nil(t, err)
	assert.True(t, beforeInsert.Before(actual))
}

func TestIT_QueryAll_UnsupportedConnection(t *testing.T) {
	_, err := QueryAll[int](context.Background(), &dummyConnection{}, sampleSqlQuery)

	assert.NotNil(t, err)
	assert.True(t, errors.IsErrorWithCode(err, UnsupportedOperation), "Actual err: %v", err)
}

func TestIT_QueryAll_WhenClosed_ExpectFailure(t *testing.T) {
	conn := newTestConnection(t)
	conn.Close(context.Background())

	_, err := QueryAll[int](context.Background(), conn, sampleSqlQuery)

	assert.NotNil(t, err)
	assert.True(t, errors.IsErrorWithCode(err, NotConnected), "Actual err: %v", err)
}

func TestIT_QueryAll_WhenConnectionFails_ExpectFailure(t *testing.T) {
	conn := newTestConnection(t)

	sqlQuery := "SELECT name FROM my_tables"
	_, err := QueryAll[string](context.Background(), conn, sqlQuery)

	assert.NotNil(t, err)
	assert.True(t, errors.IsErrorWithCode(err, pgx.GenericSqlError), "Actual err: %v", err)

	cause := errors.Unwrap(err)
	assert.NotNil(t, cause)
}

func TestIT_QueryAll_NoData(t *testing.T) {
	conn := newTestConnection(t)

	sqlQuery := "SELECT id, name FROM my_table WHERE name = $1"
	out, err := QueryAll[element](context.Background(), conn, sqlQuery, "does-not-exist")

	assert.Nil(t, err)
	assert.Empty(t, out)
}

func TestIT_QueryAll_ToStruct(t *testing.T) {
	conn := newTestConnection(t)
	v1 := insertTestData(t, conn)
	v2 := insertTestData(t, conn)

	sqlQuery := `SELECT id, name FROM my_table WHERE id IN ($1, $2)`
	actual, err := QueryAll[element](context.Background(), conn, sqlQuery, v1.Id, v2.Id)

	assert.Nil(t, err)
	expected := []element{v1, v2}
	assert.Equal(t, expected, actual)
}

func TestIT_QueryAll_ToString(t *testing.T) {
	conn := newTestConnection(t)
	v1 := insertTestData(t, conn)
	v2 := insertTestData(t, conn)

	sqlQuery := `SELECT name FROM my_table WHERE id IN ($1, $2)`
	actual, err := QueryAll[string](context.Background(), conn, sqlQuery, v1.Id, v2.Id)

	assert.Nil(t, err)
	expected := []string{v1.Name, v2.Name}
	assert.Equal(t, expected, actual)
}

func TestIT_QueryAll_ToUuid(t *testing.T) {
	conn := newTestConnection(t)
	v1 := insertTestData(t, conn)
	v2 := insertTestData(t, conn)

	sqlQuery := `SELECT id FROM my_table WHERE name IN ($1, $2)`
	actual, err := QueryAll[uuid.UUID](context.Background(), conn, sqlQuery, v1.Name, v2.Name)

	assert.Nil(t, err)
	expected := []uuid.UUID{v1.Id, v2.Id}
	assert.Equal(t, expected, actual)
}

func TestIT_QueryAll_ToTime(t *testing.T) {
	conn := newTestConnection(t)
	beforeInsert := time.Now()
	v1 := insertTestData(t, conn)
	v2 := insertTestData(t, conn)

	sqlQuery := "SELECT updated_at FROM my_table WHERE id IN ($1, $2)"
	actual, err := QueryAll[time.Time](context.Background(), conn, sqlQuery, v1.Id, v2.Id)

	assert.Nil(t, err)
	assert.True(t, beforeInsert.Before(actual[0]))
	assert.True(t, beforeInsert.Before(actual[1]))
}
