package framework

import (
	"context"

	"github.com/pingcap/tiem/library/firstparty/util"
	"github.com/pingcap/tiem/library/thirdparty/logger"
	log "github.com/sirupsen/logrus"
)

type Logger interface {
	GetDefaultLogger() *logger.LogRecord
	SetDefaultLogger(*logger.LogRecord)
	// Return the logger which has already stored inside the ctx.
	// If there is no such logger, return the default logger instead.
	WithContext(ctx context.Context) *logger.LogRecord
	// Return a new context with extra fields added to the previous logger.
	// If there is no such previous logger, use the default logger instead.
	NewContextWithField(ctx context.Context, key string, value interface{}) context.Context
	NewContextWithFields(ctx context.Context, fields log.Fields) context.Context
}

func initLogger() (Logger, error) {
	util.AssertWithInfo(false, "Not Implemented Yet")
	return nil, nil
}
