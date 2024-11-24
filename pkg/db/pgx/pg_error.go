package pgx

import (
	"strings"

	"github.com/KnoblauchPilze/backend-toolkit/pkg/errors"
	"github.com/jackc/pgx/v5/pgconn"
)

// https://www.postgresql.org/docs/current/errcodes-appendix.html
const (
	foreignKeyViolation          = "23503"
	uniqueValidation             = "23505"
	passwordAuthenticationFailed = "28P01"
)

func AnalyzeAndWrapPgError(err error) error {
	if err == nil {
		return nil
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
	switch err.Code {
	case foreignKeyViolation:
		return errors.WrapCode(err, ForeignKeyValidation)
	case uniqueValidation:
		return errors.WrapCode(err, UniqueConstraintViolation)
	}

	return errors.WrapCode(err, GenericSqlError)
}

func analyzeConnError(err *pgconn.ConnectError) error {
	msg := err.Unwrap().Error()
	if strings.Contains(msg, passwordAuthenticationFailed) {
		return errors.NewCodeWithDetails(AuthenticationFailed, "Failed to connect to database")
	}

	return errors.NewCode(GenericSqlError)
}
