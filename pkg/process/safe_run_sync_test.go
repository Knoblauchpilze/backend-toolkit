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
	t.Run("CallsProcess", func(t *testing.T) {
		var called int

		proc := func() error {
			called++
			return nil
		}

		err := SafeRunSync(proc)
		require.NoError(t, err, "Actual err: %v", err)

		assert.Equal(t, 1, called)
	})

	t.Run("NoPanic", func(t *testing.T) {
		proc := func() error {
			return nil
		}

		var err error

		run := func() {
			err = SafeRunSync(proc)
		}

		assert.NotPanics(t, run)
		require.NoError(t, err, "Actual err: %v", err)
	})

	t.Run("ReturnWithError", func(t *testing.T) {
		proc := func() error {
			return errSample
		}

		var actual error

		run := func() {
			actual = SafeRunSync(proc)
		}

		assert.NotPanics(t, run)
		assert.Equal(t, errSample, actual, "Actual err: %v", actual)
	})

	t.Run("PanicWithError", func(t *testing.T) {
		proc := func() error {
			panic(errSample)
		}

		var actual error

		run := func() {
			actual = SafeRunSync(proc)
		}

		assert.NotPanics(t, run)
		assert.Equal(t, errSample, actual)
	})

	t.Run("PanicWithRandomDatatype", func(t *testing.T) {
		proc := func() error {
			panic(2)
		}

		var actual error

		run := func() {
			actual = SafeRunSync(proc)
		}

		assert.NotPanics(t, run)
		assert.Equal(t, berrors.New("2"), actual)
	})
}
