package db

import (
	"context"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db/pgx"
	jpgx "github.com/jackc/pgx/v5"
)

func QueryOneTx[T any](ctx context.Context, tx Transaction, sql string, arguments ...any) (T, error) {
	var out T

	txImpl, ok := tx.(*transactionImpl)
	if !ok {
		return out, ErrUnsupportedOperation
	}
	rows, err := txImpl.query(ctx, sql, arguments...)
	if err != nil {
		return out, pgx.AnalyzeAndWrapPgError(err)
	}

	out, err = jpgx.CollectExactlyOneRow(rows, getCollectorForType[T]())
	if err != nil {
		switch err {
		case jpgx.ErrNoRows:
			return out, ErrNoMatchingRows
		case jpgx.ErrTooManyRows:
			return out, ErrTooManyMatchingRows
		default:
			return out, pgx.AnalyzeAndWrapPgError(err)
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
		return out, pgx.AnalyzeAndWrapPgError(err)
	}

	out, err = jpgx.CollectRows(rows, getCollectorForType[T]())
	if err != nil {
		return out, ErrUnsupportedOperation
	}

	return out, nil
}
