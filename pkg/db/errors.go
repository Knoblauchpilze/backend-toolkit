package db

import "github.com/Knoblauchpilze/backend-toolkit/pkg/errors"

const (
	errNotConnected         errors.ErrorCode = 100
	errUnsupportedOperation errors.ErrorCode = 101
	errAlreadyCommitted     errors.ErrorCode = 102

	errNoMatchingRows      errors.ErrorCode = 110
	errTooManyMatchingRows errors.ErrorCode = 111

	ErrGenericSqlError           errors.ErrorCode = 150
	ErrForeignKeyValidation      errors.ErrorCode = 151
	ErrUniqueConstraintViolation errors.ErrorCode = 152
	errAuthenticationFailed      errors.ErrorCode = 153
)

var ()

var (
	ErrNotConnected         = errors.FromCode(errNotConnected)
	ErrUnsupportedOperation = errors.FromCode(errUnsupportedOperation)
	ErrAlreadyCommitted     = errors.FromCode(errAlreadyCommitted)

	ErrNoMatchingRows      = errors.FromCode(errNoMatchingRows)
	ErrTooManyMatchingRows = errors.FromCode(errTooManyMatchingRows)

	ErrAuthenticationFailed = errors.FromCode(errAuthenticationFailed)
)
