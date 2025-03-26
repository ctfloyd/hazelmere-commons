package hz_logger

import (
	"context"
	"strings"
)

type Logger interface {
	Trace(context.Context, string)
	TraceArgs(context.Context, string, ...any)
	Debug(context.Context, string)
	DebugArgs(context.Context, string, ...any)
	Info(context.Context, string)
	InfoArgs(context.Context, string, ...any)
	Warn(context.Context, string)
	WarnArgs(context.Context, string, ...any)
	Error(context.Context, string)
	ErrorArgs(context.Context, string, ...any)
}

type LogLevel int

const (
	LogLevelTrace LogLevel = iota
	LogLevelDebug
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

func LogLevelFromString(level string) LogLevel {
	level = strings.ToLower(level)
	if level == "trace" {
		return LogLevelTrace
	} else if level == "debug" {
		return LogLevelDebug
	} else if level == "info" {
		return LogLevelInfo
	} else if level == "warn" {
		return LogLevelWarn
	} else if level == "hz_service_error" {
		return LogLevelError
	} else {
		return LogLevelInfo
	}
}
