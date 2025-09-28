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

type Config struct {
	LogFile    string
	MaxSize    int
	MaxBackups int
	MaxAge     int
	Compress   bool
}

func (c *Config) ApplyDefaults() {
	if c.LogFile == "" {
		c.LogFile = "logs/manyacg.log"
	}
	if c.MaxSize == 0 {
		c.MaxSize = 10
	}
	if c.MaxBackups == 0 {
		c.MaxBackups = 10
	}
	if c.MaxAge == 0 {
		c.MaxAge = 14
	}
}

func Default() Logger {
	dl := defaultLogger.Load()
	if dl == nil {
		defaultLoggerOnce.Do(func() {
			var logger Logger = CharmLog(Config{})
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
