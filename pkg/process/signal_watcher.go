package process

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

var defaultSignals = []os.Signal{
	syscall.SIGINT,
	os.Interrupt,
}

// watchForCancellation blocks until one of the following happens:
//   - ctx is cancelled
//   - one of the defaultSignals is received by the process
//   - a value is received on interrupted
//
// In the first two cases, interrupt is called to request that the running
// process shuts down and its return value is propagated. In the last case,
// watchForCancellation returns nil without calling interrupt, as this is
// used to signal that watching is no longer needed (e.g. because the
// process already completed on its own).
func watchForCancellation(
	ctx context.Context,
	interrupted <-chan struct{},
	interrupt InterruptFunc,
) error {
	sCtx, stop := signal.NotifyContext(ctx, defaultSignals...)
	defer stop()

	select {
	case <-sCtx.Done():
		return interrupt()
	case <-interrupted:
		return nil
	}
}
