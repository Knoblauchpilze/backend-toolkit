package logger

import (
	"io"
	"sync"
)

type safeConsoleWriter struct {
	lock   sync.Mutex
	writer io.Writer
}

func newSafeConsoleWriter(out io.Writer) io.Writer {
	return &safeConsoleWriter{
		writer: out,
	}
}

func (s *safeConsoleWriter) Write(p []byte) (n int, err error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.writer.Write(p)
}
