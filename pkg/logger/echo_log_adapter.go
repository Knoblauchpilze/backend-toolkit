package logger

import (
	"context"
	"io"
	"log/slog"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/labstack/gommon/log"
)

type slogLoggerAdapter struct {
	log Logger
}

func Wrap(log Logger) *slog.Logger {
	return slog.New(&slogLoggerAdapter{log})
}

func Duplicate(log *slog.Logger) (*slog.Logger, error) {
	handler := log.Handler()
	adapter, ok := handler.(*slogLoggerAdapter)
	if !ok {
		return log, errors.NewCode(UnsupportedLogger)
	}

	copy := &slogLoggerAdapter{
		log: Clone(adapter.log),
	}

	return slog.New(copy), nil
}

func (la *slogLoggerAdapter) Enabled(context.Context, slog.Level) bool {
	return true
}

func (la *slogLoggerAdapter) Handle(context.Context, slog.Record) error { return nil }

func (la *slogLoggerAdapter) WithAttrs(attrs []slog.Attr) slog.Handler {
	return nil
}

func (la *slogLoggerAdapter) WithGroup(name string) slog.Handler {
	return nil
}

func (la *slogLoggerAdapter) Output() io.Writer {
	return la.log.Output()
}

func (la *slogLoggerAdapter) SetOutput(w io.Writer) {
	la.log.SetOutput(w)
}

func (la *slogLoggerAdapter) Prefix() string {
	return la.log.Prefix()
}

func (la *slogLoggerAdapter) SetPrefix(p string) {
	la.log.SetPrefix(p)
}

func (la *slogLoggerAdapter) Level() log.Lvl {
	return toEchoLogLevel(la.log.Level())
}

func (la *slogLoggerAdapter) SetLevel(v log.Lvl) {
	la.log.SetLevel(fromEchoLogLevel(v))
}

func (la *slogLoggerAdapter) SetHeader(h string) {
	la.log.SetHeader(h)
}

// func (la *echoLoggerAdapter) Print(i ...interface{}) {
// 	// https://github.com/labstack/gommon/blob/2888b9ce44ed86f3cb956f95becc724d255f0a33/log/log.go#L360
// 	la.Printf(fmt.Sprint(i...))
// }

// func (la *echoLoggerAdapter) Printf(format string, args ...interface{}) {
// 	la.log.Printf(format, args...)
// }

// func (la *echoLoggerAdapter) Printj(j log.JSON) {
// 	// https://github.com/labstack/gommon/blob/2888b9ce44ed86f3cb956f95becc724d255f0a33/log/log.go#L362
// 	// Voluntarily ignore errors
// 	data, _ := json.Marshal(j)
// 	la.Printf(string(data))
// }

// func (la *echoLoggerAdapter) Debug(i ...interface{}) {
// 	la.Debugf(fmt.Sprint(i...))
// }

// func (la *echoLoggerAdapter) Debugf(format string, args ...interface{}) {
// 	la.log.Debugf(format, args...)
// }

// func (la *echoLoggerAdapter) Debugj(j log.JSON) {
// 	data, _ := json.Marshal(j)
// 	la.Debugf(string(data))
// }

// func (la *echoLoggerAdapter) Info(i ...interface{}) {
// 	la.Infof(fmt.Sprint(i...))
// }

// func (la *echoLoggerAdapter) Infof(format string, args ...interface{}) {
// 	la.log.Infof(format, args...)
// }

// func (la *echoLoggerAdapter) Infoj(j log.JSON) {
// 	data, _ := json.Marshal(j)
// 	la.Infof(string(data))
// }

// func (la *echoLoggerAdapter) Warn(i ...interface{}) {
// 	la.Infof(fmt.Sprint(i...))
// }

// func (la *echoLoggerAdapter) Warnf(format string, args ...interface{}) {
// 	la.log.Warnf(format, args...)
// }

// func (la *echoLoggerAdapter) Warnj(j log.JSON) {
// 	data, _ := json.Marshal(j)
// 	la.Warnf(string(data))
// }

// func (la *echoLoggerAdapter) Error(i ...interface{}) {
// 	la.Errorf(fmt.Sprint(i...))
// }

// func (la *echoLoggerAdapter) Errorf(format string, args ...interface{}) {
// 	la.log.Errorf(format, args...)
// }

// func (la *echoLoggerAdapter) Errorj(j log.JSON) {
// 	data, _ := json.Marshal(j)
// 	la.Errorf(string(data))
// }

// func (la *echoLoggerAdapter) Fatal(i ...interface{}) {
// 	la.Fatalf(fmt.Sprint(i...))
// }

// func (la *echoLoggerAdapter) Fatalj(j log.JSON) {
// 	data, _ := json.Marshal(j)
// 	la.Fatalf(string(data))
// }

// func (la *echoLoggerAdapter) Fatalf(format string, args ...interface{}) {
// 	la.log.Errorf(format, args...)
// }

// func (la *echoLoggerAdapter) Panic(i ...interface{}) {
// 	la.Panicf(fmt.Sprint(i...))
// }

// func (la *echoLoggerAdapter) Panicj(j log.JSON) {
// 	data, _ := json.Marshal(j)
// 	la.Panicf(string(data))
// }

// func (la *echoLoggerAdapter) Panicf(format string, args ...interface{}) {
// 	la.log.Errorf(format, args...)
// }
