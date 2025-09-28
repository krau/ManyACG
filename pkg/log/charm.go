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

func CharmLog(cfg Config) Logger {
	cfg.ApplyDefaults()
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
		Level:           log.DebugLevel,
		CallerFormatter: log.ShortCallerFormatter,
		CallerOffset:    1,
		TimeFunction:    nowLocal,
		TimeFormat:      log.DefaultTimeFormat,
	})
	// file logger
	fileLogger := log.NewWithOptions(lj, log.Options{
		Formatter:       log.JSONFormatter,
		ReportCaller:    true,
		ReportTimestamp: true,
		Level:           log.DebugLevel,
		CallerFormatter: log.ShortCallerFormatter,
		CallerOffset:    1,
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
