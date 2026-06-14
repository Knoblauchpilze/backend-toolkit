package db

import (
	"context"

	"github.com/jackc/pgx/v5"
)

func QueryOneTx[T any](ctx context.Context, tx Transaction, sql string, arguments ...any) (T, error) {
	var out T

	txImpl, ok := tx.(*transactionImpl)
	if !ok {
		return out, ErrUnsupportedOperation
	}
	rows, err := txImpl.query(ctx, sql, arguments...)
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

func QueryAllTx[T any](ctx context.Context, tx Transaction, sql string, arguments ...any) ([]T, error) {
	var out []T

	txImpl, ok := tx.(*transactionImpl)
	if !ok {
		return out, ErrUnsupportedOperation
	}
	rows, err := txImpl.query(ctx, sql, arguments...)
	if err != nil {
		return out, analyzeAndWrapDatabaseError(err)
	}

	out, err = pgx.CollectRows(rows, getCollectorForType[T]())
	if err != nil {
		return out, ErrUnsupportedOperation
	}

	return out, nil
}
