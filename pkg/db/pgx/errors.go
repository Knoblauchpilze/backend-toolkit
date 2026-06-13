package pgx

import "github.com/Knoblauchpilze/backend-toolkit/pkg/errors"

const (
	ErrGenericSqlError           errors.ErrorCode = 150
	ErrForeignKeyValidation      errors.ErrorCode = 151
	ErrUniqueConstraintViolation errors.ErrorCode = 152
	errAuthenticationFailed      errors.ErrorCode = 153
)

var (
	ErrAuthenticationFailed = errors.FromCode(errAuthenticationFailed)
)
