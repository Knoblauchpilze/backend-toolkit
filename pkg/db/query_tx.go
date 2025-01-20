package db

import (
	"context"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db/pgx"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	jpgx "github.com/jackc/pgx/v5"
)

func QueryOneTx[T any](ctx context.Context, tx Transaction, sql string, arguments ...any) (T, error) {
	var out T

	txImpl, ok := tx.(*transactionImpl)
	if !ok {
		return out, errors.NewCode(UnsupportedOperation)
	}
	rows, err := txImpl.query(ctx, sql, arguments...)
	if err != nil {
		return out, pgx.AnalyzeAndWrapPgError(err)
	}

	out, err = jpgx.CollectExactlyOneRow(rows, getCollectorForType[T]())
	if err != nil {
		if err == jpgx.ErrNoRows {
			return out, errors.WrapCode(err, NoMatchingRows)
		} else if err == jpgx.ErrTooManyRows {
			return out, errors.WrapCode(err, TooManyMatchingRows)
		}
		return out, pgx.AnalyzeAndWrapPgError(err)
	}

	return out, nil
}

func QueryAllTx[T any](ctx context.Context, tx Transaction, sql string, arguments ...any) ([]T, error) {
	var out []T

	txImpl, ok := tx.(*transactionImpl)
	if !ok {
		return out, errors.NewCode(UnsupportedOperation)
	}
	rows, err := txImpl.query(ctx, sql, arguments...)
	if err != nil {
		return out, pgx.AnalyzeAndWrapPgError(err)
	}

	out, err = jpgx.CollectRows(rows, getCollectorForType[T]())
	if err != nil {
		return out, errors.WrapCode(err, UnsupportedOperation)
	}

	return out, nil
}
