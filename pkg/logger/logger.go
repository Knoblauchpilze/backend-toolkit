package logger

import (
	"fmt"
	"io"
	"sync"

	"github.com/rs/zerolog"
)

type Logger interface {
	Level() Level
	SetLevel(Level)
	Prefix() string
	SetPrefix(string)
	Header() string
	SetHeader(string)

	Output() io.Writer
	SetOutput(io.Writer)

	Printf(format string, v ...interface{})
	Tracef(format string, v ...interface{})
	Debugf(format string, v ...interface{})
	Infof(format string, v ...interface{})
	Warnf(format string, v ...interface{})
	Errorf(format string, v ...interface{})
}

type loggerImpl struct {
	out    io.Writer
	logger zerolog.Logger

	lock   sync.Mutex
	level  Level
	prefix string
	header string
}

func New(out io.Writer) Logger {
	safeOutput := out
	if _, ok := safeOutput.(*safeConsoleWriter); !ok {
		safeOutput = newSafeConsoleWriter(out)
	}

	return &loggerImpl{
		out:    safeOutput,
		logger: zerolog.New(safeOutput),
		level:  DEBUG,
	}
}

func Clone(logger Logger) Logger {
	copy := New(logger.Output())
	copy.SetLevel(logger.Level())
	copy.SetPrefix(logger.Prefix())
	copy.SetHeader(logger.Header())
	return copy
}

func (l *loggerImpl) Level() Level {
	l.lock.Lock()
	defer l.lock.Unlock()

	return l.level
}

func (l *loggerImpl) SetLevel(level Level) {
	l.lock.Lock()
	defer l.lock.Unlock()

	l.level = level
}

func (l *loggerImpl) Prefix() string {
	l.lock.Lock()
	defer l.lock.Unlock()

	return l.prefix
}

func (l *loggerImpl) SetPrefix(prefix string) {
	l.lock.Lock()
	defer l.lock.Unlock()

	l.prefix = prefix
}

func (l *loggerImpl) Header() string {
	l.lock.Lock()
	defer l.lock.Unlock()

	return l.header
}

func (l *loggerImpl) SetHeader(header string) {
	l.lock.Lock()
	defer l.lock.Unlock()

	l.header = header
}

func (l *loggerImpl) Output() io.Writer {
	l.lock.Lock()
	defer l.lock.Unlock()

	return l.out
}

func (l *loggerImpl) SetOutput(out io.Writer) {
	l.lock.Lock()
	defer l.lock.Unlock()

	if _, ok := out.(*safeConsoleWriter); !ok {
		out = newSafeConsoleWriter(out)
	}

	l.out = out
	l.logger = l.logger.Output(out)
}

func (l *loggerImpl) Printf(format string, v ...interface{}) {
	l.logger.Printf(format, v...)
}

func (l *loggerImpl) Tracef(format string, v ...interface{}) {
	l.log(TRACE, format, v...)
}

func (l *loggerImpl) Debugf(format string, v ...interface{}) {
	l.log(DEBUG, format, v...)
}

func (l *loggerImpl) Infof(format string, v ...interface{}) {
	l.log(INFO, format, v...)
}

func (l *loggerImpl) Warnf(format string, v ...interface{}) {
	l.log(WARN, format, v...)
}

func (l *loggerImpl) Errorf(format string, v ...interface{}) {
	l.log(ERROR, format, v...)
}

func (l *loggerImpl) shouldPrint(level Level) bool {
	l.lock.Lock()
	defer l.lock.Unlock()

	return l.level <= level
}

func (l *loggerImpl) log(level Level, format string, args ...interface{}) {
	if !l.shouldPrint(level) {
		return
	}

	event := l.createEvent(level)
	out := l.prependPrefixAndHeader(format)
	event.Timestamp().Msgf(out, args...)
}

func (l *loggerImpl) createEvent(level Level) *zerolog.Event {
	switch level {
	case DEBUG:
		return l.logger.Debug()
	case INFO:
		return l.logger.Info()
	case WARN:
		return l.logger.Warn()
	case ERROR:
		return l.logger.Error()
	default:
		return l.logger.Trace()
	}
}

func (l *loggerImpl) prependPrefixAndHeader(in string) string {
	out := in

	l.lock.Lock()
	defer l.lock.Unlock()

	if l.prefix != "" {
		out = fmt.Sprintf("[%s] %s", l.prefix, out)
	}
	if l.header != "" {
		out = fmt.Sprintf("[%s] %s", l.header, out)
	}

	return out
}
