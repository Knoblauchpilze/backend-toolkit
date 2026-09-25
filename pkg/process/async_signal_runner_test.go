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

func TestUnit_AsyncStartWithSignalHandler_WhenSIGINTReceived_ExpectCloseToBeCalled(t *testing.T) {
	// Case where we need to wait for a signal
	if *waitForInterruption {
		runInterruptedProcess(nil)
		return
	}

	// Body of the test: we need to start the part above as a subprocess
	// and send a SIGINT to the corresponding child process
	args := []string{
		"-test.v",
		"-test.run=^TestUnit_AsyncStartWithSignalHandler_WhenSIGINTReceived_ExpectCloseToBeCalled$",
		"-wait_for_interruption",
	}

	cmd := exec.Command(os.Args[0], args...)

	// Voluntarily ignoring errors: the subprocess sometimes does not return
	// any error and sometimes an error status.
	output, _ := cmd.Output()

	actual := formatTestOutput(output)

	expected := []string{
		"start called",
		"interrupt called",
		"stopping process",
	}
	assert.ElementsMatch(t, expected, actual)
}

func TestUnit_AsyncStartWithSignalHandler_ExpectInterruptErrorToBeReturned(t *testing.T) {
	if *waitForInterruption {
		runInterruptedProcess(errSample)
		return
	}

	args := []string{
		"-test.v",
		"-test.run=^TestUnit_AsyncStartWithSignalHandler_ExpectInterruptErrorToBeReturned$",
		"-wait_for_interruption",
	}

	cmd := exec.Command(os.Args[0], args...)

	output, _ := cmd.Output()

	actual := formatTestOutput(output)

	expected := []string{
		"start called",
		"interrupt called",
		"stopping process",
		"error waiting for process: sample error",
	}
	assert.ElementsMatch(t, expected, actual)
}

func runInterruptedProcess(interruptError error) {
	stop := make(chan bool, 2)

	process := Process{
		Run: func() error {
			fmt.Println("start called")
			ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
			defer cancel()
			select {
			case <-ctx.Done():
				fmt.Println("process reached timeout")
			case <-stop:
				fmt.Println("stopping process")
			}
			return nil
		},
		Interrupt: func() error {
			fmt.Println("interrupt called")
			stop <- true
			return interruptError
		},
	}

	go func() {
		time.AfterFunc(100*time.Millisecond, func() {
			err := syscall.Kill(syscall.Getpid(), syscall.SIGINT)
			if err != nil {
				panic(err)
			}
		})
	}()

	wait, err := AsyncStartWithSignalHandler(context.Background(), process)
	if err != nil {
		fmt.Println("error starting process:", err)
	}

	err = wait()
	if err != nil {
		fmt.Println("error waiting for process:", err)
	}
}
