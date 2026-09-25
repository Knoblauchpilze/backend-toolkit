package process

import "github.com/Knoblauchpilze/backend-toolkit/pkg/errors"

const (
	errInvalidProcess errors.ErrorCode = 200
	errPanicRecovered errors.ErrorCode = 201
)

var (
	ErrInvalidProcess = errors.FromCode(errInvalidProcess)
)
