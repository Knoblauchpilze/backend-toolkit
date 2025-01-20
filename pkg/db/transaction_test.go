package db

import (
	"context"
	"testing"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db/pgx"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIT_Transaction_Exec_AlreadyCommitted(t *testing.T) {
	_, tx := newTestTransaction(t)
	tx.Close(context.Background())

	affectedRows, err := tx.Exec(context.Background(), "SELECT COUNT(*) FROM my_table WHERE name = 'test-name'")

	assert.Equal(t, int64(0), affectedRows)
	assert.True(t, errors.IsErrorWithCode(err, AlreadyCommitted), "Actual err: %v", err)
}

func TestIT_Transaction_Exec_Select(t *testing.T) {
	_, tx := newTestTransaction(t)
	defer tx.Close(context.Background())

	affectedRows, err := tx.Exec(context.Background(), "SELECT COUNT(*) FROM my_table WHERE name = 'test-name'")
	assert.Equal(t, int64(1), affectedRows)
	assert.Nil(t, err)
}

func TestIT_Transaction_Exec_Insert(t *testing.T) {
	conn, tx := newTestTransaction(t)

	id := uuid.New()
	// Also using a uuid for the name to easily generate characters
	name := uuid.New()
	_, err := tx.Exec(context.Background(), "INSERT INTO my_table VALUES ($1, $2)", id, name)
	require.Nil(t, err)

	tx.Close(context.Background())

	assertNameForId(t, conn, id, name.String())
}

func TestIT_Transaction_Exec_Update(t *testing.T) {
	conn, tx := newTestTransaction(t)
	element := insertTestDataTx(t, tx)

	newName := uuid.New().String()
	_, err := tx.Exec(context.Background(), "UPDATE my_table SET name = $1 WHERE id = $2", newName, element.Id)
	require.Nil(t, err)

	tx.Close(context.Background())

	assertNameForId(t, conn, element.Id, newName)
}

func TestIT_Transaction_Exec_Delete(t *testing.T) {
	conn, tx := newTestTransaction(t)
	element := insertTestDataTx(t, tx)

	_, err := tx.Exec(context.Background(), "DELETE FROM my_table WHERE id = $1", element.Id)
	require.Nil(t, err)

	tx.Close(context.Background())
	assertIdDoesNotExist(t, conn, element.Id)
}

func TestIT_Transaction_Exec_WithArguments(t *testing.T) {
	_, tx := newTestTransaction(t)
	defer tx.Close(context.Background())

	affectedRows, err := tx.Exec(context.Background(), "SELECT COUNT(*) FROM my_table WHERE name = $1", "test-name")

	assert.Equal(t, int64(1), affectedRows)
	assert.Nil(t, err)
}

func TestIT_Transaction_Exec_WrongSyntax(t *testing.T) {
	_, tx := newTestTransaction(t)
	defer tx.Close(context.Background())

	affectedRows, err := tx.Exec(context.Background(), "DESELECT COUNT(*) FROM my_table WHERE name = 'test-name'")

	assert.Equal(t, int64(0), affectedRows)
	assert.True(t, errors.IsErrorWithCode(err, pgx.GenericSqlError), "Actual err: %v", err)
}

func TestIT_Transaction_Exec_WhenError_ExpectRollback(t *testing.T) {
	conn, tx := newTestTransaction(t)

	element := insertTestDataTx(t, tx)
	_, err := tx.Exec(context.Background(), "DESELECT COUNT(*) FROM my_table WHERE name = $1", element.Name)
	require.NotNil(t, err)

	tx.Close(context.Background())

	assertIdDoesNotExist(t, conn, element.Id)
}
