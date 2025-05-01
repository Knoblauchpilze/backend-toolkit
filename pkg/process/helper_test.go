package process

import (
	"context"
	"flag"
	"fmt"
	"strings"
	"syscall"
	"time"
)

// https://github.com/golang/go/blob/master/src/os/signal/signal_test.go#L713
var (
	waitForInterruption = flag.Bool(
		"wait_for_interruption",
		false,
		"if true, the test will wait will emit a signal to itself and wait for it to be received for 5 seconds",
	)
)

func runTestInterruptedProcess(interruptError error) {
	stop := make(chan bool, 2)

	process := Process{
		Run: func() error {
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

	runInterruptedProcess(process)
}

func runInterruptedProcess(process Process) {
	go func() {
		time.AfterFunc(100*time.Millisecond, func() {
			syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		})
	}()

	wait, err := AsyncStartWithSignalHandler(context.Background(), process)
	if err != nil {
		fmt.Printf("error starting process: %v\n", err)
	}

	err = wait()
	if err != nil {
		fmt.Printf("error waiting for process: %v\n", err)
	}
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

var testFrameworkPrefixes = []string{
	"=== RUN",
	"--- PASS",
	"--- FAIL",
	"PASS",
	"FAIL",
	// Happens in CI when the coverage is on"
	"coverage:",
}

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
		fmt.Printf("line: \"%s\" -> %t\n", line, shouldBeFiltered(line))
		if !shouldBeFiltered(line) {
			out = append(out, line)
		}
	}

	return out
}
