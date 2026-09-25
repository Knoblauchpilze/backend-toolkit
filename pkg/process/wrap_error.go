package process

import (
	"fmt"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
)

func wrapError(cause any) error {
	asErr, ok := cause.(error)
	if ok {
		return asErr
	}

	msg := fmt.Sprintf("%v", cause)
	return errors.FromCodeAndDetails(errPanicRecovered, msg)
}
