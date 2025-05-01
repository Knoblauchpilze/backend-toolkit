package process

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestUnit_AsyncStartWithSignalHandler_WhenProcessInvalid_ExpectError(t *testing.T) {
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
				Interrupt: func() error {
					return nil
				},
			},
		},
		{
			name: "no run func",
			process: Process{
				Run: func() error { return nil },
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {

			_, err := AsyncStartWithSignalHandler(context.Background(), testCase.process)

			assert.True(
				t,
				errors.IsErrorWithCode(err, ErrInvalidProcess),
				"Actual err: %v",
				err,
			)

		})
	}
}

func TestUnit_AsyncStartWithSignalHandler_ContextCancelled(t *testing.T) {
	process := Process{
		Run: func() error { return nil },
		Interrupt: func() error {
			return nil
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	wait, err := AsyncStartWithSignalHandler(ctx, process)
	assert.Nil(t, err, "Actual err: %v", err)

	err = wait()
	assert.Nil(t, err, "Actual err: %v", err)

}

func TestUnit_AsyncStartWithSignalHandler_ProcessCalled(t *testing.T) {
	var called int
	process := Process{
		Run: func() error {
			called++
			return nil
		},
		Interrupt: func() error {
			return nil
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	wait, err := AsyncStartWithSignalHandler(ctx, process)
	assert.Nil(t, err, "Actual err: %v", err)

	err = wait()
	assert.Equal(t, 1, called)
	assert.Nil(t, err, "Actual err: %v", err)
}

func TestUnit_AsyncStartWithSignalHandler_ReturnsProcessError(t *testing.T) {
	process := Process{
		Run: func() error {
			return errSample
		},
		Interrupt: func() error {
			return nil
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	wait, err := AsyncStartWithSignalHandler(ctx, process)
	assert.Nil(t, err, "Actual err: %v", err)

	err = wait()
	assert.Equal(t, errSample, err, "Actual err: %v", err)
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

func TestUnit_AsyncStartWithSignalHandler_WhenProcessReturnsError_ExpectWaitStopsAndReturnsError(t *testing.T) {
	process := Process{
		Run: func() error {
			return errSample
		},
		Interrupt: func() error {
			return nil
		},
	}

	wait, err := AsyncStartWithSignalHandler(context.Background(), process)
	assert.Nil(t, err, "Actual err: %v", err)

	err = wait()
	assert.Equal(t, errSample, err, "Actual err: %v", err)
}

func TestUnit_AsyncStartWithSignalHandler_WhenProcessPanics_ExpectWaitStopsAndReturnsError(t *testing.T) {
	process := Process{
		Run: func() error {
			panic(errSample)
		},
		Interrupt: func() error {
			return nil
		},
	}

	wait, err := AsyncStartWithSignalHandler(context.Background(), process)
	assert.Nil(t, err, "Actual err: %v", err)

	err = wait()
	assert.Equal(t, errSample, err, "Actual err: %v", err)
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
			syscall.Kill(syscall.Getpid(), syscall.SIGINT)
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
