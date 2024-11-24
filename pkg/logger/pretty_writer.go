package logger

import (
	"io"
	"time"

	"github.com/rs/zerolog"
)

func NewPrettyWriter(out io.Writer) io.Writer {
	// https://github.com/rs/zerolog?tab=readme-ov-file#pretty-logging
	return zerolog.ConsoleWriter{
		Out:          out,
		TimeFormat:   time.DateTime,
		TimeLocation: time.UTC,
	}
}
