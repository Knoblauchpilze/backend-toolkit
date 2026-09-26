package db

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type DbTransaction interface {
	Query(ctx context.Context, sql string, arguments ...any) (pgx.Rows, error)
}

func QueryOneTx[T any](ctx context.Context, tx DbTransaction, sql string, arguments ...any) (T, error) {
	var out T

	rows, err := tx.Query(ctx, sql, arguments...)
	if err != nil {
		return out, analyzeAndWrapDatabaseError(err)
	}

	out, err = pgx.CollectExactlyOneRow(rows, getCollectorForType[T]())
	if err != nil {
		switch err {
		case pgx.ErrNoRows:
			return out, ErrNoMatchingRows
		case pgx.ErrTooManyRows:
			return out, ErrTooManyMatchingRows
		default:
			return out, analyzeAndWrapDatabaseError(err)
		}
	}

	return out, nil
}

func QueryAllTx[T any](ctx context.Context, tx DbTransaction, sql string, arguments ...any) ([]T, error) {
	var out []T

	rows, err := tx.Query(ctx, sql, arguments...)
	if err != nil {
		return out, analyzeAndWrapDatabaseError(err)
	}

	out, err = pgx.CollectRows(rows, getCollectorForType[T]())
	if err != nil {
		return out, ErrUnsupportedOperation
	}

	return out, nil
}
