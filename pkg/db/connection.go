package db

import (
	"context"
	"time"

	"github.com/KnoblauchPilze/backend-toolkit/pkg/db/pgx"
	"github.com/KnoblauchPilze/backend-toolkit/pkg/errors"
	jpgx "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Connection interface {
	Close(ctx context.Context)
	Ping(ctx context.Context) error

	BeginTx(ctx context.Context) (Transaction, error)

	Exec(ctx context.Context, sql string, arguments ...any) (int64, error)
}

type connectionImpl struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, config Config) (Connection, error) {
	connStr := config.ToConnectionString()

	pool, err := pgx.New(ctx, connStr)
	if err != nil {
		return nil, err
	}

	conn := &connectionImpl{
		pool: pool,
	}

	err = conn.Ping(ctx)

	return conn, pgx.AnalyzeAndWrapPgError(err)
}

func (ci *connectionImpl) Close(ctx context.Context) {
	if ci.pool != nil {
		ci.pool.Close()
		ci.pool = nil
	}
}

func (ci *connectionImpl) Ping(ctx context.Context) error {
	if ci.pool == nil {
		return errors.NewCode(NotConnected)
	}
	return ci.pool.Ping(ctx)
}

func (ci *connectionImpl) BeginTx(ctx context.Context) (Transaction, error) {
	if ci.pool == nil {
		return nil, errors.NewCode(NotConnected)
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

func (ci *connectionImpl) Exec(ctx context.Context, sql string, arguments ...any) (int64, error) {
	if ci.pool == nil {
		return 0, errors.NewCode(NotConnected)
	}

	tag, err := ci.pool.Exec(ctx, sql, arguments...)
	if err != nil {
		return tag.RowsAffected(), pgx.AnalyzeAndWrapPgError(err)
	}

	return tag.RowsAffected(), err
}

func (ci *connectionImpl) query(ctx context.Context, sql string, arguments ...any) (jpgx.Rows, error) {
	if ci.pool == nil {
		return nil, errors.NewCode(NotConnected)
	}
	rows, err := ci.pool.Query(ctx, sql, arguments...)

	err = pgx.AnalyzeAndWrapPgError(err)
	if err != nil {
		return nil, err
	}

	return rows, nil
}
