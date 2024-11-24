package pgx

import "github.com/KnoblauchPilze/backend-toolkit/pkg/errors"

const (
	GenericSqlError           errors.ErrorCode = 150
	ForeignKeyValidation      errors.ErrorCode = 151
	UniqueConstraintViolation errors.ErrorCode = 152
	AuthenticationFailed      errors.ErrorCode = 153
)
