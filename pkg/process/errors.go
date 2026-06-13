package process

import "github.com/Knoblauchpilze/backend-toolkit/pkg/errors"

const (
	errInvalidProcess errors.ErrorCode = 200
)

var (
	ErrInvalidProcess = errors.FromCode(errInvalidProcess)
)
