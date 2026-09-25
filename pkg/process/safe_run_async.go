package process

func SafeRunAsync(proc RunFunc) <-chan error {
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

	return out
}
