package process

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
