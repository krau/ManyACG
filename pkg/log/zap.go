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

func toZapLevel(l Level) zapcore.Level {
	switch l {
	case LevelDebug:
		return zapcore.DebugLevel
	case LevelInfo:
		return zapcore.InfoLevel
	case LevelWarn:
		return zapcore.WarnLevel
	case LevelError:
		return zapcore.ErrorLevel
	case LevelFatal:
		return zapcore.FatalLevel
	}
	return zapcore.InfoLevel
}

func ZapLog(cfg Config) Logger {
	consoleLevel := toZapLevel(cfg.Level)
	fileLevel := toZapLevel(cfg.FileLevel)
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
		zapcore.NewCore(consoleEncoder, consoleWriter, consoleLevel),
		zapcore.NewCore(fileEncoder, fileWriter, fileLevel),
	)

	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(cfg.CallerOffset))
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

// Debugf implements Logger.
func (z *zapLogger) Debugf(format string, args ...any) {
	z.logger.Sugar().Debugf(format, args...)
}

// Errorf implements Logger.
func (z *zapLogger) Errorf(format string, args ...any) {
	z.logger.Sugar().Errorf(format, args...)
}

// Fatalf implements Logger.
func (z *zapLogger) Fatalf(format string, args ...any) {
	z.logger.Sugar().Fatalf(format, args...)
}

// Infof implements Logger.
func (z *zapLogger) Infof(format string, args ...any) {
	z.logger.Sugar().Infof(format, args...)
}

// Warnf implements Logger.
func (z *zapLogger) Warnf(format string, args ...any) {
	z.logger.Sugar().Warnf(format, args...)
}
