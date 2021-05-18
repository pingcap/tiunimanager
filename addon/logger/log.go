package logger

import (
	"context"
	"os"

	log "github.com/sirupsen/logrus"
)

type logCtxKeyType struct{}

var logCtxKey logCtxKeyType

var defaultLogEntry *log.Entry

func init() {
	logger := log.New()
	logger.SetFormatter(&log.JSONFormatter{})
	logger.SetOutput(os.Stdout)
	logger.SetLevel(log.DebugLevel)
	defaultLogEntry = log.NewEntry(logger)
}

func NewContext(ctx context.Context, fields log.Fields) context.Context {
	return context.WithValue(ctx, logCtxKey, WithContext(ctx).WithFields(fields))
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
