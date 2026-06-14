package db

import (
	"strings"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/jackc/pgx/v5/pgconn"
)

// https://www.postgresql.org/docs/current/errcodes-appendix.html
const (
	foreignKeyViolation          = "23503"
	uniqueValidation             = "23505"
	passwordAuthenticationFailed = "28P01"
)

func analyzeAndWrapDatabaseError(err error) error {
	if err == nil {
		return nil
	}

	if prepErr, ok := err.(*pgconn.PrepareError); ok {
		return analyzeAndWrapDatabaseError(prepErr.Unwrap())
	}

	if pgErr, ok := err.(*pgconn.PgError); ok {
		return analyzePgError(pgErr)
	}

	if connErr, ok := err.(*pgconn.ConnectError); ok {
		return analyzeConnError(connErr)
	}

	return err
}

func analyzePgError(err *pgconn.PgError) error {
	return &DatabaseError{
		Code:       mapPostgreCodeToErrorCode(err.Code),
		Message:    err.Message,
		SqlCode:    err.Code,
		Schema:     err.SchemaName,
		Table:      err.TableName,
		Column:     err.ColumnName,
		Constraint: err.ConstraintName,
		Cause:      err,
	}
}

func analyzeConnError(err *pgconn.ConnectError) error {
	msg := err.Unwrap().Error()

	if strings.Contains(msg, passwordAuthenticationFailed) {
		return ErrAuthenticationFailed
	}

	return errors.FromCode(ErrGenericSqlError)
}

func mapPostgreCodeToErrorCode(postgreCode string) errors.ErrorCode {
	switch postgreCode {
	case foreignKeyViolation:
		return ErrForeignKeyValidation
	case uniqueValidation:
		return ErrUniqueConstraintViolation
	}

	return ErrGenericSqlError
}
