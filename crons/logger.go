package crons

import (
	"github.com/robfig/cron/v3"
	"github.com/weitrue/kit/logger/meta"
	"go.uber.org/zap"
)

type Logger struct {
	logger *zap.Logger
	fields []meta.Field
}

// NewLogger 新建日志记录器
func NewLogger(logger *zap.Logger) cron.Logger {
	return &Logger{
		logger: logger,
		fields: make([]meta.Field, 0),
	}
}

func (l *Logger) WithField(key string, val any) *Logger {
	l.fields = append(l.fields, meta.NewField(key, val))
	return l
}

func (l *Logger) Info(msg string, keysAndValues ...any) {
	field := l.fields
	for i := 0; i+1 < len(keysAndValues); {
		field = append(field, meta.NewField(keysAndValues[i].(string), keysAndValues[i+1]))
		i = i + 2
	}

	l.logger.WithOptions(zap.AddCallerSkip(1)).Info(msg, meta.WrapZapMeta(nil, l.fields...)...)
}

func (l *Logger) Error(err error, msg string, keysAndValues ...any) {
	field := l.fields
	for i := 0; i+1 < len(keysAndValues); {
		field = append(field, meta.NewField(keysAndValues[i].(string), keysAndValues[i+1]))
		i = i + 2
	}

	l.logger.WithOptions(zap.AddCallerSkip(1)).Error(msg, meta.WrapZapMeta(err, l.fields...)...)
}
