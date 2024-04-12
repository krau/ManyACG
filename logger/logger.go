package logger

import (
	"ManyACG-Bot/config"

	"github.com/gookit/slog"
	"github.com/gookit/slog/handler"
	"github.com/gookit/slog/rotatefile"
)

var Logger *slog.Logger

func init() {
	slog.DefaultChannelName = "ManyACG-Bot"
	newLogger := slog.New()
	defer newLogger.Flush()
	logLevel := slog.LevelByName(config.Cfg.Log.Level)
	logFilePath := config.Cfg.Log.FilePath
	logBackupNum := config.Cfg.Log.BackupNum
	var logLevels []slog.Level
	for _, level := range slog.AllLevels {
		if level <= logLevel {
			logLevels = append(logLevels, level)
		}
	}
	consoleH := handler.NewConsoleHandler(logLevels)
	fileH, err := handler.NewTimeRotateFile(
		logFilePath,
		rotatefile.EveryDay,
		handler.WithLogLevels(slog.AllLevels),
		handler.WithBackupNum(logBackupNum),
	)
	if err != nil {
		panic(err)
	}
	newLogger.AddHandlers(consoleH, fileH)
	Logger = newLogger
}
