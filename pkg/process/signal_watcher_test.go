package process

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnit_WatchForCancellation(t *testing.T) {
	t.Run("stops when the context is cancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		interrupted := make(chan struct{})

		var interruptCalled int
		interrupt := func() error {
			interruptCalled++
			return nil
		}

		done := make(chan error, 1)
		go func() {
			done <- watchForCancellation(ctx, interrupted, interrupt)
		}()

		cancel()

		err := <-done
		require.NoError(t, err, "Actual err: %v", err)
		assert.Equal(t, 1, interruptCalled)
	})

	t.Run("stops when the interrupted channel is closed", func(t *testing.T) {
		interrupted := make(chan struct{})

		var interruptCalled int
		interrupt := func() error {
			interruptCalled++
			return nil
		}

		done := make(chan error, 1)
		go func() {
			done <- watchForCancellation(t.Context(), interrupted, interrupt)
		}()

		close(interrupted)

		err := <-done
		require.NoError(t, err, "Actual err: %v", err)
		assert.Zero(t, interruptCalled)
	})

	t.Run("returns the interrupt error", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		interrupted := make(chan struct{})

		interrupt := func() error {
			return errSample
		}

		done := make(chan error, 1)
		go func() {
			done <- watchForCancellation(ctx, interrupted, interrupt)
		}()

		cancel()

		err := <-done
		assert.Equal(t, errSample, err, "Actual err: %v", err)
	})

	t.Run("does not call interrupt when the interrupted channel fires first", func(t *testing.T) {
		interrupted := make(chan struct{})

		var interruptCalled int
		interrupt := func() error {
			interruptCalled++
			return nil
		}

		done := make(chan error, 1)
		go func() {
			done <- watchForCancellation(t.Context(), interrupted, interrupt)
		}()

		close(interrupted)

		err := <-done
		require.NoError(t, err, "Actual err: %v", err)
		assert.Zero(t, interruptCalled)

		// Give some time to make sure no delayed call to interrupt happens.
		time.Sleep(50 * time.Millisecond)
		assert.Zero(t, interruptCalled)
	})
}

func TestUnit_WatchForCancellation_WhenSIGINTReceived_ExpectInterruptToBeCalled(t *testing.T) {
	// Case where we need to wait for a signal
	if *waitForInterruption {
		runWatchForCancellationUntilSignal(t)
		return
	}

	// Body of the test: we need to start the part above as a subprocess
	// and send a SIGINT to the corresponding child process
	args := []string{
		"-test.v",
		"-test.run=^TestUnit_WatchForCancellation_WhenSIGINTReceived_ExpectInterruptToBeCalled$",
		"-wait_for_interruption",
	}

	cmd := exec.Command(os.Args[0], args...)

	// Voluntarily ignoring errors: the subprocess sometimes does not return
	// any error and sometimes an error status.
	output, _ := cmd.Output()

	actual := formatTestOutput(output)

	expected := []string{
		"interrupt called",
	}
	assert.ElementsMatch(t, expected, actual)
}

func runWatchForCancellationUntilSignal(t *testing.T) {
	t.Helper()

	interrupted := make(chan struct{})

	interrupt := func() error {
		fmt.Println("interrupt called")
		return nil
	}

	go func() {
		time.AfterFunc(100*time.Millisecond, func() {
			err := syscall.Kill(syscall.Getpid(), syscall.SIGINT)
			if err != nil {
				panic(err)
			}
		})
	}()

	err := watchForCancellation(t.Context(), interrupted, interrupt)
	require.NoError(t, err, "Actual err: %v", err)
}
