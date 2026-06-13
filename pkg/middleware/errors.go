package middleware

import "github.com/Knoblauchpilze/backend-toolkit/pkg/errors"

const (
	errUncaughtPanic errors.ErrorCode = 400
)

var (
	ErrUncaughtPanic = errors.FromCode(errUncaughtPanic)
)
