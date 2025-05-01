package process

import "context"

type Runnable interface {
	Start() error
	Stop() error
}

func StartWithSignalHandler(ctx context.Context, runnable Runnable) (WaitFunc, error) {
	process := Process{
		Run:       runnable.Start,
		Interrupt: runnable.Stop,
	}

	return AsyncStartWithSignalHandler(ctx, process)
}
