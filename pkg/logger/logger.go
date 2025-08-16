package logger

import (
	"fmt"
	"noteApp/internal/config"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	*zap.Logger
}

func Init(cfg config.LoggerCfg) (*Logger, func(), error) {
	if cfg.Level == "" {
		cfg.Level = "info"
	}
	if cfg.Encoding == "" {
		cfg.Encoding = "json"
	}

	zapConfig, err := setConfig(cfg)
	if err != nil {
		return nil, nil, err
	}

	logs, err := zapConfig.Build(
		zap.AddStacktrace(zapcore.PanicLevel),
		zap.AddCallerSkip(1),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build logger: %w", err)
	}

	s := func() {
		_ = logs.Sync()
	}

	return &Logger{logs}, s, nil
}

func setConfig(cfg config.LoggerCfg) (zap.Config, error) {
	level, err := zap.ParseAtomicLevel(cfg.Level)
	if err != nil {
		return zap.Config{}, fmt.Errorf("invalid log level '%s': %w", cfg.Level, err)
	}

	encoderCfg := zap.NewProductionEncoderConfig()
	if cfg.Development {
		encoderCfg = zap.NewDevelopmentEncoderConfig()
	}

	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderCfg.EncodeLevel = zapcore.CapitalLevelEncoder

	if cfg.Encoding != "json" && cfg.Encoding != "console" {
		return zap.Config{}, fmt.Errorf("invalid encoding '%s': must be 'json' or 'console'", cfg.Encoding)
	}

	outputPaths, err := setupOutputs(cfg.OutputPaths)
	if err != nil {
		return zap.Config{}, fmt.Errorf("invalid output paths: %v", err)
	}

	errorOutputPaths, err := setupOutputs(cfg.ErrorOutputPaths)
	if err != nil {
		return zap.Config{}, fmt.Errorf("invalid error output paths: %v", err)
	}

	return zap.Config{
		Level:             level,
		Development:       cfg.Development,
		DisableCaller:     cfg.DisableCaller,
		DisableStacktrace: cfg.DisableStacktrace,
		Encoding:          cfg.Encoding,
		EncoderConfig:     encoderCfg,
		OutputPaths:       outputPaths,
		ErrorOutputPaths:  errorOutputPaths,
	}, nil
}

func setupOutputs(paths []string) ([]string, error) {
	if len(paths) == 0 {
		return nil, fmt.Errorf("output paths cannot be empty")
	}
	visited := make(map[string]struct{})

	result := []string{}

	for _, path := range paths {
		if _, ok := visited[path]; ok {
			continue
		}

		visited[path] = struct{}{}

		if path == "stdout" || path == "stderr" {
			result = append(result, path)
			continue
		}

		file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return nil, fmt.Errorf("cannot open log file %s: %w", path, err)
		} else {
			result = append(result, path)
		}

		_ = file.Close()
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no valid output paths provided")
	}

	return result, nil
}

func LoggerForTest() *Logger {
	return &Logger{zap.NewNop()}
}

func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.Logger.Debug(msg, fields...)
}

func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.Logger.Info(msg, fields...)
}

func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.Logger.Warn(msg, fields...)
}

func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.Logger.Error(msg, fields...)
}

func (l *Logger) Panic(msg string, fields ...zap.Field) {
	l.Logger.Panic(msg, fields...)
}
