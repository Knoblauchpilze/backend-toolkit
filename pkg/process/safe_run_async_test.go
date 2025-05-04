package process

import (
	"testing"
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestUnit_SafeRunAsync_CallsProcess(t *testing.T) {
	var called int

	proc := func() error {
		called++
		return nil
	}

	wait := SafeRunAsync(proc)
	actual := <-wait

	assert.Nil(t, actual, "Actual err: %v", actual)
	assert.Equal(t, 1, called)
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

func TestUnit_SafeRunAsync_ReturnWithError(t *testing.T) {
	proc := func() error {
		return errSample
	}

	wait := SafeRunAsync(proc)
	actual := <-wait

	assert.Equal(t, errSample, actual, "Actual err: %v", actual)
}

func TestUnit_SafeRunAsync_PanicWithError(t *testing.T) {
	proc := func() error {
		panic(errSample)
	}

	wait := SafeRunAsync(proc)
	actual := <-wait

	assert.Equal(t, errSample, actual)
}

func TestUnit_SafeRunAync_PanicWithRandomDatatype(t *testing.T) {
	proc := func() error {
		panic(2)
	}

	wait := SafeRunAsync(proc)
	actual := <-wait

	assert.Equal(t, errors.New("2"), actual)
}
