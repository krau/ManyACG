package log

import (
	"sync"
	"sync/atomic"
)

type Logger interface {
	Debug(msg any, args ...any)
	Info(msg any, args ...any)
	Warn(msg any, args ...any)
	Error(msg any, args ...any)
	Fatal(msg any, args ...any)
}

var (
	defaultLogger     atomic.Pointer[Logger]
	defaultLoggerOnce sync.Once
)

func Default() Logger {
	dl := defaultLogger.Load()
	if dl == nil {
		defaultLoggerOnce.Do(func() {
			var logger Logger = CharmLog("logs/manyacg.log")
			defaultLogger.CompareAndSwap(nil, &logger)
		})
		dl = defaultLogger.Load()
	}
	return *dl
}

func SetDefault(logger Logger) {
	defaultLogger.Store(&logger)
}

func Debug(msg any, args ...any) {
	Default().Debug(msg, args...)
}

func Info(msg any, args ...any) {
	Default().Info(msg, args...)
}

func Warn(msg any, args ...any) {
	Default().Warn(msg, args...)
}

func Error(msg any, args ...any) {
	Default().Error(msg, args...)
}

func Fatal(msg any, args ...any) {
	Default().Fatal(msg, args...)
}
