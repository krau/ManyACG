package log

import (
	"sync"
)

type Logger interface {
	Debug(msg any, args ...any)
	Info(msg any, args ...any)
	Warn(msg any, args ...any)
	Error(msg any, args ...any)
	Fatal(msg any, args ...any)

	Debugf(format string, args ...any)
	Infof(format string, args ...any)
	Warnf(format string, args ...any)
	Errorf(format string, args ...any)
	Fatalf(format string, args ...any)
}

type Config struct {
	Level        Level
	LogFile      string
	FileLevel    Level
	MaxSize      int
	MaxBackups   int
	MaxAge       int
	Compress     bool
	CallerOffset int
}

type Level uint

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

func (c *Config) ApplyDefaults() {
	if c.Level == 0 {
		c.Level = LevelDebug
	}
	if c.FileLevel == 0 {
		c.FileLevel = LevelError
	}
	if c.LogFile == "" {
		c.LogFile = "logs/app.log"
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
	if c.CallerOffset == 0 {
		c.CallerOffset = 2
	}
}

var (
	defaultLogger     Logger
	defaultLoggerOnce sync.Once
)

func Default() Logger {
	if defaultLogger == nil {
		defaultLoggerOnce.Do(func() {
			cfg := Config{}
			cfg.ApplyDefaults()
			defaultLogger = CharmLog(cfg)
		})
	}
	return defaultLogger
}

func New(cfg Config) Logger {
	cfg.ApplyDefaults()
	return CharmLog(cfg)
}

func SetDefault(logger Logger) {
	defaultLoggerOnce.Do(func() {
		defaultLogger = logger
	})
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

func Debugf(format string, args ...any) {
	Default().Debugf(format, args...)
}

func Infof(format string, args ...any) {
	Default().Infof(format, args...)
}

func Warnf(format string, args ...any) {
	Default().Warnf(format, args...)
}

func Errorf(format string, args ...any) {
	Default().Errorf(format, args...)
}

func Fatalf(format string, args ...any) {
	Default().Fatalf(format, args...)
}
