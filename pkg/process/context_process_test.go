package process

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnit_NewContextProcess(t *testing.T) {
	t.Run("interrupt cancels run context", func(t *testing.T) {
		started := make(chan struct{})
		runDone := make(chan error, 1)
		proc := NewContextProcess(func(ctx context.Context) error {
			close(started)
			<-ctx.Done()
			return ctx.Err()
		})
		require.True(t, proc.Valid())

		go func() {
			runDone <- proc.Run()
		}()
		<-started

		err := proc.Interrupt()
		require.NoError(t, err, "Actual err: %v", err)
		assert.ErrorIs(t, <-runDone, context.Canceled)
	})

	t.Run("interrupt call is idempotent", func(t *testing.T) {
		started := make(chan struct{})
		runDone := make(chan error, 1)
		proc := NewContextProcess(func(ctx context.Context) error {
			close(started)
			<-ctx.Done()
			return ctx.Err()
		})
		require.True(t, proc.Valid())

		go func() {
			runDone <- proc.Run()
		}()
		<-started

		err := proc.Interrupt()
		require.NoError(t, err, "Actual err: %v", err)
		assert.ErrorIs(t, <-runDone, context.Canceled)

		err = proc.Interrupt()
		require.NoError(t, err, "Actual err: %v", err)
	})

	t.Run("attaches run function to process run function", func(t *testing.T) {
		var called int

		proc := NewContextProcess(func(context.Context) error {
			called++
			return nil
		})

		err := proc.Run()
		require.NoError(t, err, "Actual err: %v", err)

		assert.Equal(t, 1, called)
	})

	t.Run("returns error produced by run function", func(t *testing.T) {
		proc := NewContextProcess(func(context.Context) error {
			return errSample
		})

		err := proc.Run()
		assert.ErrorIs(t, err, errSample, "Actual err: %v", err)
	})

	t.Run("returns invalid process when run function is nil", func(t *testing.T) {
		proc := NewContextProcess(nil)

		assert.False(t, proc.Valid())
	})
}
