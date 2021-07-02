package logger

import (
	"context"
	"os"

	log "github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

type logCtxKeyType struct{}
type Fields log.Fields

var logCtxKey logCtxKeyType

var defaultLogEntry *log.Entry

func init() {
	if defaultLogEntry == nil {
		logger := log.New()
		logger.SetFormatter(&log.JSONFormatter{})
		logger.SetOutput(os.Stdout)
		logger.SetLevel(log.DebugLevel)
		defaultLogEntry = log.NewEntry(logger)
	}
}

// usage: please refer to https://github.com/natefinch/lumberjack#type-logger
func GenerateRollingLogEntry(rollingLogger *lumberjack.Logger) *log.Entry {
	logger := log.New()
	logger.SetFormatter(&log.JSONFormatter{})
	logger.SetOutput(rollingLogger)
	logger.SetLevel(log.DebugLevel)
	return log.NewEntry(logger)
}

func SetDefaultLogEntry(entry *log.Entry) {
	defaultLogEntry = entry
}

func NewContext(ctx context.Context, fields Fields) context.Context {
	return context.WithValue(ctx, logCtxKey, WithContext(ctx).WithFields(log.Fields(fields)))
}

func WithContext(ctx context.Context) *log.Entry {
	if ctx == nil {
		return defaultLogEntry
	}
	le, ok := ctx.Value(logCtxKey).(*log.Entry)
	if ok {
		return le
	} else {
		return defaultLogEntry
	}
}
