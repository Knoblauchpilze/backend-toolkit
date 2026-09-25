package process

import (
	"errors"
	"testing"

	berrors "github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errSample = errors.New("sample error")

func TestUnit_SafeRunSync(t *testing.T) {
	t.Run("calls run function", func(t *testing.T) {
		var called int

		proc := func() error {
			called++
			return nil
		}

		err := SafeRunSync(proc)

		require.NoError(t, err, "Actual err: %v", err)
		assert.Equal(t, 1, called)
	})

	t.Run("does not panic when function returns successfully", func(t *testing.T) {
		proc := func() error {
			return nil
		}

		err := SafeRunSync(proc)

		require.NoError(t, err, "Actual err: %v", err)
	})

	t.Run("returns error when function returns error", func(t *testing.T) {
		proc := func() error {
			return errSample
		}

		err := SafeRunSync(proc)

		assert.Equal(t, errSample, err, "Actual err: %v", err)
	})

	t.Run("recovers panic and returns error", func(t *testing.T) {
		proc := func() error {
			panic(errSample)
		}

		err := SafeRunSync(proc)

		assert.Equal(t, errSample, err, "Actual err: %v", err)
	})

	t.Run("recovers panic with random type and returns wrapped error", func(t *testing.T) {
		proc := func() error {
			panic(2)
		}

		actual := SafeRunSync(proc)

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
