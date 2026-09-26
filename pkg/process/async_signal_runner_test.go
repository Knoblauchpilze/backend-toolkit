package process

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnit_AsyncStartWithSignalHandler(t *testing.T) {
	t.Run("returns error for invalid process", func(t *testing.T) {
		type testCase struct {
			name    string
			process Process
		}

		testCases := []testCase{
			{
				name:    "empty process",
				process: Process{},
			},
			{
				name: "no interrupt func",
				process: Process{
					Run: func() error { return nil },
				},
			},
			{
				name: "no run func",
				process: Process{
					Interrupt: func() error {
						return nil
					},
				},
			},
		}

		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				_, err := AsyncStartWithSignalHandler(context.Background(), testCase.process)

				assert.Equal(t, ErrInvalidProcess, err, "Actual err: %v", err)
			})
		}
	})

	t.Run("calls the process function", func(t *testing.T) {
		proc := newControlledProcess(nil, nil)

		wait, err := AsyncStartWithSignalHandler(context.Background(), proc.Process())
		require.NoError(t, err, "Actual err: %v", err)

		requireClosed(t, proc.started, "run was not started before being unblocked")

		proc.unblockRun()
		err = wait()
		require.NoError(t, err, "Actual err: %v", err)

		assert.Equal(t, int32(1), proc.runCalls.Load())
	})

	t.Run("calls interrupt when context is cancelled while running", func(t *testing.T) {
		proc := newControlledProcess(nil, nil)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		wait, err := AsyncStartWithSignalHandler(ctx, proc.Process())
		require.NoError(t, err, "Actual err: %v", err)

		requireClosed(t, proc.started, "run was not started")
		cancel()
		requireClosed(t, proc.interruptCalled, "interrupt was not called")

		err = wait()
		require.NoError(t, err, "Actual err: %v", err)
	})

	t.Run("does not call interrupt when context is cancelled after completion", func(t *testing.T) {
		proc := newControlledProcess(nil, nil)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		wait, err := AsyncStartWithSignalHandler(ctx, proc.Process())
		require.NoError(t, err, "Actual err: %v", err)

		proc.unblockRun()
		err = wait()
		require.NoError(t, err, "Actual err: %v", err)

		cancel()

		select {
		case <-proc.interruptCalled:
			t.Fatal("interrupt called after natural completion")
		default:
		}
	})

	t.Run("does not call interrupt when process completes naturally", func(t *testing.T) {
		proc := newControlledProcess(nil, nil)

		wait, err := AsyncStartWithSignalHandler(context.Background(), proc.Process())
		require.NoError(t, err, "Actual err: %v", err)

		proc.unblockRun()
		err = wait()
		require.NoError(t, err, "Actual err: %v", err)

		select {
		case <-proc.interruptCalled:
			t.Fatal("interrupt called after natural completion")
		default:
		}
	})

	t.Run("returns interrupt error when interruption fails", func(t *testing.T) {
		proc := newControlledProcess(nil, errSample)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		wait, err := AsyncStartWithSignalHandler(ctx, proc.Process())
		require.NoError(t, err, "Actual err: %v", err)

		requireClosed(t, proc.started, "run was not started")
		cancel()

		err = wait()
		assert.Equal(t, errSample, err, "Actual err: %v", err)
	})

	t.Run("forwards processing error to the caller", func(t *testing.T) {
		proc := newControlledProcess(errSample, nil)

		wait, err := AsyncStartWithSignalHandler(context.Background(), proc.Process())
		require.NoError(t, err, "Actual err: %v", err)

		proc.unblockRun()
		err = wait()
		assert.Equal(t, errSample, err, "Actual err: %v", err)
	})

	t.Run("recovers from a panic in the process function", func(t *testing.T) {
		proc := newControlledProcess(nil, nil)
		proc.runPanic = errSample

		wait, err := AsyncStartWithSignalHandler(context.Background(), proc.Process())
		require.NoError(t, err, "Actual err: %v", err)

		proc.unblockRun()
		err = wait()
		assert.Equal(t, errSample, err, "Actual err: %v", err)
	})
}

func TestUnit_AsyncStartWithSignalHandler_WhenSIGINTReceived_ExpectProcessExitsCleanly(t *testing.T) {
	if *waitForInterruption {
		proc := newControlledProcess(nil, nil)
		wait, err := AsyncStartWithSignalHandler(context.Background(), proc.Process())
		require.NoError(t, err, "Actual err: %v", err)
		requireClosed(t, proc.started, "run was not started")

		err = wait()
		requireClosed(t, proc.interruptCalled, "interrupt was not called")
		require.NoError(t, err, "Actual err: %v", err)
		return
	}

	// Body of the test: we need to start the part above as a subprocess and send
	// a SIGINT to the corresponding child process.

	// The 5 second timeout should be enough to allow the child process to pick up
	// the SIGINT and terminate.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(
		ctx,
		os.Args[0],
		"-test.run=^TestUnit_AsyncStartWithSignalHandler_WhenSIGINTReceived_ExpectProcessExitsCleanly$",
		"-wait_for_interruption",
	)
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	err := cmd.Start()
	require.NoError(t, err, "Actual err: %v", err)

	// This delay assumes the child's signal watcher registers before SIGINT arrives.
	time.Sleep(250 * time.Millisecond)
	signalErr := cmd.Process.Signal(syscall.SIGINT)
	waitErr := cmd.Wait()
	require.NoError(t, signalErr, "Actual err: %v, output:\n%s", signalErr, output.String())
	require.NoError(t, waitErr, "Actual err: %v, output:\n%s", waitErr, output.String())
}
