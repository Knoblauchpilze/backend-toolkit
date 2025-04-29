package process

import (
	"context"
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
				Run: func() {},
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
		Run: func() {},
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
		Run: func() {
			called++
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

func TestUnit_AsyncStartWithSignalHandler_WhenSignalReceived_ExpectCloseToBeCalled(t *testing.T) {
	var called int
	process := Process{
		Run: func() {},
		Interrupt: func() error {
			called++
			return nil
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	wait, err := AsyncStartWithSignalHandler(ctx, process)
	assert.Nil(t, err, "Actual err: %v", err)

	err = wait()
	// TODO: This does not work because we don't receive a signal
	assert.Equal(t, 1, called)
	assert.Nil(t, err, "Actual err: %v", err)
}

func TestUnit_AsyncStartWithSignalHandler_ExpectInterruptErrorToBeReturned(t *testing.T) {
	// TODO: Write this test.
}

func TestUnit_AsyncStartWithSignalHandler_WhenProcessPanics_ExpectWaitStopsAndReturnsError(t *testing.T) {
	process := Process{
		Run: func() {
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
