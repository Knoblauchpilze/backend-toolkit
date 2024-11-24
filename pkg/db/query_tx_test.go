package db

import (
	"context"
	"testing"

	"github.com/KnoblauchPilze/backend-toolkit/pkg/db/pgx"
	"github.com/KnoblauchPilze/backend-toolkit/pkg/errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type dummyTransaction struct {
	Transaction
}

func TestUnit_QueryOneTx_UnsupportedConnection(t *testing.T) {
	_, err := QueryOneTx[int](context.Background(), &dummyTransaction{}, sampleSqlQuery)

	assert.NotNil(t, err)
	assert.True(t, errors.IsErrorWithCode(err, UnsupportedOperation), "Actual err: %v", err)
}

func TestIT_QueryOneTx_WhenCommitted_ExpectFailure(t *testing.T) {
	_, tx := newTestTransaction(t)
	tx.Close(context.Background())

	_, err := QueryOneTx[int](context.Background(), tx, sampleSqlQuery)

	assert.NotNil(t, err)
	assert.True(t, errors.IsErrorWithCode(err, AlreadyCommitted), "Actual err: %v", err)
}

func TestIT_QueryOneTx_WhenConnectionFails_ExpectFailure(t *testing.T) {
	_, tx := newTestTransaction(t)

	sqlQuery := "SELECT name FROM my_tables"
	_, err := QueryOneTx[string](context.Background(), tx, sqlQuery)

	assert.NotNil(t, err)
	assert.True(t, errors.IsErrorWithCode(err, pgx.GenericSqlError), "Actual err: %v", err)

	cause := errors.Unwrap(err)
	assert.NotNil(t, cause)
}

func TestIT_QueryOneTx_WhenNoData_ExpectFailure(t *testing.T) {
	_, tx := newTestTransaction(t)

	sqlQuery := "SELECT id, name FROM my_table WHERE name = $1"
	_, err := QueryOneTx[element](context.Background(), tx, sqlQuery, "does-not-exist")

	assert.NotNil(t, err)
	assert.True(t, errors.IsErrorWithCode(err, NoMatchingRows), "Actual err: %v", err)
}

func TestIT_QueryOneTx_WhenTooManyRows_ExpectFailure(t *testing.T) {
	_, tx := newTestTransaction(t)
	v1 := insertTestDataTx(t, tx)
	v2 := insertTestDataTx(t, tx)

	sqlQuery := "SELECT id, name FROM my_table WHERE id IN ($1, $2)"
	_, err := QueryOneTx[element](context.Background(), tx, sqlQuery, v1.Id, v2.Id)

	assert.NotNil(t, err)
	assert.True(t, errors.IsErrorWithCode(err, TooManyMatchingRows), "Actual err: %v", err)
}

func TestIT_QueryOneTx_ToStruct(t *testing.T) {
	_, tx := newTestTransaction(t)
	expected := insertTestDataTx(t, tx)

	sqlQuery := "SELECT id, name FROM my_table WHERE name = $1"
	actual, err := QueryOneTx[element](context.Background(), tx, sqlQuery, expected.Name)

	assert.Nil(t, err)
	assert.Equal(t, expected, actual)
}

func TestIT_QueryOneTx_ToString(t *testing.T) {
	_, tx := newTestTransaction(t)
	expected := insertTestDataTx(t, tx)

	sqlQuery := "SELECT name FROM my_table WHERE id = $1"
	actual, err := QueryOneTx[string](context.Background(), tx, sqlQuery, expected.Id)

	assert.Nil(t, err)
	assert.Equal(t, expected.Name, actual)
}

func TestIT_QueryOneTx_ToUuid(t *testing.T) {
	_, tx := newTestTransaction(t)
	expected := insertTestDataTx(t, tx)

	sqlQuery := "SELECT id FROM my_table WHERE name = $1"
	actual, err := QueryOneTx[uuid.UUID](context.Background(), tx, sqlQuery, expected.Name)

	assert.Nil(t, err)
	assert.Equal(t, expected.Id, actual)
}

func TestIT_QueryAllTx_UnsupportedConnection(t *testing.T) {
	_, err := QueryAllTx[int](context.Background(), &dummyTransaction{}, sampleSqlQuery)

	assert.NotNil(t, err)
	assert.True(t, errors.IsErrorWithCode(err, UnsupportedOperation), "Actual err: %v", err)
}

func TestIT_QueryAllTx_WhenCommitted_ExpectFailure(t *testing.T) {
	_, tx := newTestTransaction(t)
	tx.Close(context.Background())

	_, err := QueryAllTx[int](context.Background(), tx, sampleSqlQuery)

	assert.NotNil(t, err)
	assert.True(t, errors.IsErrorWithCode(err, AlreadyCommitted), "Actual err: %v", err)
}

func TestIT_QueryAllTx_WhenConnectionFails_ExpectFailure(t *testing.T) {
	_, tx := newTestTransaction(t)

	sqlQuery := "SELECT name FROM my_tables"
	_, err := QueryAllTx[string](context.Background(), tx, sqlQuery)

	assert.NotNil(t, err)
	assert.True(t, errors.IsErrorWithCode(err, pgx.GenericSqlError), "Actual err: %v", err)

	cause := errors.Unwrap(err)
	assert.NotNil(t, cause)
}

func TestIT_QueryAllTx_NoData(t *testing.T) {
	_, tx := newTestTransaction(t)

	sqlQuery := "SELECT id, name FROM my_table WHERE name = $1"
	out, err := QueryAllTx[element](context.Background(), tx, sqlQuery, "does-not-exist")

	assert.Nil(t, err)
	assert.Empty(t, out)
}

func TestIT_QueryAllTx_ToStruct(t *testing.T) {
	_, tx := newTestTransaction(t)
	v1 := insertTestDataTx(t, tx)
	v2 := insertTestDataTx(t, tx)

	sqlQuery := `SELECT id, name FROM my_table WHERE id IN ($1, $2)`
	actual, err := QueryAllTx[element](context.Background(), tx, sqlQuery, v1.Id, v2.Id)

	assert.Nil(t, err)
	expected := []element{v1, v2}
	assert.Equal(t, expected, actual)
}

func TestIT_QueryAllTx_ToString(t *testing.T) {
	_, tx := newTestTransaction(t)
	v1 := insertTestDataTx(t, tx)
	v2 := insertTestDataTx(t, tx)

	sqlQuery := `SELECT name FROM my_table WHERE id IN ($1, $2)`
	actual, err := QueryAllTx[string](context.Background(), tx, sqlQuery, v1.Id, v2.Id)

	assert.Nil(t, err)
	expected := []string{v1.Name, v2.Name}
	assert.Equal(t, expected, actual)
}

func TestIT_QueryAllTx_ToUuid(t *testing.T) {
	_, tx := newTestTransaction(t)
	v1 := insertTestDataTx(t, tx)
	v2 := insertTestDataTx(t, tx)

	sqlQuery := `SELECT id FROM my_table WHERE name IN ($1, $2)`
	actual, err := QueryAllTx[uuid.UUID](context.Background(), tx, sqlQuery, v1.Name, v2.Name)

	assert.Nil(t, err)
	expected := []uuid.UUID{v1.Id, v2.Id}
	assert.Equal(t, expected, actual)
}
