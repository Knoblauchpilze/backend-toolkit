package process

import "context"

// NewContextProcess adapts a context-aware function to a Process. Interrupt
// cancels the context passed to run.
func NewContextProcess(run func(context.Context) error) Process {
	if run == nil {
		return Process{}
	}

	ctx, cancel := context.WithCancel(context.Background())

	return Process{
		Run: func() error {
			return run(ctx)
		},
		Interrupt: func() error {
			cancel()
			return nil
		},
	}
}
