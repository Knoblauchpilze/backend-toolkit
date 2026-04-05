package logger

import (
	"io"
	"log/slog"

	"github.com/rs/zerolog"
)

func New(out io.Writer) *slog.Logger {
	safeOutput := out
	if _, ok := safeOutput.(*safeConsoleWriter); !ok {
		safeOutput = newSafeConsoleWriter(out)
	}

	zlog := zerolog.New(NewPrettyWriter(safeOutput))
	handler := zerolog.NewSlogHandler(zlog)

	return slog.New(handler)
}

func NewWithLevel(out io.Writer, level zerolog.Level) *slog.Logger {
	safeOutput := out
	if _, ok := safeOutput.(*safeConsoleWriter); !ok {
		safeOutput = newSafeConsoleWriter(out)
	}

	zlog := zerolog.New(NewPrettyWriter(safeOutput)).Level(level)
	handler := zerolog.NewSlogHandler(zlog)

	return slog.New(handler)
}
