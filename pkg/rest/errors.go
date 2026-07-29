package rest

import "github.com/Knoblauchpilze/backend-toolkit/pkg/errors"

const (
	errMultipleBodyWrite errors.ErrorCode = 350
)

var (
	ErrMultipleBodyWrite = errors.FromCodeAndDetails(errMultipleBodyWrite, "response body already written")
)