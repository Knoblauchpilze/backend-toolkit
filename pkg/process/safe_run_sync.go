package process

import (
	"fmt"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
)

func SafeRunSync(proc RunFunc) error {
	var err error

	func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				err = wrapError(recovered)
			}
		}()

		err = proc()
	}()

	return err
}

func wrapError(cause any) error {
	asErr, ok := cause.(error)
	if ok {
		return asErr
	}

	msg := fmt.Sprintf("%v", cause)
	return errors.FromCodeAndDetails(errPanicRecovered, msg)
}
