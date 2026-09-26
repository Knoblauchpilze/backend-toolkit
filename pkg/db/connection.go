package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Connection struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, config Config) (*Connection, error) {
	connStr := config.ToConnectionString()

	pool, err := newPool(ctx, connStr)
	if err != nil {
		return nil, err
	}

	conn := &Connection{
		pool: pool,
	}

	err = conn.Ping(ctx)
	if err != nil {
		return nil, analyzeAndWrapDatabaseError(err)
	}

	return conn, nil
}

func (ci *Connection) Close(ctx context.Context) {
	if ci.pool != nil {
		ci.pool.Close()
		ci.pool = nil
	}
}

func (ci *Connection) Ping(ctx context.Context) error {
	if ci.pool == nil {
		return ErrNotConnected
	}
	return ci.pool.Ping(ctx)
}

func (ci *Connection) BeginTx(ctx context.Context) (Transaction, error) {
	if ci.pool == nil {
		return nil, ErrNotConnected
	}

	pgxTx, err := ci.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}

	tx := &transactionImpl{
		timeStamp: time.Now(),
		tx:        pgxTx,
	}

	return tx, nil
}

func (ci *Connection) Exec(ctx context.Context, sql string, arguments ...any) (int64, error) {
	if ci.pool == nil {
		return 0, ErrNotConnected
	}

	tag, err := ci.pool.Exec(ctx, sql, arguments...)
	if err != nil {
		return tag.RowsAffected(), analyzeAndWrapDatabaseError(err)
	}

	return tag.RowsAffected(), err
}

func (ci *Connection) Query(ctx context.Context, sql string, arguments ...any) (pgx.Rows, error) {
	if ci.pool == nil {
		return nil, ErrNotConnected
	}
	rows, err := ci.pool.Query(ctx, sql, arguments...)

	err = analyzeAndWrapDatabaseError(err)
	if err != nil {
		return nil, err
	}

	return rows, nil
}
