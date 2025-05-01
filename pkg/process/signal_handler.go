package process

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
)

var defaultSignals = []os.Signal{
	syscall.SIGINT,
	os.Interrupt,
}

type WaitFunc func() error

func AsyncStartWithSignalHandler(
	ctx context.Context,
	process Process,
) (WaitFunc, error) {
	if !process.Valid() {
		return nil, errors.NewCode(ErrInvalidProcess)
	}

	sCtx, stop := signal.NotifyContext(ctx, defaultSignals...)

	done := make(chan error, 1)

	go func() {
		err := SafeRunSync(process.Run)
		done <- err
	}()

	waitFunc := func() error {
		defer stop()

		var err error

		select {
		case <-sCtx.Done():
			err = process.Interrupt()
		case err = <-done:
		}

		// It can be that the process was interrupted by sCtx and that
		// we have an error ready in the done channel. Here we read it
		// and replace the error.
		// Note: this overrides a potential error from the interrupt
		// process.
		// https://stackoverflow.com/questions/3398490/checking-if-a-channel-has-a-ready-to-read-value-using-go
		select {
		case runErr, ok := <-done:
			if ok && runErr != nil {
				err = runErr
			}
		default:
			// No error in done channel, continuing
		}

		return err
	}

	return waitFunc, nil
}
