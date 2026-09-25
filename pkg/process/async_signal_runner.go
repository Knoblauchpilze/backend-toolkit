package process

import "context"

// AsyncStartWithSignalHandler starts process.Run in the background and
// returns a WaitFunc that blocks until it completes, either on its own,
// when ctx is cancelled, or when one of the defaultSignals is received.
// In the last two cases, process.Interrupt is called to request that
// process.Run returns; its error is only reported if process.Run did not
// already return one of its own.
// The error returned by this function can only be not nil when the process
// failed to be started (e.g. because it is invalid).
func AsyncStartWithSignalHandler(
	ctx context.Context,
	process Process,
) (WaitFunc, error) {
	if !process.Valid() {
		return nil, ErrInvalidProcess
	}

	runWait := SafeRunAsync(process.Run)

	interrupted := make(chan struct{})
	watchDone := make(chan error, 1)
	go func() {
		watchDone <- watchForCancellation(ctx, interrupted, process.Interrupt)
	}()

	wait := func() error {
		runErr := runWait()
		// Lets watchForCancellation know it can stop watching if it has
		// not already been triggered by ctx or a signal.
		close(interrupted)
		watchErr := <-watchDone

		if runErr != nil {
			return runErr
		}
		return watchErr
	}

	return wait, nil
}
