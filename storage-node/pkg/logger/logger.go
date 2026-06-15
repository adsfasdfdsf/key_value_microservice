package logger

import (
	"context"
	"go.uber.org/zap"
)

const (
	LoggerKey   = "logger"
	RequestId   = "requestId"
	ServiceName = "storage_node"
)

type Logger interface {
	Info(ctx context.Context, msg string, fields ...zap.Field)
	Error(ctx context.Context, msg string, fields ...zap.Field)
	Warn(ctx context.Context, msg string, fields ...zap.Field)
	Debug(ctx context.Context, msg string, fields ...zap.Field)
}

type logger struct {
	serviceName string
	logger      *zap.Logger
}

func (l logger) Debug(ctx context.Context, msg string, fields ...zap.Field) {
	fields = append(fields, zap.String(ServiceName, l.serviceName))

	if ctx.Value(RequestId) != nil {
		fields = append(fields, zap.String(RequestId, ctx.Value(RequestId).(string)))
	}

	l.logger.Debug(msg, fields...)
}

func (l logger) Info(ctx context.Context, msg string, fields ...zap.Field) {
	fields = append(fields, zap.String(ServiceName, l.serviceName))

	if ctx.Value(RequestId) != nil {
		fields = append(fields, zap.String(RequestId, ctx.Value(RequestId).(string)))
	}

	l.logger.Info(msg, fields...)
}

func (l logger) Warn(ctx context.Context, msg string, fields ...zap.Field) {
	fields = append(fields, zap.String(ServiceName, l.serviceName))

	if ctx.Value(RequestId) != nil {
		fields = append(fields, zap.String(RequestId, ctx.Value(RequestId).(string)))
	}

	l.logger.Warn(msg, fields...)
}

func (l logger) Error(ctx context.Context, msg string, fields ...zap.Field) {
	fields = append(fields, zap.String(ServiceName, l.serviceName))

	if ctx.Value(RequestId) != nil {
		fields = append(fields, zap.String(RequestId, ctx.Value(RequestId).(string)))
	}

	l.logger.Error(msg, fields...)
}

func New(serviceName string) Logger {
	zapLogger, _ := zap.NewDevelopment()
	defer zapLogger.Sync()
	return &logger{
		serviceName: serviceName,
		logger:      zapLogger,
	}
}

func GetLogger(ctx context.Context) Logger {
	return ctx.Value(LoggerKey).(Logger)
}