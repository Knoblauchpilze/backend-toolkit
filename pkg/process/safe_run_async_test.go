package process

import (
	"testing"
	"time"

	berrors "github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnit_SafeRunAsync(t *testing.T) {
	t.Run("calls run function", func(t *testing.T) {
		var called int

		proc := func() error {
			called++
			return nil
		}

		wait := SafeRunAsync(proc)
		err := <-wait

		require.NoError(t, err, "Actual err: %v", err)
		assert.Equal(t, 1, called)
	})

	t.Run("runs asynchronously", func(t *testing.T) {
		called := 0
		processStarted := make(chan struct{})
		allowProcessToComplete := make(chan struct{})

		proc := func() error {
			close(processStarted)
			<-allowProcessToComplete
			called++
			return nil
		}

		wait := SafeRunAsync(proc)
		// Wait for the process to start
		<-processStarted
		// At this point even the routine should be blocked on the wait
		// on the second channel.
		assert.Zero(t, called)
		// This releases the routine to complete
		close(allowProcessToComplete)
		err := <-wait

		require.NoError(t, err, "Actual err: %v", err)
		assert.Equal(t, 1, called)
	})

	t.Run("does not panic when function returns successfully", func(t *testing.T) {
		proc := func() error {
			return nil
		}

		wait := SafeRunAsync(proc)
		err := <-wait

		require.NoError(t, err, "Actual err: %v", err)
	})

	t.Run("returns error when function returns error", func(t *testing.T) {
		proc := func() error {
			return errSample
		}

		wait := SafeRunAsync(proc)
		err := <-wait

		assert.Equal(t, errSample, err, "Actual err: %v", err)
	})

	t.Run("recovers panic and returns error", func(t *testing.T) {
		proc := func() error {
			panic(errSample)
		}

		wait := SafeRunAsync(proc)
		err := <-wait

		assert.Equal(t, errSample, err, "Actual err: %v", err)
	})

	t.Run("recovers panic with random type and returns wrapped error", func(t *testing.T) {
		proc := func() error {
			panic(2)
		}

		wait := SafeRunAsync(proc)
		actual := <-wait

		err, ok := berrors.AsErrorWithCode(actual)
		require.True(t, ok)
		expected := &berrors.ErrorWithCode{
			Code:    errPanicRecovered,
			Message: "2",
			Cause:   nil,
		}
		assert.Equal(t, expected, err, "Actual err: %v", err)
	})
}

func TestUnit_SafeRunAsync_RunsAsync(t *testing.T) {
	proc := func() error {
		time.Sleep(100 * time.Millisecond)
		return nil
	}

	start := time.Now()
	wait := SafeRunAsync(proc)
	end := time.Now()
	actual := <-wait

	assert.Nil(t, actual, "Actual err: %v", actual)
	assert.LessOrEqual(t, end.Sub(start), 80*time.Millisecond)
}
