package process

import (
	"fmt"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
)

func SafeRunAsync(proc RunFunc) <-chan error {
	out := make(chan error, 1)

	go func() {
		var err error
		defer func() {
			out <- err
		}()

		defer func() {
			if recovered := recover(); recovered != nil {
				if asErr, ok := recovered.(error); ok {
					err = asErr
				} else {
					err = errors.New(fmt.Sprintf("%v", recovered))
				}
			}
		}()

		err = proc()
	}()

	return out
}
