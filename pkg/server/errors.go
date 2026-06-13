package server

import "github.com/Knoblauchpilze/backend-toolkit/pkg/errors"

const (
	errUnsupportedMethod errors.ErrorCode = 300
)

var (
	ErrUnsupportedMethod = errors.FromCode(errUnsupportedMethod)
)
