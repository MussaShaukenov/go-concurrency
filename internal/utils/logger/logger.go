package logger

import (
	"strings"

	"github.com/MussaShaukenov/go-concurrency/internal/config"
	"go.uber.org/zap"
)

func NewLogger(cfg *config.Config) (*zap.SugaredLogger, error) {
	loggerCfg := zap.NewDevelopmentConfig()

	// Set level from loggerCfg
	switch strings.ToLower(cfg.Logging.Level) {
	case "debug":
		loggerCfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		loggerCfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		loggerCfg.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		loggerCfg.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	}

	// Set output path from loggerCfg
	if cfg.Logging.Output == "stderr" {
		loggerCfg.OutputPaths = []string{"stderr"}
	} else if cfg.Logging.Output == "stdout" {
		loggerCfg.OutputPaths = []string{"stdout"}
	} else {
		// It's a file path
		loggerCfg.OutputPaths = []string{cfg.Logging.Output}
	}

	logger, err := loggerCfg.Build()
	if err != nil {
		return nil, err
	}

	return logger.Sugar(), nil
}
