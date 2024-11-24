package logger

import (
	"github.com/labstack/gommon/log"
)

type Level int

const (
	TRACE Level = 1
	DEBUG Level = 2
	INFO  Level = 3
	WARN  Level = 4
	ERROR Level = 5
)

func fromEchoLogLevel(level log.Lvl) Level {
	switch level {
	case log.DEBUG:
		return DEBUG
	case log.INFO:
		return INFO
	case log.WARN:
		return WARN
	case log.ERROR:
		return ERROR
	default:
		return TRACE
	}
}

func toEchoLogLevel(level Level) log.Lvl {
	switch level {
	case DEBUG:
		return log.DEBUG
	case INFO:
		return log.INFO
	case WARN:
		return log.WARN
	case ERROR:
		return log.ERROR
	default:
		return log.OFF
	}
}
