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
				if asErr, ok := recovered.(error); ok {
					err = asErr
				} else {
					err = errors.New(fmt.Sprintf("%v", recovered))
				}
			}
		}()

		proc()
	}()

	return err
}
