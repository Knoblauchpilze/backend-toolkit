package process

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUnit_StartWithSignalHandler_StopsWhenContextIsCancelled(t *testing.T) {
	d := newDummyRunnable()

	ctx, cancel := context.WithCancel(context.Background())

	wait, err := StartWithSignalHandler(ctx, d)
	assert.Nil(t, err, "Actual err: %v", err)

	cancel()

	err = wait()
	assert.Nil(t, err, "Actual err: %v", err)

	assert.Equal(t, int32(1), d.runCalled.Load())
	assert.Equal(t, int32(1), d.interruptCalled.Load())
}

func TestUnit_StartWithSignalHandler_HandlesCorrectlyInterruptSignal(t *testing.T) {
	if *waitForInterruption {
		d := newDummyRunnable()
		runInterruptedRunnable(d)
		return
	}

	args := []string{
		"-test.v",
		"-test.run=^TestUnit_StartWithSignalHandler_HandlesCorrectlyInterruptSignal$",
		"-wait_for_interruption",
	}

	cmd := exec.Command(os.Args[0], args...)

	output, _ := cmd.Output()

	actual := formatTestOutput(output)

	expected := []string{
		"start called",
		"stop called",
	}
	assert.ElementsMatch(t, expected, actual)
}

func TestUnit_StartWithSignalHandler_HandlesCorrectlyRunnableError(t *testing.T) {
	if *waitForInterruption {
		d := newDummyRunnableWithRunError(errSample, nil)
		runInterruptedRunnable(d)
		return
	}

	args := []string{
		"-test.v",
		"-test.run=^TestUnit_StartWithSignalHandler_HandlesCorrectlyRunnableError$",
		"-wait_for_interruption",
	}

	cmd := exec.Command(os.Args[0], args...)

	output, _ := cmd.Output()

	actual := formatTestOutput(output)

	expected := []string{
		"start called",
		"stop called",
		"error waiting for process: sample error",
	}
	assert.ElementsMatch(t, expected, actual)
}

func TestUnit_StartWithSignalHandler_HandlesCorrectlyInterruptError(t *testing.T) {
	if *waitForInterruption {
		d := newDummyRunnableWithRunError(nil, errSample)
		runInterruptedRunnable(d)
		return
	}

	args := []string{
		"-test.v",
		"-test.run=^TestUnit_StartWithSignalHandler_HandlesCorrectlyInterruptError$",
		"-wait_for_interruption",
	}

	cmd := exec.Command(os.Args[0], args...)

	output, _ := cmd.Output()

	actual := formatTestOutput(output)

	expected := []string{
		"start called",
		"stop called",
		"error waiting for process: sample error",
	}
	assert.ElementsMatch(t, expected, actual)
}

func TestUnit_StartWithSignalHandler_RunErrorOverridesInterruptError(t *testing.T) {
	if *waitForInterruption {
		errRun := fmt.Errorf("run error")
		errInterrupt := fmt.Errorf("interrupt error")
		d := newDummyRunnableWithRunError(errRun, errInterrupt)
		runInterruptedRunnable(d)
		return
	}

	args := []string{
		"-test.v",
		"-test.run=^TestUnit_StartWithSignalHandler_RunErrorOverridesInterruptError$",
		"-wait_for_interruption",
	}

	cmd := exec.Command(os.Args[0], args...)

	output, _ := cmd.Output()

	actual := formatTestOutput(output)

	expected := []string{
		"start called",
		"stop called",
		"error waiting for process: run error",
	}
	assert.ElementsMatch(t, expected, actual)
}

type dummyRunnable struct {
	stop chan bool
	done chan bool

	runCalled       atomic.Int32
	runError        error
	interruptCalled atomic.Int32
	interruptError  error
}

func newDummyRunnable() *dummyRunnable {
	return newDummyRunnableWithRunError(nil, nil)
}

func newDummyRunnableWithRunError(runError error, interruptError error) *dummyRunnable {
	return &dummyRunnable{
		stop:           make(chan bool, 1),
		done:           make(chan bool, 1),
		runError:       runError,
		interruptError: interruptError,
	}
}

func (d *dummyRunnable) Start() error {
	defer func() {
		d.done <- true
	}()

	d.runCalled.Add(1)
	fmt.Printf("start called\n")
	<-d.stop
	return d.runError
}

func (d *dummyRunnable) Stop() error {
	d.interruptCalled.Add(1)
	fmt.Printf("stop called\n")
	d.stop <- true
	<-d.done
	return d.interruptError
}

func runInterruptedRunnable(runnable Runnable) {
	go func() {
		time.AfterFunc(100*time.Millisecond, func() {
			syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		})
	}()

	wait, err := StartWithSignalHandler(context.Background(), runnable)
	if err != nil {
		fmt.Printf("error starting process: %v\n", err)
	}

	err = wait()
	if err != nil {
		fmt.Printf("error waiting for process: %v\n", err)
	}
}
