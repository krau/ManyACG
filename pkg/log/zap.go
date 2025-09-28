package log

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type zapLogger struct {
	logger *zap.Logger
}

func ZapLog(cfg Config) Logger {
	cfg.ApplyDefaults()
	level := zapcore.DebugLevel
	consoleEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	consoleWriter := zapcore.Lock(os.Stdout)

	fileWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   cfg.LogFile,
		MaxSize:    cfg.MaxSize, // megabytes
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAge, // days
		Compress:   cfg.Compress,
	})
	fileEncoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())

	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, consoleWriter, level),
		zapcore.NewCore(fileEncoder, fileWriter, level),
	)

	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	return &zapLogger{logger: logger}
}

func (z *zapLogger) Debug(msg any, args ...any) {
	z.logger.Sugar().Debugw(toString(msg), args...)
}

func (z *zapLogger) Info(msg any, args ...any) {
	z.logger.Sugar().Infow(toString(msg), args...)
}

func (z *zapLogger) Warn(msg any, args ...any) {
	z.logger.Sugar().Warnw(toString(msg), args...)
}

func (z *zapLogger) Error(msg any, args ...any) {
	z.logger.Sugar().Errorw(toString(msg), args...)
}

func (z *zapLogger) Fatal(msg any, args ...any) {
	z.logger.Sugar().Fatalw(toString(msg), args...)
}
