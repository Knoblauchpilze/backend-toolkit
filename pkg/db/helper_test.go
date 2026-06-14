package db

import (
	"context"
	"testing"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db/postgresql"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

var dbTestConfig = postgresql.NewConfigForLocalhost("test_db", "test_user", "test_password")

type element struct {
	Id   uuid.UUID
	Name string
}

func newTestConnection(t *testing.T) Connection {
	t.Helper()

	conn, err := New(context.Background(), dbTestConfig)
	require.NoError(t, err, "Actual err: %v", err)

	t.Cleanup(func() {
		conn.Close(t.Context())
	})

	return conn
}

func newTestTransaction(t *testing.T) (Connection, Transaction) {
	t.Helper()

	conn := newTestConnection(t)

	tx, err := conn.BeginTx(context.Background())
	require.NoError(t, err, "Actual err: %v", err)

	t.Cleanup(func() {
		tx.Close(t.Context())
	})

	return conn, tx
}

func insertTestData(t *testing.T, conn Connection) element {
	t.Helper()

	element := element{
		Id:   uuid.New(),
		Name: uuid.NewString(),
	}
	_, err := conn.Exec(context.Background(), "INSERT INTO my_table VALUES ($1, $2)", element.Id, element.Name)
	require.NoError(t, err, "Actual err: %v", err)

	return element
}

func insertTestDataTx(t *testing.T, tx Transaction) element {
	t.Helper()

	element := element{
		Id:   uuid.New(),
		Name: uuid.NewString(),
	}
	_, err := tx.Exec(context.Background(), "INSERT INTO my_table VALUES ($1, $2)", element.Id, element.Name)
	require.NoError(t, err, "Actual err: %v", err)

	return element
}

func assertNameForId(t *testing.T, conn Connection, id uuid.UUID, expectedName string) {
	t.Helper()

	value, err := QueryOne[string](context.Background(), conn, "SELECT name FROM my_table WHERE id = $1", id)
	require.NoError(t, err, "Actual err: %v", err)
	require.Equal(t, expectedName, value)
}

func assertIdExists(t *testing.T, conn Connection, id uuid.UUID) {
	t.Helper()

	value, err := QueryOne[int](context.Background(), conn, "SELECT COUNT(*) FROM my_table WHERE id = $1", id)
	require.NoError(t, err, "Actual err: %v", err)
	require.Equal(t, 1, value)
}

func assertIdDoesNotExist(t *testing.T, conn Connection, id uuid.UUID) {
	t.Helper()

	value, err := QueryOne[int](context.Background(), conn, "SELECT COUNT(*) FROM my_table WHERE id = $1", id)
	require.NoError(t, err, "Actual err: %v", err)
	require.Equal(t, 0, value)
}
