package process

type WaitFunc func() error

func SafeRunAsync(proc RunFunc) WaitFunc {
	out := make(chan error, 1)

	go func() {
		var err error
		defer func() {
			out <- err
		}()

		defer func() {
			if recovered := recover(); recovered != nil {
				err = wrapError(recovered)
			}
		}()

		err = proc()
	}()

	return func() error {
		return <-out
	}
}
