package db

import (
	"context"
	"reflect"

	"github.com/KnoblauchPilze/backend-toolkit/pkg/db/pgx"
	"github.com/KnoblauchPilze/backend-toolkit/pkg/errors"
	jpgx "github.com/jackc/pgx/v5"
)

func QueryOne[T any](ctx context.Context, conn Connection, sql string, arguments ...any) (T, error) {
	var out T

	connImpl, ok := conn.(*connectionImpl)
	if !ok {
		return out, errors.NewCode(UnsupportedOperation)
	}
	rows, err := connImpl.query(ctx, sql, arguments...)
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
		return out, err
	}

	return out, nil
}

func QueryAll[T any](ctx context.Context, conn Connection, sql string, arguments ...any) ([]T, error) {
	var out []T

	connImpl, ok := conn.(*connectionImpl)
	if !ok {
		return out, errors.NewCode(UnsupportedOperation)
	}
	rows, err := connImpl.query(ctx, sql, arguments...)
	if err != nil {
		return out, pgx.AnalyzeAndWrapPgError(err)
	}

	out, err = jpgx.CollectRows(rows, getCollectorForType[T]())
	if err != nil {
		return out, errors.WrapCode(err, UnsupportedOperation)
	}

	return out, nil
}

func getCollectorForType[T any]() jpgx.RowToFunc[T] {
	var value T

	// https://pkg.go.dev/github.com/jackc/pgx/v5#RowToStructByName
	if reflect.ValueOf(value).Kind() == reflect.Struct {
		return jpgx.RowToStructByName[T]
	}

	return jpgx.RowTo[T]
}