package logger

import (
	"go.uber.org/zap"
)

// NewLogger creates and returns a logger with the provided service name.
func NewLogger(serviceName string) (*zap.Logger, error) {
	config := zap.NewProductionConfig()
	config.DisableStacktrace = true
	var logger *zap.Logger
	var err error

	logger, err = config.Build()
	if err != nil {
		return nil, err
	}

	logger = logger.With(zap.String("service", serviceName))

	return logger, nil
}
