package process

import (
	"flag"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// testTimeout only bounds how long a broken test hangs: passing tests never wait for it.
const testTimeout = time.Second

// https://github.com/golang/go/blob/master/src/os/signal/signal_test.go#L713
var (
	waitForInterruption = flag.Bool(
		"wait_for_interruption",
		false,
		"if true, the test will emit an interrupt signal to itself",
	)

	testFrameworkPrefixes = []string{
		"=== RUN",
		"--- PASS",
		"--- FAIL",
		"PASS",
		"FAIL",
		// Happens in CI when the coverage is on"
		"coverage:",
	}
)

func shouldBeFiltered(line string) bool {
	for _, prefix := range testFrameworkPrefixes {
		if strings.HasPrefix(line, prefix) {
			return true
		}
	}

	return line == ""
}

func formatTestOutput(output []byte) []string {
	var out []string

	for _, line := range strings.Split(string(output), "\n") {
		// Uncomment for easy debug in the CI or locally.
		// fmt.Printf("line: \"%s\" -> %t\n", line, shouldBeFiltered(line))
		if !shouldBeFiltered(line) {
			out = append(out, line)
		}
	}

	return out
}

// controlledProcess lets a test decide when each step of a process happens.
//
// Timers can't guarantee that something happens while Run is still going,
// such as cancelling the context. They also make tests slow when the
// durations are generous and flaky when they are tight. Here, Run stays
// blocked until the test or Interrupt releases it. This makes it possible
// to test paths that a Run returning right away can never reach: an
// interrupt caused by cancellation, error precedence between Run and
// Interrupt, and wait() blocking while Run is still running.
type controlledProcess struct {
	// started is closed once Run executes, so the test knows it can act on
	// a process that is actually running.
	started chan struct{}
	// allowRun releases Run when closed, either by the test (natural
	// completion) or by Interrupt (requested shutdown).
	allowRun chan struct{}
	// interruptCalled is closed by Interrupt, so the test can check that
	// Interrupt was called (or not) without polling a counter.
	interruptCalled chan struct{}
	runErr          error
	interruptErr    error
	// runPanic, when set before the process starts, makes Run panic instead
	// of returning runErr once unblocked.
	runPanic any
	// runCalls is incremented before anything else in Run, so a second call
	// is still counted even though it panics on closing started.
	runCalls atomic.Int32
}

func newControlledProcess(runErr, interruptErr error) *controlledProcess {
	return &controlledProcess{
		started:         make(chan struct{}),
		allowRun:        make(chan struct{}),
		interruptCalled: make(chan struct{}),
		runErr:          runErr,
		interruptErr:    interruptErr,
	}
}

// Process may only be run once: its Run and Interrupt close channels that
// can't be closed twice.
func (c *controlledProcess) Process() Process {
	return Process{
		Run: func() error {
			c.runCalls.Add(1)
			close(c.started)
			<-c.allowRun
			if c.runPanic != nil {
				panic(c.runPanic)
			}
			return c.runErr
		},
		Interrupt: func() error {
			close(c.interruptCalled)
			c.unblockRun()
			return c.interruptErr
		},
	}
}

// unblockRun lets Run return, as if the process finished its work.
func (c *controlledProcess) unblockRun() {
	close(c.allowRun)
}

func requireClosed(t *testing.T, ch <-chan struct{}, msg string) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(testTimeout):
		t.Fatal(msg)
	}
}
