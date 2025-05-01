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
	// Case where we need to wait for a signal
	if *waitForInterruption {
		d := newDummyRunnable()
		runInterruptedRunnable(d)
		return
	}

	// Body of the test: we need to start the part above as a subprocess
	// and send a SIGINT to the corresponding child process
	args := []string{
		"-test.v",
		"-test.run=^TestUnit_StartWithSignalHandler_HandlesCorrectlyInterruptSignal$",
		"-wait_for_interruption",
	}

	cmd := exec.Command(os.Args[0], args...)

	// Voluntarily ignoring errors: the subprocess sometimes does not return
	// any error and sometimes an error status.
	output, _ := cmd.Output()

	actual := formatTestOutput(output)

	expected := []string{
		"start called",
		"stop called",
	}
	assert.ElementsMatch(t, expected, actual)
}

type dummyRunnable struct {
	done chan bool

	runCalled       atomic.Int32
	interruptCalled atomic.Int32
}

func newDummyRunnable() *dummyRunnable {
	return &dummyRunnable{
		done: make(chan bool, 1),
	}
}

func (d *dummyRunnable) Start() error {
	d.runCalled.Add(1)
	fmt.Println("start called")
	<-d.done
	return nil
}

func (d *dummyRunnable) Stop() error {
	d.interruptCalled.Add(1)
	fmt.Println("stop called")
	d.done <- true
	return nil
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
