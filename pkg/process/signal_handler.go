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

		return err
	}

	return waitFunc, nil
}
