package process

import (
	"context"
	"fmt"
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
		fmt.Printf("started async process\n")
		err := SafeRunSync(process.Run)
		fmt.Printf("process returned with err: %v\n", err)
		done <- err
		fmt.Printf("async loop finished\n")
	}()

	waitFunc := func() error {
		defer stop()

		var err error

		select {
		case <-sCtx.Done():
			fmt.Printf("interrupting async process\n")
			err = process.Interrupt()
			fmt.Printf("interrupt returned: %v\n", err)
		case err = <-done:
			fmt.Printf("async process finished: %v\n", err)
		}

		fmt.Printf("waiting finished, err: %v\n", err)
		return err
	}

	return waitFunc, nil
}
