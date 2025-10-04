package log

import (
	"os"

	"github.com/charmbracelet/log"
	"gopkg.in/natefinch/lumberjack.v2"
)

type charmLog struct {
	console *log.Logger
	file    *log.Logger
}

func toCharmLevel(l Level) log.Level {
	switch l {
	case LevelDebug:
		return log.DebugLevel
	case LevelInfo:
		return log.InfoLevel
	case LevelWarn:
		return log.WarnLevel
	case LevelError:
		return log.ErrorLevel
	case LevelFatal:
		return log.FatalLevel
	}
	return log.InfoLevel
}

func CharmLog(cfg Config) Logger {
	lj := &lumberjack.Logger{
		Filename:   cfg.LogFile,
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAge,
		Compress:   cfg.Compress,
	}
	// console logger
	consoleLogger := log.NewWithOptions(os.Stdout, log.Options{
		Formatter:       log.TextFormatter,
		ReportCaller:    true,
		ReportTimestamp: true,
		Level:           toCharmLevel(cfg.Level),
		CallerFormatter: log.ShortCallerFormatter,
		CallerOffset:    2,
		TimeFunction:    nowLocal,
		TimeFormat:      log.DefaultTimeFormat,
	})
	// file logger
	fileLogger := log.NewWithOptions(lj, log.Options{
		Formatter:       log.JSONFormatter,
		ReportCaller:    true,
		ReportTimestamp: true,
		Level:           toCharmLevel(cfg.FileLevel),
		CallerFormatter: log.ShortCallerFormatter,
		CallerOffset:    2,
		TimeFunction:    nowLocal,
		TimeFormat:      log.DefaultTimeFormat,
	})

	return &charmLog{
		console: consoleLogger,
		file:    fileLogger,
	}
}

func (c *charmLog) Debug(msg any, args ...any) {
	c.console.Debug(msg, args...)
	c.file.Debug(msg, args...)
}

func (c *charmLog) Info(msg any, args ...any) {
	c.console.Info(msg, args...)
	c.file.Info(msg, args...)
}

func (c *charmLog) Warn(msg any, args ...any) {
	c.console.Warn(msg, args...)
	c.file.Warn(msg, args...)
}

func (c *charmLog) Error(msg any, args ...any) {
	c.console.Error(msg, args...)
	c.file.Error(msg, args...)
}

func (c *charmLog) Fatal(msg any, args ...any) {
	c.console.Fatal(msg, args...)
	c.file.Fatal(msg, args...)
}

// Debugf implements Logger.
func (c *charmLog) Debugf(format string, args ...any) {
	c.console.Debugf(format, args...)
	c.file.Debugf(format, args...)
}

// Errorf implements Logger.
func (c *charmLog) Errorf(format string, args ...any) {
	c.console.Errorf(format, args...)
	c.file.Errorf(format, args...)
}

// Fatalf implements Logger.
func (c *charmLog) Fatalf(format string, args ...any) {
	c.console.Fatalf(format, args...)
	c.file.Fatalf(format, args...)
}

// Infof implements Logger.
func (c *charmLog) Infof(format string, args ...any) {
	c.console.Infof(format, args...)
	c.file.Infof(format, args...)
}

// Warnf implements Logger.
func (c *charmLog) Warnf(format string, args ...any) {
	c.console.Warnf(format, args...)
	c.file.Warnf(format, args...)
}
