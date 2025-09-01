package process

import (
	"flag"
	"strings"
)

// https://github.com/golang/go/blob/master/src/os/signal/signal_test.go#L713
var (
	waitForInterruption = flag.Bool(
		"wait_for_interruption",
		false,
		"if true, the test will wait will emit a signal to itself and wait for it to be received for 5 seconds",
	)
)

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
		// Uncomment for easy debug in the CI or locally.
		// fmt.Printf("line: \"%s\" -> %t\n", line, shouldBeFiltered(line))
		if !shouldBeFiltered(line) {
			out = append(out, line)
		}
	}

	return out
}
