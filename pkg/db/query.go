package db

import (
	"context"
	"reflect"
	"time"

	"github.com/jackc/pgx/v5"
)

type DbConnection interface {
	Query(ctx context.Context, sql string, arguments ...any) (pgx.Rows, error)
}

func QueryOne[T any](ctx context.Context, conn DbConnection, sql string, arguments ...any) (T, error) {
	var out T

	rows, err := conn.Query(ctx, sql, arguments...)
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

func QueryAll[T any](ctx context.Context, conn DbConnection, sql string, arguments ...any) ([]T, error) {
	var out []T

	rows, err := conn.Query(ctx, sql, arguments...)
	if err != nil {
		return out, analyzeAndWrapDatabaseError(err)
	}

	out, err = pgx.CollectRows(rows, getCollectorForType[T]())
	if err != nil {
		return out, ErrUnsupportedOperation
	}

	return out, nil
}

var timeStructName = reflect.ValueOf(time.Time{}).Type().Name()

func getCollectorForType[T any]() pgx.RowToFunc[T] {
	var value T

	kind := reflect.ValueOf(value).Kind()
	typeName := reflect.ValueOf(value).Type().Name()

	// https://pkg.go.dev/github.com/jackc/pgx/v5#RowToStructByName
	if kind == reflect.Struct &&
		typeName != timeStructName {
		return pgx.RowToStructByName[T]
	}

	return pgx.RowTo[T]
}
